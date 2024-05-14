package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"text/tabwriter"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	iterFlag := flag.Int("n", 3, "packets per ttl")
	ttlFlag := flag.Int("l", 20, "ttl limit")
	dstFlag := flag.String("dst", "www.yandex.ru", "destination ip")
	flag.Parse()
	dst, err := net.ResolveIPAddr("ip4:icmp", *dstFlag)
	if err != nil {
		log.Fatalf("target resolve: %s\n", err.Error())
	}
	sock, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("dial resolve: %s\n", err.Error())
	}
	defer sock.Close()

	conn := ipv4.NewPacketConn(sock)
	defer conn.Close()

	if err := conn.SetControlMessage(ipv4.FlagTTL|ipv4.FlagDst|ipv4.FlagInterface|ipv4.FlagSrc, true); err != nil {
		log.Fatalf("control msg: %s\n", err)
	}
	conn.SetReadDeadline(time.Now().Add(time.Second))

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff,
			Data: []byte("leave me be")},
	}
	respBuf := make([]byte, 1500)
	outp := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
loop:
	for ttl := 1; ttl <= *ttlFlag; ttl++ {
		var tbuf []time.Duration
		buf, err := msg.Marshal(nil)
		if err != nil {
			log.Fatalf("encoding: %s\n", err.Error())
		}
		if err = conn.SetTTL(ttl); err != nil {
			log.Fatalf("ttl upd: %s\n", err.Error())
		}
		for iter := 0; iter < *iterFlag; iter++ {
			msg.Body.(*icmp.Echo).Seq = 1

			st := time.Now()
			if _, err := conn.WriteTo(buf, nil, dst); err != nil {
				log.Fatalf("write: %s\n", err.Error())
			}

			respLen, _, node, err := conn.ReadFrom(respBuf)
			if err != nil {
				// log.Printf("read: %s\n", err.Error())
				continue
			}
			tbuf = append(tbuf, time.Since(st))
			ans, err := icmp.ParseMessage(1, respBuf[:respLen])
			if err != nil {
				log.Printf("parse: %s\n", err.Error())
				continue
			}
			if iter == *iterFlag-1 {
				hostname := node.String()
				opts, _ := net.LookupAddr(node.String())
				if len(opts) > 0 {
					hostname = opts[0]
				}
				rttRes := rtt(tbuf)
				fmt.Fprintf(outp, "%d\t%s\t( %s )\t%v\t%v\t%v\n", ttl, hostname, node.String(), rttRes[0], rttRes[1], rttRes[2])
				if ans.Type == ipv4.ICMPTypeEchoReply {
					break loop
				}
			}
		}
	}
	outp.Flush()
}

func rtt(data []time.Duration) []time.Duration {
	minT := time.Hour
	maxT := time.Second * 0
	avg := time.Second * 0
	for _, t := range data {
		minT = min(minT, t)
		maxT = max(maxT, t)
		avg += t
	}
	return []time.Duration{minT, maxT, avg / time.Duration(len(data))}
}
