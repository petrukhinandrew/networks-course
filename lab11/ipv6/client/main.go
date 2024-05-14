package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	portFlag := flag.Int("port", 8080, "server port")
	msgFlag := flag.String("msg", "echo test msg", "msg to send")
	flag.Parse()
	addr, err := net.ResolveTCPAddr("tcp6", ":"+fmt.Sprintf("%d", *portFlag))
	if err != nil {
		log.Fatalf("resolve: %s\n", err.Error())
	}

	d, err := net.DialTCP("tcp6", nil, addr)
	if err != nil {
		log.Fatalf("dial: %s\n", err.Error())
	}
	defer d.Close()

	log.Printf("connected to %s\n", d.RemoteAddr().String())

	_, err = d.Write([]byte(*msgFlag))
	if err != nil {
		log.Fatalf("write: %s\n", err.Error())
	}
	log.Printf("write: %s\n", string(*msgFlag))
	d.CloseWrite()
	resp, err := io.ReadAll(d)
	if err != nil {
		log.Fatalf("read: %s\n", err.Error())
	}

	log.Println(string(resp))
}
