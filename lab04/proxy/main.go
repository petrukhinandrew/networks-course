package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"proxy/slashfix"
	"strings"
	"time"
)

type Logger struct {
	Reqs  log.Logger
	Debug log.Logger
}
type ProxyCache struct{}
type ProxyConfig struct {
	addr  string
	cache ProxyCache
}

func NewProxyConfig() *ProxyConfig {
	return &ProxyConfig{}
}

var myLog Logger
var config = NewProxyConfig()

func main() {
	portFlag := flag.Int("p", 8080, "port")
	hostFlag := flag.String("h", "localhost", "host")
	flag.Parse()

	logFname := "req-logs/" + time.Now().Format("2006.01.02-15:04:05") + ".log"
	logF, err := os.OpenFile(logFname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Error opening log file" + err.Error())
	}
	defer logF.Close()
	myLog.Reqs.SetOutput(logF)
	myLog.Debug.SetOutput(os.Stdout)

	config.addr = fmt.Sprintf("%s:%d", *hostFlag, *portFlag)

	myLog.Debug.Printf("Starting proxy on %s", config.addr)

	mux := slashfix.NewSkipSlashMux()

	mux.HandleFunc("/", proxyHandler)

	if err = http.ListenAndServe(config.addr, mux); err != nil {
		panic("Error serving " + err.Error())
	}
}
func interceptReferer(req *http.Request) (string, error) {
	res := req.RequestURI
	for k, v := range req.Header {
		if k == "Referer" {
			if len(v) < 1 {
				return "", errors.New("bad referer header")
			}
			src := v[0]
			myLog.Debug.Printf("Handling referer: %s\n", src)
			src = strings.TrimPrefix(src, "http://")
			src = strings.TrimPrefix(src, "https://")
			if target := strings.TrimPrefix(src, config.addr); target != src {
				myLog.Debug.Printf("Composed URI: %s\n", target+res)
				return target + res, nil
			}
		}
	}
	return res, nil
}

func proxyHandler(w http.ResponseWriter, req *http.Request) {
	myLog.Debug.Printf("Got request to %s\n", req.RequestURI)
	targetURL, err := interceptReferer(req)

	if err != nil {
		myLog.Debug.Printf("Error handling headers - %s\n", req.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		myLog.Debug.Printf("Got request with unsupported method %s\n", req.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	targetURL = strings.Trim(targetURL, "/")
	targetURL = strings.Replace(targetURL, "https:/", "http:/", 1)
	targetURL = strings.Replace(targetURL, "http:/", "http://", 1)
	targetURL = strings.Replace(targetURL, "http:///", "http://", 1)

	myLog.Debug.Printf("Target URL is %s\n", targetURL)

	targetReq, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		myLog.Debug.Printf("Error making target request - %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tmpClient := http.Client{}
	targetResp, err := tmpClient.Do(targetReq)

	if err != nil {
		myLog.Debug.Printf("Error sending target request - %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer targetResp.Body.Close()

	targetBody, err := io.ReadAll(targetResp.Body)
	if err != nil {
		myLog.Debug.Printf("Error reading target body - %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for k, v := range targetResp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(targetResp.StatusCode)
	if _, err = w.Write(targetBody); err != nil {
		myLog.Debug.Printf("Error writing body - %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	myLog.Debug.Printf("Written response %d\n", targetResp.StatusCode)
	myLog.Reqs.Printf("URL: %s, Status: %s, Method: %s\n", targetResp.Request.URL.String(), targetResp.Status, targetReq.Method)
}
