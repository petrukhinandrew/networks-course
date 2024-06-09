package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/secsy/goftp"
)

func main() {
	client, err := goftp.DialConfig(
		goftp.Config{
			User:     "andrew",
			Password: "***",
			Timeout:  10 * time.Second,
		},
		"127.0.0.1:21")
	if err != nil {
		log.Fatalf("err dial %s\n", err.Error())
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
loop:
	for {
		buf, err := reader.ReadString('\n')
		buf = strings.Trim(buf, "\r\n ")
		if err != nil {
			log.Println(err.Error())
			continue
		}
		tokens := strings.Split(buf, " ")
		log.Println(tokens)
		switch tokens[0] {
		case "h":
		case "?":
		case "help":
			fmt.Printf("help : current help text\n ls   : list files tree\nload : [ftp src] [local dst]\nstore: [local src] [ftp dest]\nquit : quit client\n")
		case "ls":
			if len(tokens) > 1 {
				log.Printf("ignoring %v on ls\n", tokens[1:])
			}
			ls(client, "./", 0)
		case "store":
			if len(tokens) != 3 {
				log.Printf("bad store format %s, should be store local_src ftp_dst\n", buf)
			}
			store(client)
		case "load":
			if len(tokens) != 3 {
				log.Printf("bad load format %s, should be load ftp_src local_dst\n", buf)
			}
			load(client)
		case "quit":
			if len(tokens) > 1 {
				log.Printf("ignoring %v on quit\n", tokens[1:])
			}
			break loop
		default:
			fmt.Printf("unsupported %s\n", buf)
		}
	}
}

func ls(client *goftp.Client, cur string, depth int) {
	if depth == 2 {
		return
	}
	inf, err := client.ReadDir(cur)
	if err != nil {
		log.Printf("error ls %s\n", err.Error())
	}
	dx := ""
	if depth > 0 {
		dx = strings.Repeat("   ", depth-1) + " |-"
	}
	for _, e := range inf {
		fmt.Printf("%s%s\n", dx, e.Name())
		if e.IsDir() {
			ls(client, cur+e.Name()+"/", depth+1)
		}
	}

}
func store(client *goftp.Client) {

}
func load(client *goftp.Client) {

}
