package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	modeFlag := flag.String("mode", "ip", "use `ip` for task A solution, `ports` for task B solution")
	portsIPFlag := flag.String("pip", "127.0.0.1", "")
	portsBFlag := flag.Int("pb", 0, "")
	portsEFlag := flag.Int("pe", 65535, "")
	flag.Parse()
	switch *modeFlag {
	case "ip":
		ShowIPs()
	case "ports":
		ShowPorts(*portsIPFlag, *portsBFlag, *portsEFlag)
	default:
		log.Fatalf("mode not supported: %s", *modeFlag)
	}
}

func ShowIPs() {
	ints, err := net.Interfaces()
	if err != nil {
		log.Fatalf("error reading interfaces: %s\n", err.Error())
	}

	for _, i := range ints {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("error getting addr: %s\n", err.Error())
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if i.Flags&net.FlagLoopback == 0 {
					prefix := "v4 "
					if len(v.Mask) == 16 {
						prefix = "v6 "
					}
					o, _ := v.Mask.Size()
					fmt.Printf("%s %v/%d [%s]\n", prefix, v.IP, o, maskFormat(v.Mask))
				}
			}

		}
	}
}

func maskFormat(mask net.IPMask) string {
	if len(mask) == 4 {
		return v4MaskFormat(mask)
	}
	return v6MaskFormat(mask)
}

func v4MaskFormat(mask net.IPMask) string {
	buf := ""
	for i, b := range mask {
		if i != 0 {
			buf += "."
		}
		buf += fmt.Sprintf("%d", b)
	}
	return buf
}
func v6MaskFormat(mask net.IPMask) string {
	buf := ""

	for i := 0; i < 16; i += 2 {
		if i != 0 {
			buf += "."
		}
		buf += fmt.Sprintf("%.2x%.2x", mask[i], mask[i+1])
	}
	return buf
}

func ShowPorts(ip string, pb int, pe int) {
	threadLimit := 200
	maybePorts := make(chan int)
	freePorts := []int{}
	var lock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < threadLimit; i++ {
		wg.Add(1)
		go func() {
		work:
			for {
				select {
				case port, ok := <-maybePorts:
					if !ok {
						wg.Done()
						break work
					}
					log.Printf("scanning %s:%d", ip, port)
					conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*2)
					if err == nil && conn != nil {
						lock.Lock()
						freePorts = append(freePorts, port)
						lock.Unlock()
					}
				default:
				}
				time.Sleep(time.Millisecond * 20)
			}
		}()
	}
	for port := pb; port < pe; port++ {
		maybePorts <- port
	}
	close(maybePorts)
	wg.Wait()
	fmt.Printf("free ports: %d\n", freePorts)
}
