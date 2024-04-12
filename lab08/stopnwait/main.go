package main

import (
	"flag"
	"log"
)

func main() {
	modeFlag := flag.String("mode", "server", "must be set server/client")
	serverPortFlag := flag.Int("server-port", 8080, "server-port")
	clientPortFlag := flag.Int("client-port", 8081, "client-port")
	timeoutFlag := flag.Int("timeout", 5, "timeout in seconds")
	sendFlag := flag.String("send", "./client-resources/example.txt", "path to file to send")
	recvFlag := flag.String("receive", "./server-resources/", "path to file received")
	flag.Parse()
	switch *modeFlag {
	case "server":
		runServer(*serverPortFlag, *timeoutFlag, *recvFlag)
	case "client":
		runClient(*serverPortFlag, *clientPortFlag, *timeoutFlag, *sendFlag)
	default:
		log.Fatalf("error parsing `mode` flag, got %s", *modeFlag)
	}
}
