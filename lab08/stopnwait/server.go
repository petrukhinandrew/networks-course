package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type ReliableServerClient struct {
	Addr         net.Addr
	LastRecvTime time.Time
	LastRecvMsg  Packet
	LastRespTime time.Time
	LastResp     Packet
	Buf          []byte
	ExpSize      int
}
type ServerClients struct {
	ResendPollStop chan struct{}
	Clients        []ReliableServerClient
	SavePath       string
}

func runServer(port int, timeout int, savePath string) {
	udpServer, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("error starting serever: %s", err.Error())
	}
	defer udpServer.Close()

	shouldStop := make(chan struct{}, 1)
	sc := &ServerClients{ResendPollStop: shouldStop, SavePath: savePath}

	log.Printf("starting server on %s", udpServer.LocalAddr().String())
	go sc.resendOnTimeout(udpServer, timeout)

	for {
		buf := make([]byte, 1024)
		n, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			log.Printf("error reading package: %s", err.Error())
			continue
		}
		p, err := DecodePacket(buf[:n])
		if err != nil {
			log.Printf("error decoding packet: %s", err.Error())
			continue
		}
		go sc.serverRespond(udpServer, addr, *p)
	}

}

func (sc *ServerClients) resendOnTimeout(server net.PacketConn, timeout int) {
	shouldResend := time.NewTicker(time.Second * time.Duration(timeout/2))
	log.Println("starting polling losses")
	for {
		select {
		case <-sc.ResendPollStop:
			return
		case <-shouldResend.C:
			for _, c := range sc.Clients {
				if time.Since(c.LastRespTime) > time.Duration(timeout) {
					_, err := server.WriteTo(c.LastResp.Bytes(), c.Addr)
					if err != nil {
						log.Printf("error on resend: %s", err.Error())
						continue
					}
					c.LastRespTime = time.Now()
				}
			}
		}
	}
}

func (sc *ServerClients) lookup(addr string) int {
	for idx, c := range sc.Clients {
		if c.Addr.String() == addr {
			return idx
		}
	}
	return -1
}

func (sc *ServerClients) serverRespond(server net.PacketConn, addr net.Addr, p Packet) {
	connIdx := sc.lookup(addr.String())
	if connIdx == -1 {
		sc.Clients = append(sc.Clients, ReliableServerClient{Addr: addr})
		connIdx = len(sc.Clients) - 1
	}
	sc.Clients[connIdx].LastRecvTime = time.Now()
	sc.Clients[connIdx].LastRecvMsg = p
	log.Printf("got msg %s from %s", string(p.data), addr.String())

	if err := sc.handlePacket(connIdx); err != nil {
		log.Printf("error handling packet: %s", err.Error())
		return
	}

	resp := *AckPacket(int(p.index))
	_, err := server.WriteTo(resp.Bytes(), addr)
	if err != nil {
		log.Printf("error writing reponse to %s: %s", addr.String(), err.Error())
		return
	}
	sc.Clients[connIdx].LastRespTime = time.Now()
	sc.Clients[connIdx].LastResp = resp
}
func isMeta(opt []byte) bool {
	return opt[0] == byte('M') && opt[1] == byte('E') && opt[2] == byte('T') && opt[3] == byte('A')
}
func (sc *ServerClients) handlePacket(idx int) (err error) {
	lp := sc.Clients[idx].LastRecvMsg
	if isMeta(lp.data[:4]) {
		err = sc.handleMeta(idx)
	} else {
		err = sc.handleChunk(idx)
	}
	return
}

func (sc *ServerClients) handleMeta(idx int) error {
	lp := sc.Clients[idx].LastRecvMsg
	if len(lp.data) != 12 {
		return errors.New("bad meta body size")
	}
	bSize := binary.LittleEndian.Uint64(lp.data[4:])
	sc.Clients[idx].Buf = []byte{}
	sc.Clients[idx].ExpSize = int(bSize)
	return nil
}

func (sc *ServerClients) handleChunk(idx int) error {
	sc.Clients[idx].Buf = append(sc.Clients[idx].Buf, sc.Clients[idx].LastRecvMsg.data...)
	println(len(sc.Clients[idx].Buf))
	if len(sc.Clients[idx].Buf) == sc.Clients[idx].ExpSize {
		sc.storeFile(idx)
	}
	return nil
}

func (sc *ServerClients) storeFile(idx int) {
	of, err := os.Create(sc.SavePath + sc.Clients[idx].Addr.String())
	if err != nil {
		log.Fatalf("error opening save file: %s", err.Error())
	}
	defer of.Close()

	if _, err = of.Write(sc.Clients[idx].Buf); err != nil {
		log.Fatalf("error writing save file: %s", err.Error())
	}
	log.Printf("created save file for %s", sc.Clients[idx].Addr.String())
}
