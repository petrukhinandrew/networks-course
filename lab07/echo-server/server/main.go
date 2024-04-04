package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	modeFlag := flag.String("m", "echo", "must be `echo` for task A solution, `heartbeat` for task D solution")
	thresholdFlag := flag.Int("t", 3, "time (seconds) to assume client disconnect")
	flag.Parse()

	addr := net.UDPAddr{IP: net.ParseIP("localhost"), Port: 8080}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("error serving on %s - %s\n", addr.String(), err.Error())
	}
	defer conn.Close()
	log.Printf("Serving on %s", addr.String())

	switch *modeFlag {
	case "echo":
		runEcho(conn)
	case "heartbeat":
		runHeartbeat(conn, *thresholdFlag)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runEcho(conn *net.UDPConn) {
	buf := make([]byte, 2048)
	log.Println("running echo mode")
	for {
		log.Println("waiting package")
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("error reading from UDP - %s", err.Error())
			continue
		}
		log.Printf("got package %d bytes from %s", n, raddr.String())
		log.Println("package:")

		rawBytesString := ""
		for _, b := range buf[:n] {
			rawBytesString += fmt.Sprintf("%x ", b)
		}
		log.Println(rawBytesString)

		shouldLost := rand.Float32() < 0.2
		if shouldLost {
			log.Println("simulating 20%% loss")
			continue
		}

		resp := strings.ToUpper(string(buf[:n]))
		respN, err := conn.WriteTo([]byte(resp), raddr)
		if err != nil || respN != n {
			log.Printf("error writing %d bytes response - %s\n", respN, err.Error())
		}
		log.Printf("sent %d bytes response\n", respN)
	}

}

type ClientMetrics struct {
	lastPackIdx  int
	lastPackTime time.Time
	lossCnt      int
	initSentIdx  int
}

type ClientStore map[string]*ClientMetrics

type udpPair struct {
	addr *net.UDPAddr
	buf  []byte
}

func runHeartbeat(conn *net.UDPConn, threshold int) {
	clients := make(ClientStore)

	log.Println("running heartbeat mode")
	connChan := make(chan udpPair)

	go func(cn *net.UDPConn, pr chan udpPair) {
		buf := make([]byte, 2048)
		for {
			n, raddr, err := cn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("error reading from UDP - %s", err.Error())
				continue
			}
			pr <- udpPair{raddr, buf[:n]}
		}
	}(conn, connChan)
	thresholdDur := time.Second * time.Duration(threshold)
	cleanupTimer := time.NewTicker(thresholdDur)
	for {
		np := udpPair{}
		select {
		case <-cleanupTimer.C:
			clients.removeDisconnected(thresholdDur)
			continue
		case np = <-connChan:
		}
		log.Printf("got package bytes from %s", np.addr.String())

		shouldLost := rand.Float32() < 0.2
		if shouldLost {
			log.Println("simulating 20%% loss")
			continue
		}

		respPackIdx, err := clients.submitPacket(np.addr.String(), np.buf)

		resp := fmt.Sprintf("HB %d %s", respPackIdx, time.Now().String())

		if err != nil {
			log.Printf("error handling hb packet - %s\n", err.Error())
			resp = err.Error()
		}

		respN, err := conn.WriteTo([]byte(resp), np.addr)
		if err != nil {
			log.Printf("error writing %d bytes response - %s\n", respN, err.Error())
		}
		log.Printf("sent response to %s\n", np.addr.String())
		clients.removeDisconnected(time.Second * time.Duration(threshold))
	}
}

func (store ClientStore) submitPacket(addr string, pack []byte) (idx int, err error) {
	packet, err := bindHbPacket(pack)
	if err != nil {
		return 0, err
	}
	if m, ok := store[addr]; ok && m != nil {
		idx, err = store.updateHbClient(addr, packet)
	} else {
		idx, err = store.newHbClient(addr, packet)
	}
	return
}

type ClientPacket struct {
	idx  int
	time time.Time
}

var (
	ErrBadHbPacketFormat = errors.New("bad hb packet format")
	ErrBadHbPacketIdx    = errors.New("bad hb packet index")
	ErrBadHbPacketTime   = errors.New("bad hb packet time")
)

func bindHbPacket(pack []byte) (packet *ClientPacket, err error) {
	src := string(pack)
	tokens := strings.Split(src, " ")

	packet = &ClientPacket{}

	switch {
	case len(tokens) != 3:
		return nil, ErrBadHbPacketFormat
	case tokens[0] != "HB":
		return nil, ErrBadHbPacketFormat
	}

	if packet.idx, err = strconv.Atoi(tokens[1]); err != nil {
		return nil, ErrBadHbPacketIdx
	}

	if packet.time, err = time.Parse(time.RFC3339Nano, tokens[2]); err != nil {
		log.Fatalln(err)
		return nil, ErrBadHbPacketTime
	}
	return packet, nil
}

func (store ClientStore) newHbClient(addr string, packet *ClientPacket) (int, error) {

	store[addr] = &ClientMetrics{packet.idx, packet.time, 0, packet.idx}
	store.logClient(addr)
	return packet.idx, nil
}

func (store ClientStore) updateHbClient(addr string, packet *ClientPacket) (int, error) {
	loss := store[addr].lastPackIdx - packet.idx
	store[addr].lastPackIdx = packet.idx
	store[addr].lastPackTime = packet.time
	store[addr].lossCnt += loss
	store.logClient(addr)
	return packet.idx, nil
}

func (store ClientStore) removeDisconnected(threshold time.Duration) {
	for k, v := range store {
		if dt := time.Since(v.lastPackTime); dt > threshold {
			delete(store, k)
			log.Printf("removing client %s after %s", k, dt.String())
		}
	}
}

func (store ClientStore) logClient(addr string) {
	m := *store[addr]
	var lossP float32 = 0
	if m.lastPackIdx-m.initSentIdx > 0 {
		lossP = float32(m.lossCnt) / float32(m.lastPackIdx-m.initSentIdx) * 100
	}
	log.Printf("%s status: last packet %d at %s with %f%% loss", addr, m.lastPackIdx, m.lastPackTime.String(), lossP)
}
