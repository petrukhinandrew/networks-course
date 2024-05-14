package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	portFlag := flag.Int("port", 8080, "port")
	flag.Parse()
	port := fmt.Sprintf("%d", *portFlag)
	addr, err := net.ResolveTCPAddr("tcp6", ":"+port)
	if err != nil {
		log.Fatalf("resolve: %s\n", err.Error())
	}
	listener, err := net.ListenTCP("tcp6", addr)
	if err != nil {
		log.Fatalf("listen: %s\n", err.Error())
	}
	defer listener.Close()

	log.Println("listening on: port")
	for {
		func() {
			c, err := listener.Accept()
			if err != nil {
				log.Println("accept: " + err.Error())
				return
			}
			defer c.Close()
			log.Println("accept from " + c.RemoteAddr().String())
			var buf []byte
			buf, err = io.ReadAll(c)

			if err != nil {
				log.Println("read: " + err.Error())
				return
			}
			log.Printf("read: %s\n", string(buf))

			_, err = c.Write([]byte(strings.ToUpper(string(buf))))
			if err != nil {
				log.Println("write: " + err.Error())
				return
			}

		}()
	}

}
