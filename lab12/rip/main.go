package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

type NetConfig struct {
	Routers []string   `json:"routers"`
	Edges   [][]string `json:"edges"`
}

type Router struct {
	IP              string
	Neighbors       []string
	Table           RouterTable
	TableChan       chan RouterMessage
	NeighborsTables map[string]RouterTable
}

type RouterPath struct {
	Dist int
	Next string
}
type RouterTable map[string]RouterPath
type RouterMessage struct {
	Src   string
	Table RouterTable
}

var (
	routers   = make(map[string]*Router)
	globMutex = sync.Mutex{}
	globCnt   = 0
	outp      = tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', tabwriter.TabIndent)
)

func NewRouter(ip string) *Router {
	routers[ip] = &Router{IP: ip, Neighbors: []string{}, Table: make(RouterTable), TableChan: make(chan RouterMessage, 100), NeighborsTables: make(map[string]RouterTable)}
	return routers[ip]
}
func NewEdge(e []string) {
	routers[e[0]].NewNeighbor(e[1])
	routers[e[1]].NewNeighbor(e[0])
}
func (r *Router) NewNeighbor(ip string) {
	r.Neighbors = append(r.Neighbors, ip)
}

func (r *Router) Run() {
	r.UpdateTable()
	iter := 0
	r.Print(iter)
	for {
		globMutex.Lock()
		if globCnt == 0 {
			globMutex.Unlock()
			break
		}
		globMutex.Unlock()
		select {
		case v := <-r.TableChan:
			r.NeighborsTables[v.Src] = v.Table
			r.UpdateTable()
			globMutex.Lock()
			globCnt--
			globMutex.Unlock()
			iter++
			r.Print(iter)
		default:
		}
	}
	r.Print(-1)
}

func (r *Router) UpdateTable() {
	t := make(RouterTable)
	t[r.IP] = RouterPath{0, r.IP}
	for _, n := range r.Neighbors {
		for nr, nt := range r.NeighborsTables[n] {
			if _, ok := t[nr]; !ok || (1+nt.Dist < t[nr].Dist) {
				t[nr] = RouterPath{1 + nt.Dist, n}
			}
		}
	}
	if !mapsEqual(t, r.Table) {
		globMutex.Lock()
		globCnt += len(r.Neighbors)
		globMutex.Unlock()

		r.Table = t
		for _, n := range r.Neighbors {
			Dispatch(r.IP, n, t)
		}
	}
}
func (r *Router) Print(step int) {
	globMutex.Lock()
	defer globMutex.Unlock()
	header := fmt.Sprintf("Simulation on router %s at step %d\n", r.IP, step)
	if step == -1 {
		header = fmt.Sprintf("Final state on router %s\n", r.IP)
	}
	outp.Write([]byte(header))
	outp.Write([]byte("Source\tDestination\tNext Hop\tDistance\n"))
	for t, p := range r.Table {
		outp.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%d\n", r.IP, t, p.Next, p.Dist)))
	}
	outp.Flush()
}

func Dispatch(source, target string, table RouterTable) {
	for _, r := range routers {
		if r.IP == target {
			r.TableChan <- RouterMessage{source, table}
			break
		}
	}
}

func main() {
	configFlag := flag.String("c", "config.json", "config file path")
	flag.Parse()
	cfgFile, err := os.Open(*configFlag)
	if err != nil {
		log.Fatalf("config file: %s\n", err.Error())
	}
	cfgData, err := io.ReadAll(cfgFile)
	cfgFile.Close()
	if err != nil {
		log.Fatalf("config file: %s\n", err.Error())
	}

	var cfg NetConfig
	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		log.Fatalf("config file: %s\n", err.Error())
	}

	for _, r := range cfg.Routers {
		NewRouter(r)
	}
	for _, e := range cfg.Edges {
		NewEdge(e)
	}
	wg := sync.WaitGroup{}
	for _, r := range routers {
		wg.Add(1)
		go func() {
			r.Run()
			wg.Done()
		}()
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
}

func mapsEqual(a, b RouterTable) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}
