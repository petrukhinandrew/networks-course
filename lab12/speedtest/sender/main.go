package main

import (
	"net"
	"strconv"
	"time"
	"unicode"

	"github.com/rivo/tview"
)

var (
	app        *tview.Application
	badIPModal *tview.Modal
	form       *tview.Form
)
var targetIPValue string = "127.0.0.1"
var targetIP net.IP
var targetPort string = "8080"
var numOfPacks int = 5
var protocol = "TCP"

func showErrModal(text string) {
	m := tview.NewModal().SetText(text).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) { app.SetRoot(form, false) })
	app.SetRoot(m, false)
}

func main() {
	badIPModal = tview.NewModal().SetText("Bad target IP").AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) { app.SetRoot(form, false) })
	form = tview.NewForm().
		AddInputField("Target IP",
			"127.0.0.1",
			20,
			func(textToCheck string, lastChar rune) bool {
				return len(textToCheck) < 16 && (lastChar == rune('.') || unicode.IsDigit(lastChar))
			},
			func(text string) {
				targetIPValue = text
			}).
		AddInputField("Target port",
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
				targetPort = text
			}).
		AddInputField("Num of packets",
			"5",
			5,
			func(textToCheck string, lastChar rune) bool {
				c, err := strconv.Atoi(textToCheck)
				if err != nil {
					return false
				}
				return c >= 0 && c < 100
			},
			func(text string) {
				c, err := strconv.Atoi(text)
				if err != nil {
					return
				}
				numOfPacks = c
			}).
		AddDropDown(
			"Protocol",
			[]string{"TCP", "UDP"},
			0,
			func(option string, optionIndex int) {
				protocol = option
			}).
		AddButton("Send", func() {
			targetIP = net.ParseIP(targetIPValue)
			if targetIP == nil {
				app.SetRoot(badIPModal, false)
				return
			}
			check()
		})
	form.SetBorder(true).SetTitle("Sender")
	app = tview.NewApplication().SetRoot(form, true).EnableMouse(true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func check() {
	switch protocol {
	case "TCP":
		checkTCP()
	case "UDP":
		checkUDP()
	default:
	}
}

func checkTCP() {
	c, err := net.Dial("tcp", net.JoinHostPort(targetIPValue, targetPort))
	if err != nil {
		showErrModal(err.Error())
		return
	}
	defer c.Close()
	for i := range numOfPacks {
		p := NewPacket(int8(i), int8(numOfPacks))
		mp, _ := p.Marshal()
		_, err = c.Write(mp)
		if err != nil {
			showErrModal(err.Error())
			return
		}
	}
	showErrModal("Sent over TCP")
}

func checkUDP() {
	c, err := net.Dial("udp", net.JoinHostPort(targetIPValue, targetPort))
	if err != nil {
		showErrModal(err.Error())
		return
	}
	defer c.Close()
	for i := range numOfPacks {
		p := NewPacket(int8(i), int8(numOfPacks))
		mp, _ := p.Marshal()
		_, err = c.Write(mp)
		if err != nil {
			showErrModal(err.Error())
			return
		}
	}
	showErrModal("Sent over UDP")
}

type Packet struct {
	Id    int8
	MaxId int8
	S     time.Time
}

func NewPacket(id, max int8) *Packet {
	return &Packet{Id: id, MaxId: max}
}

func (p *Packet) Marshal() ([]byte, error) {
	dt, err := time.Now().MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(p.Id), byte(p.MaxId)}, dt...), nil
}
