package main

import (
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func NewPacket(seq int) *icmp.Message {
	return &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  seq,
			Data: []byte("qwerty"),
		},
	}
}

func WriteMessageTo(packet *icmp.Message, conn *icmp.PacketConn, addr *net.IPAddr) error {
	data, err := (packet).Marshal(nil)
	if err != nil {
		return err
	}

	_, err = conn.WriteTo(data, addr)
	if err != nil {
		return err
	}
	return nil
}

func BodySum(packet *icmp.Message) int {
	res := 0
	sumData, _ := packet.Body.Marshal(0)
	for _, elem := range sumData {
		res += int(elem)
	}
	return res
}

func ReadAndParseMessage(conn *icmp.PacketConn) (*icmp.Message, error) {
	resp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err := conn.ReadFrom(resp)

	if err != nil {
		return nil, err
	}

	responseData, err := icmp.ParseMessage(1, resp)
	if err != nil {
		return nil, err
	}
	return responseData, nil

}
func main() {
	addr, err := net.ResolveIPAddr("ip4", os.Args[1])
	if err != nil {
		log.Fatalln("resolve: ", err.Error())
	}
	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		log.Fatalln("listen:", err.Error())
	}
	defer conn.Close()

	seq := 0
	for {
		start := time.Now()
		packet := NewPacket(seq)

		if err := WriteMessageTo(packet, conn, addr); err != nil {
			log.Println("write: ", err.Error())
			continue
		}

		seq += 1
		sum := BodySum(packet)
		resp, err := ReadAndParseMessage(conn)
		if err != nil {
			log.Println("response: ", err.Error())
			continue
		}

		if sum+resp.Checksum != 65535 && packet.Checksum != 0 {
			log.Println("checksum: bad checksum")
			continue
		} else if resp.Checksum == 0 {
			log.Println("checksum: ignore checksum")
		}

		log.Printf("icmp seq=%d time=%v\n", seq, time.Since(start))
		time.Sleep(time.Second)
	}
}
