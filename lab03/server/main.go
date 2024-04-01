package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	modeFlag := flag.String("m", "simple", "server mode")
	portFlag := flag.Int("p", 8080, "server port")
	boundFlag := flag.Int("b", -1, "concurrency level")

	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *modeFlag == "bound" && boundFlag == nil {
		boundFlag = new(int)
		*boundFlag = -1
	}
	if *boundFlag < -1 {
		log.Fatalf("Concurrency level has to be >= -1, got %d", *boundFlag)
	}
	host := "localhost"
	port := fmt.Sprintf("%d", *portFlag)
	addr := host + ":" + port
	switch *modeFlag {
	case "simple":
		runSimpleServer(addr)
	case "par":
		runParServer(addr)
	case "bound":
		if *boundFlag == -1 {
			runParServer(addr)
		} else {
			runBoundServer(addr, *boundFlag)
		}
	default:
		panic("Mode not supported: " + *modeFlag)
	}
}

// Task A solution
func runSimpleServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Error listening %s - %s", addr, err.Error())
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting - %s", err.Error())
		}
		handleConn(conn)
	}
}

// Task B solution
func runParServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Error listening %s - %s", addr, err.Error())
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting - %s", err.Error())
		}
		go handleConn(conn)
	}
}

// Task D solution
func runBoundServer(addr string, limit int) {
	runners := make(chan struct{}, limit)
	reqs := make(chan net.Conn)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Error listening %s - %s", addr, err.Error())
	}
	defer listener.Close()
	go handleBound(runners, reqs)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting - %s", err.Error())
		}
		reqs <- conn
	}
}

const (
	BUF_CAP      = 1024
	RESOURCE_DIR = "resources/"
)

var (
	NOT_FOUND_HDR  = []byte("HTTP/1.1 404 Not Found" + crlf() + crlf())
	OK_HDR         = []byte("HTTP/1.1 200 OK" + crlf())
	CONTENT_TYPE   = []byte("Content-Type: text/plain; charset=UTF-8" + crlf())
	CONTENT_LENGTH = []byte("Content-Length: ")
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, BUF_CAP)
	n, err := conn.Read(buf)

	if err != nil && err != io.EOF {
		log.Printf("Error reading connection - %s", err.Error())
		return
	}
	buf = buf[:n]
	req := strings.Split(string(buf), " ")
	if len(req) < 2 {
		log.Printf("Error parsing request - %v", req)
		conn.Write(NOT_FOUND_HDR)
		return
	}
	fname := strings.Trim(req[1], "/")
	fbytes, err := os.ReadFile(RESOURCE_DIR + fname)
	if err != nil {
		log.Printf("Error reading file - %s\n", err.Error())
		conn.Write(NOT_FOUND_HDR)
		return
	}
	rawResponse := fmt.Sprintf("%s%s%d%s%s%s%s%s", OK_HDR, CONTENT_LENGTH, len(fbytes), crlf(), CONTENT_TYPE, crlf(), crlf(), fbytes)
	n, err = conn.Write([]byte(rawResponse))
	if err != nil {
		log.Printf("Error writing response: %s\n", err.Error())
		return
	}
	log.Printf("Written response: %d bytes", n)
}

func handleBound(runners chan struct{}, reqs chan net.Conn) {
	for req := range reqs {
		runners <- struct{}{}
		go func(retChan chan struct{}) {
			handleConn(req)
			<-retChan
		}(runners)
	}
}
func crlf() string {
	var osSep = fmt.Sprintf("%v", os.PathSeparator)
	var lb = "\n"
	if osSep != "/" {
		lb = "\r\n"
	}
	return lb
}
