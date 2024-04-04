package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	modeFlag := flag.String("m", "ping", "must be `ping` for task A solution, `heartbeat` for task D solution")
	hbcFlag := flag.Int("c", 1, "number of heartbeat clients")
	flag.Parse()
	switch *modeFlag {
	case "ping":
		runPing()
	case "heartbeat":
		runHeartbeat(*hbcFlag)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runPing() {
	log.Printf("running client in ping mode")
	addr := "localhost:8080"
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatalf("error dialing %s", err.Error())
	}
	defer conn.Close()

	log.Printf("local addr %s", conn.LocalAddr().String())
	log.Printf("connected to %s", conn.RemoteAddr().String())

	lossCnt := 0
	rtts := make([]time.Duration, 10)
	minT := time.Second * 2
	maxT := time.Second * 0
	avgT := time.Second * 0

	for iter := range 10 {
		msg := fmt.Sprintf("Ping %d %s", iter+1, time.Now().String())
		wtime := time.Now()
		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("error writing msg %s", err.Error())
			continue
		}
		log.Printf("sent %d bytes with %s", n, msg)
		buf := make([]byte, 2048)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, err = conn.Read(buf)
		restime := time.Since(wtime)
		if err != nil {
			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout() {
				log.Println("request timed out")
				lossCnt += 1
			} else {
				log.Printf("error reading response %s", err.Error())
			}
		} else {
			log.Printf("got %d bytes response: %s time: %s", n, string(buf[:n]), restime.String())
			if restime < minT {
				minT = restime
			}
			if restime > maxT {
				maxT = restime
			}
			avgT += restime
			rtts = append(rtts, restime)
		}
	}
	avgT = avgT / time.Duration(10-lossCnt)

	log.Printf("rtt min/avg/max/mdev = %s/%s/%s/%s", minT.String(), avgT.String(), maxT.String(), mdev(avgT, rtts))
	log.Printf("Package loss = %d%%", 10*lossCnt)
}

func mdev(avg time.Duration, src []time.Duration) time.Duration {
	acc := time.Duration(0)
	for _, t := range src {
		acc += (avg - t) * (avg - t)
	}
	ms := math.Sqrt(float64((acc / time.Duration(len(src))).Milliseconds()))
	return time.Microsecond * time.Duration(ms)
}

func runHeartbeat(clientCnt int) {
	log.Printf("running client in heartbeat mode with %d clients", clientCnt)
	wg := sync.WaitGroup{}
	for i := range clientCnt {
		go func(idx int) {
			log.Printf("starting client %d", idx)
			wg.Add(1)
			runHbClient(idx)
			wg.Done()
		}(i)
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait()
}

func runHbClient(idx int) {
	addr := "localhost:8080"
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatalf("client %d, error dialing %s", idx, err.Error())
	}
	defer conn.Close()

	log.Printf("client %d on %s", idx, conn.LocalAddr().String())

	lossCnt := 0
	rtts := make([]time.Duration, 10)
	minT := time.Second * 2
	maxT := time.Second * 0
	avgT := time.Second * 0

	for iter := range 10 {
		msg := fmt.Sprintf("HB %d %s", iter+1, time.Now().Format(time.RFC3339Nano))
		wtime := time.Now()
		_, err := conn.Write([]byte(msg))

		if err != nil {
			log.Printf("client %d error writing msg %s", idx, err.Error())
			continue
		}
		log.Printf("client %d sent %d HB", idx, iter)
		buf := make([]byte, 2048)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		_, err = conn.Read(buf)
		restime := time.Since(wtime)
		if err != nil {
			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout() {
				log.Printf("client %d request timed out", idx)
				lossCnt += 1
			} else {
				log.Printf("client %d error reading response %s", idx, err.Error())
			}
		} else {
			log.Printf("client %d got response. time: %s", idx, restime.String())
			if restime < minT {
				minT = restime
			}
			if restime > maxT {
				maxT = restime
			}
			avgT += restime
			rtts = append(rtts, restime)
		}
	}
	avgT = avgT / time.Duration(10-lossCnt)

	log.Printf("client %d: rtt min/avg/max/mdev = %s/%s/%s/%s", idx, minT.String(), avgT.String(), maxT.String(), mdev(avgT, rtts))
	log.Printf("client %d: package loss = %d%%", idx, 10*lossCnt)
}
