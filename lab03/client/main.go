package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	hostFlag := flag.String("h", "localhost", "server_host")
	portFlag := flag.Int("p", 8080, "server_port")
	port := fmt.Sprintf("%d", *portFlag)
	fnameFlag := flag.String("f", "Quickstart.md", "filename")
	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	addr := *hostFlag + ":" + port
	run(addr, *fnameFlag)
}

func run(addr string, fname string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Dial error - %s\n", err.Error())
	}
	defer conn.Close()
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", fname, addr)

	if _, err = conn.Write([]byte(req)); err != nil {
		log.Fatalf("Write error - %s\n", err.Error())
	}

	buf := make([]byte, 1024)

	if _, err = conn.Read(buf); err != nil {
		log.Fatalf("Read error - %s\n", err.Error())
	}

	log.Println(string(buf))
}
