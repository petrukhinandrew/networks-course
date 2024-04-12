package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"time"
)

func runClient(serverPort int, clientPort int, timeout int, path string) {

	conn, err := setupConnection(serverPort, clientPort)
	if err != nil {
		log.Fatalf("error establishing connection: %s", err.Error())
	}
	defer conn.Close()

	log.Printf("strating client %s -> %s", conn.LocalAddr().String(), conn.RemoteAddr().String())

	recv := make(chan Packet)
	recvErr := make(chan error)
	timeLimiter := time.NewTicker(time.Second * time.Duration(timeout))

	go pollRead(conn, recv, recvErr)
	time.Sleep(time.Millisecond)
	go sendFile(conn, path, timeLimiter, recv, recvErr)
	time.Sleep(time.Millisecond)
}

func sendFile(conn *net.UDPConn, path string, timeLimiter *time.Ticker, recv chan Packet, recvErr chan error) {
	const chunkSize = 4

	fileChunks, err := getFileChunks(path, chunkSize)
	if err != nil {
		log.Fatalf("error reading file chunks: %s", err.Error())
	}
	for chunkIdx := -1; chunkIdx < len(fileChunks); {
		if chunkIdx == -1 {
			err = trySendMeta(conn, chunkIdx%2, uint64(len(fileChunks))*uint64(chunkSize))
		} else {
			err = trySendChunk(conn, chunkIdx%2, fileChunks[chunkIdx])
		}
		if err != nil {
			log.Printf("error writing packet: %s", err.Error())
		}
		log.Printf("sent chunk %d", chunkIdx)
		select {
		case <-timeLimiter.C:
			log.Println("read timeout reached, trying again")
		case p := <-recv:
			log.Printf("got response `%s`", string(p.data))
			if !validateReceivedPacket(p, chunkIdx%2) {
				log.Printf("received bad packet, index does not match")
				continue
			}
			chunkIdx += 1
		case rErr := <-recvErr:
			log.Printf("error reading response: %s", rErr.Error())
		}
	}
}

func setupConnection(serverPort int, clientPort int) (conn *net.UDPConn, err error) {
	udpServer, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		return nil, err
	}

	udpClient, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", clientPort))
	if err != nil {
		return nil, err
	}

	conn, err = net.DialUDP("udp", udpClient, udpServer)

	return conn, err
}

func getFileChunks(path string, chunkSize int) ([][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileBytes := make([]byte, 1024)
	if n, err := file.Read(fileBytes); err != nil {
		return nil, err
	} else {
		fileBytes = fileBytes[:n]
	}

	fileChunks := [][]byte{}
	for idx, b := range fileBytes {
		if idx%chunkSize == 0 {
			fileChunks = append(fileChunks, []byte{})
		}
		fileChunks[len(fileChunks)-1] = append(fileChunks[len(fileChunks)-1], b)
	}

	return fileChunks, nil
}

func pollRead(conn *net.UDPConn, recv chan Packet, recvErr chan error) {
	for {
		resp := make([]byte, 1024)
		n, err := conn.Read(resp)
		if err != nil {
			recvErr <- err
			continue
		}
		if p, err := DecodePacket(resp[:n]); err != nil {
			recvErr <- err
		} else {
			recv <- *p
		}
	}
}

func trySendChunk(conn *net.UDPConn, idx int, chunk []byte) error {
	p := EncodePacket(idx, chunk)
	_, err := conn.Write(p.Bytes())
	return err
}

func trySendMeta(conn *net.UDPConn, idx int, fLen uint64) error {
	encFileLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(encFileLen, fLen)

	body := append([]byte("META"), encFileLen...)
	fmt.Printf("%d %b\n", fLen, body)
	p := EncodePacket(idx, body)
	_, err := conn.Write(p.Bytes())
	return err
}

func validateReceivedPacket(p Packet, lastSentIdx int) bool {
	return p.index == uint8(lastSentIdx) && slices.Equal(p.data, []byte("ACK"))
}
