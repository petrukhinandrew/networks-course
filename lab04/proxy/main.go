package main

import (
	"bytes"
	_ "embed"
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

type ProxyCacheEntry struct {
	body    []byte
	lastMod string
}

type ProxyConfig struct {
	addr      string
	cache     map[string]*ProxyCacheEntry
	blacklist []string
}

func NewProxyConfig() *ProxyConfig {
	return &ProxyConfig{addr: "", cache: make(map[string]*ProxyCacheEntry)}
}

var myLog Logger
var config = NewProxyConfig()

//go:embed blacklist.txt
var blackpool string

func setupBlacklist() {
	blacklist := strings.Split(blackpool, "\n")
	for _, e := range blacklist {
		config.blacklist = append(config.blacklist, dropAnchor(dropHTTP(e)))
	}
}
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
	setupBlacklist()
	myLog.Debug.Printf("Starting proxy on %s", config.addr)

	mux := slashfix.NewSkipSlashMux()

	mux.HandleFunc("/", proxyHandler)

	if err = http.ListenAndServe(config.addr, mux); err != nil {
		panic("Error serving " + err.Error())
	}
}
func interceptReferer(req *http.Request) (string, error) {
	res := req.RequestURI
	referer := req.Header.Get("Referer")
	if referer == "" {
		return res, nil
	}
	myLog.Debug.Printf("Handling referer: %s\n", referer)
	referer = strings.TrimPrefix(referer, "http://")
	referer = strings.TrimPrefix(referer, "https://")
	if target := strings.TrimPrefix(referer, config.addr); target != referer {
		myLog.Debug.Printf("Composed URI: %s\n", target+res)
		return target + res, nil
	}
	return "", errors.New("bad referrer " + referer)
}

func newProxyCacheEntry(url string, resp *http.Response, body *bytes.Buffer) (err error) {
	newEntry := &ProxyCacheEntry{}
	newEntry.body, err = io.ReadAll(body)
	if err != nil {
		return
	}
	newEntry.lastMod = resp.Header["Date"][0]
	config.cache[url] = newEntry
	return nil
}

func checkProxyCache(url string) (*ProxyCacheEntry, bool) {
	if v, ok := config.cache[url]; ok {
		myLog.Debug.Printf("Found cache entry for %s\n", url)
		return v, ok
	}
	myLog.Debug.Printf("No cache entry for %s\n", url)
	return nil, false
}

func checkNotModified(url string, cached *ProxyCacheEntry) (bool, error) {
	tmpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		myLog.Debug.Printf("Error making request for %s\n", url)
		return false, err
	}
	req.Header["If-Modified-Since"] = []string{cached.lastMod}
	// req.Header["If-None-Match"] = []string{cached.etag}
	resp, err := tmpClient.Do(req)
	if err != nil {
		return false, err
	}
	myLog.Debug.Printf("Got mod response for %s got %s\n", url, resp.Status)
	return resp.StatusCode == http.StatusNotModified, nil
}

func blacklistLookup(url string) bool {
	target := dropAnchor(dropHTTP(url))
	for _, e := range config.blacklist {
		if e == target {
			return true
		}
	}
	return false
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
	if blacklistLookup(targetURL) {
		w.WriteHeader(http.StatusForbidden)
		myLog.Reqs.Printf("URL: %s, Status: %d, Method: %s found in black list\n", targetURL, http.StatusForbidden, http.MethodGet)
		myLog.Debug.Printf("Black list entry found %s\n", targetURL)
		return
	}
	if entry, cached := checkProxyCache(targetURL); cached {
		if notMod, err := checkNotModified(targetURL, entry); err == nil && notMod {
			w.WriteHeader(http.StatusOK)
			var n int
			if n, err = w.Write(entry.body); err != nil {
				myLog.Debug.Printf("Error writing body - %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			myLog.Reqs.Printf("URL: %s, Status: %d, Method: %s from cache %d bytes\n", targetURL, http.StatusOK, http.MethodGet, n)
			myLog.Debug.Printf("Written body from cache %d bytes for %s with response %d\n", n, targetURL, http.StatusOK)
			return
		}
	}

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
	var targetBodyBuf bytes.Buffer
	teeReader := io.TeeReader(targetResp.Body, &targetBodyBuf)
	targetBody, err := io.ReadAll(teeReader)
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
	if err = newProxyCacheEntry(targetURL, targetResp, &targetBodyBuf); err != nil {
		myLog.Debug.Printf("Error updating cache %s\n", err.Error())
	}
	myLog.Reqs.Printf("URL: %s, Status: %s, Method: %s\n", targetResp.Request.URL.String(), targetResp.Status, targetReq.Method)
}

func dropHTTP(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
}
func dropAnchor(url string) string {
	return strings.Split(url, "#")[0]
}
