package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

func main() {
	isServer := flag.Bool("s", false, "is server")
	isClient := flag.Bool("c", false, "is client")
	cc := flag.String("cmd", "ping -c 3 yandex.ru", "client command that should be run on server")
	flag.Parse()
	if *isServer == *isClient {
		log.Fatalln("cannot work in both modes")
	}

	switch {
	case *isServer:
		runServer()
	case *isClient:
		runClient(*cc)
	}
}

func runServer() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error listening %s\n", err.Error())
		return
	}
	defer server.Close()
	log.Println("server started")

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("error accepting %s\n", err.Error())
			continue
		}
		log.Printf("accepted conn %s\n", conn.RemoteAddr())
		go serverConn(conn)
		time.Sleep(time.Millisecond * 20)
	}
}

func serverConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error reading cmd %s\n", err.Error())
		return
	}
	log.Printf("received %s\n", string(buf[:n]))
	tokens := strings.Split(string(buf[:n]), " ")
	target, args := tokens[0], tokens[1:]

	output, err := exec.Command(target, args...).Output()
	if err != nil {
		log.Printf("error executing %s\n", err.Error())
		return
	}
	log.Println("execution completed")
	_, err = conn.Write(output)
	if err != nil {
		log.Printf("error writing result %s\n", err.Error())
	}
	log.Println("result sent")
}

func runClient(cmd string) {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatalf("error connecting %s\n", err.Error())
	}
	defer conn.Close()
	log.Println("client started")
	conn.Write([]byte(cmd))
	log.Printf("requesting %s\n", cmd)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatalf("error reading response %s\n", err.Error())
	}
	log.Printf("response %s", string(buf[:n]))
}
