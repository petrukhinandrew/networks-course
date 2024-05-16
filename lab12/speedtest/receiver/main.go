package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"unicode"

	"github.com/rivo/tview"
)

var (
	app  *tview.Application
	form *tview.Form
)

var recvIP string = "127.0.0.1"
var recvPort string = "8080"
var protocol = "TCP"

func showErrModal(text string) {
	m := tview.NewModal().SetText(text).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) { app.SetRoot(form, false) })
	app.SetRoot(m, false)
}

func main() {
	form = tview.NewForm().
		AddInputField("Receiver IP",
			"127.0.0.1",
			20,
			func(textToCheck string, lastChar rune) bool {
				return len(textToCheck) < 16 && (lastChar == rune('.') || unicode.IsDigit(lastChar))
			},
			func(text string) {
				recvIP = text
			}).
		AddInputField("Receiver port",
			"8080",
			10,
			func(textToCheck string, lastChar rune) bool {
				c, err := strconv.Atoi(textToCheck)
				if err != nil {
					return false
				}
				return c >= 0 && c < 65535
			},
			func(text string) {
				recvPort = text
			}).
		AddDropDown(
			"Protocol",
			[]string{"TCP", "UDP"},
			0,
			func(option string, optionIndex int) {
				protocol = option
			}).
		AddButton("Start receiving", func() {
			optIP := net.ParseIP(recvIP)
			if optIP == nil {
				showErrModal("Bad receiver IP")
				return
			}
			check()
		})
	form.SetBorder(true).SetTitle("Receiver")
	app = tview.NewApplication().SetRoot(form, true).EnableMouse(true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func check() {
	time.Sleep(time.Millisecond * 50)
	switch protocol {
	case "TCP":
		checkTCP()
	case "UDP":
		checkUDP()
	default:
	}
}

func checkTCP() {
	saddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(recvIP, recvPort))
	if err != nil {
		showErrModal(err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", saddr)
	if err != nil {
		showErrModal(err.Error())
		return
	}

	defer ln.Close()
	c, err := ln.AcceptTCP()
	if err != nil {
		showErrModal(err.Error())
		return
	}
	defer c.Close()

	buf, err := io.ReadAll(c)
	if err != nil {
		showErrModal(err.Error())
		return
	}
	for i := range len(buf) / 17 {
		p, err := Unmarshal(buf[i*17 : (i+1)*17])
		if err != nil {
			showErrModal(err.Error())
			return
		}
		addr := c.RemoteAddr()
		if _, ok := store[addr.String()]; !ok {
			store[addr.String()] = &PacketBucket{Last: p.Id, Max: p.MaxId, Durs: []time.Duration{p.S}}
		} else {
			store[addr.String()].Last = p.Id
			store[addr.String()].Durs = append(store[addr.String()].Durs, p.S)
		}

	}
	avg := time.Duration(0)

	for _, d := range store[c.RemoteAddr().String()].Durs {
		avg += d
	}
	avg /= time.Duration(len(store[c.RemoteAddr().String()].Durs))
	showErrModal(fmt.Sprintf("Source: %s TCP\nAverage time: %v\nReceived: %d/%d", c.RemoteAddr().String(), avg, len(store[c.RemoteAddr().String()].Durs), store[c.RemoteAddr().String()].Max))
}
func checkUDP() {
	_, err := net.ResolveUDPAddr("udp", net.JoinHostPort(recvIP, recvPort))
	if err != nil {
		showErrModal(err.Error())
		return
	}
	c, err := net.ListenPacket("udp", net.JoinHostPort(recvIP, recvPort))
	if err != nil {
		showErrModal(err.Error())
		return
	}
	defer c.Close()
	for i := range 100 {
		tmp := make([]byte, 1024)
		n, addr, err := c.ReadFrom(tmp)
		tmp = tmp[:17]
		if err != nil || n != 17 {
			showErrModal(err.Error())
			return
		}
		p, err := Unmarshal(tmp)
		if err != nil {
			showErrModal(err.Error())
			return
		}
		if _, ok := store[addr.String()]; !ok {
			store[addr.String()] = &PacketBucket{Last: p.Id, Max: p.MaxId, Durs: []time.Duration{p.S}}
		} else {
			store[addr.String()].Last = p.Id
			store[addr.String()].Durs = append(store[addr.String()].Durs, p.S)
		}
		if p.Id == p.MaxId-1 || int8(i) > p.MaxId {
			avg := time.Duration(0)

			for _, d := range store[addr.String()].Durs {
				avg += d
			}
			avg /= time.Duration(len(store[addr.String()].Durs))
			showErrModal(fmt.Sprintf("Source: %s UDP\nAverage time: %v\nReceived: %d/%d", addr.String(), avg, len(store[addr.String()].Durs), store[addr.String()].Max))
			break
		}
	}
}

type Packet struct {
	Id    int8
	MaxId int8
	S     time.Duration
}

func NewPacket(id, max int8) *Packet {
	return &Packet{Id: id, MaxId: max}
}

func Unmarshal(src []byte) (*Packet, error) {
	p := &Packet{}
	p.Id = int8(src[0])
	p.MaxId = int8(src[1])
	var it time.Time
	err := it.UnmarshalBinary(src[2:])
	if err != nil {
		return nil, err
	}
	p.S = time.Since(it)
	return p, nil
}

type PacketBucket struct {
	Last int8
	Max  int8
	Durs []time.Duration
}

var store = make(map[string]*PacketBucket)
