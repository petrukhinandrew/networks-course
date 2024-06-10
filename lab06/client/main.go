package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
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
		switch tokens[0] {
		case "h":
			fallthrough
		case "?":
			fallthrough
		case "help":
			fmt.Printf("help : current help text\n ls   : list files tree [opt -d n for depth limit]\nload : [ftp src] [local dst]\nstore: [local src] [ftp dest]\nquit : quit client\n")
		case "ls":
			dl := 3
			if len(tokens) == 3 && tokens[1] == "-d" {
				if n, err := strconv.Atoi(tokens[2]); err != nil {
					log.Printf("error parsing depth limit, defaults to 3")
				} else {
					dl = n
					log.Printf("depth limit set to %d\n", dl)
				}
			} else if len(tokens) > 1 {
				log.Printf("ignoring %v on ls\n", tokens[1:])
			}
			ls(client, "./", 0, dl)
		case "store":
			if len(tokens) != 3 {
				log.Printf("bad store format %s, should be store local_src ftp_dst\n", buf)
			}
			store(client, tokens[1], tokens[2])
		case "load":
			if len(tokens) != 3 {
				log.Printf("bad load format %s, should be load ftp_src local_dst\n", buf)
			}
			load(client, tokens[1], tokens[2])
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

func ls(client *goftp.Client, cur string, depth, depthLimit int) {
	if depth == depthLimit {
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
			ls(client, cur+e.Name()+"/", depth+1, depthLimit)
		}
	}

}

func store(client *goftp.Client, src, dst string) {
	f, err := os.Open(src)
	if err != nil {
		log.Printf("error on local file open %s\n", err.Error())
		return
	}
	defer f.Close()

	if err = client.Store(dst, f); err != nil {
		log.Printf("error on store %s\n", err.Error())
		return
	}
	log.Printf("store complete from %s to %s\n", src, dst)
}

func load(client *goftp.Client, src, dst string) {
	f, err := os.Create(dst)
	if err != nil {
		log.Printf("error on local file open %s\n", err.Error())
		return
	}
	defer f.Close()
	if err = client.Retrieve(src, f); err != nil {
		log.Printf("error on load %s\n", err.Error())
		return
	}
	log.Printf("load complete from %s to %s\n", src, dst)
}
