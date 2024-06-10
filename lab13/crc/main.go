package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
)

func main() {
	pathFlag := flag.String("src", "sample.txt", "path to file with data")
	flag.Parse()

	f, err := os.Open(*pathFlag)
	if err != nil {
		log.Fatalf("error on src open %s\n", err.Error())
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("error on src read %s\n", err.Error())
	}
	for len(buf)%5 != 0 {
		buf = append(buf, byte(0))
	}
	for i := 0; 5*(i+1) <= len(buf); i++ {
		processPacket(buf[5*i : 5*(i+1)])
	}
}
func processPacket(data []byte) {
	crc := makeCRC5(data)
	fmt.Printf("Text: %s\nRaw : %v\nCRC : %v\n", string(data), data, crc)
	packet := transfer(append(data, crc))
	data = packet[:len(packet)-1]
	fmt.Printf("=== Tranfer ===\nText: %s\nRaw : %v\nExpected CRC : %v\n", string(data), data, packet[len(packet)-1])
	check := chechCRC5(packet)
	fmtCheck := "OK"
	if !check {
		fmtCheck = "ERROR"
	}
	fmt.Printf("CRC status: %s\n==========\n", fmtCheck)

}
func makeCRC5(data []byte) byte {
	const poly byte = 0x85
	b := data[0]
	for i := range 8 * (len(data) - 1) {
		doXor := (b>>7)&0x1 == 0x1
		if doXor {
			b ^= poly
		}
		b <<= 1
		b |= data[i/8+1] & (0x1 << (7 - i%7))
	}
	return b
}

func transfer(data []byte) []byte {
	if rand.Float32() < 0.6 {
		return data
	}
	i := rand.Intn(len(data) - 1)
	j := rand.Intn(8)
	data[i] ^= 0x1 << j
	return data
}

func chechCRC5(data []byte) bool {
	rem := data[len(data)-1]
	res := makeCRC5(data[:len(data)-1])
	return rem == res
}
