package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
)

func SendSock(from, pwd, to, subj, body, ct string) error {
	conn, err := createConn("smtp.gmail.com:465")
	if err != nil {
		log.Fatalf("error creating connection %s\n", err.Error())
	}
	defer conn.Close()
	msg := NewFmtEmail(from, to, subj, ct, body)
	seq := createSeq(msg, pwd)
	for _, msg := range seq {
		if err := processMsg(conn, msg); err != nil {
			log.Fatalf("error processing %s\n", err.Error())
		}
	}
	return nil
}

func createConn(addr string) (*tls.Conn, error) {
	connTcp, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn := tls.Client(connTcp, cfg)

	if conn.Handshake() != nil {
		return nil, err
	}

	return conn, nil

}

func createSeq(e *FmtEmail, pwd string) []string {
	res := []string{"HELO localhost\r\n", "AUTH LOGIN\r\n"}
	res = append(res, base64.StdEncoding.EncodeToString([]byte(e.From))+"\r\n")
	res = append(res, base64.StdEncoding.EncodeToString([]byte(pwd))+"\r\n")
	res = append(res, "MAIL FROM: <"+e.From+">\r\n")
	res = append(res, "RCPT TO: <"+e.To+">\r\n")
	res = append(res, "DATA\r\n")
	res = append(res, e.Format()+"\nsent with sockets\n"+"\r\n.\r\n")
	res = append(res, "QUIT\r\n")
	return res
}

func processMsg(conn *tls.Conn, msg string) error {
	log.Printf("> %s\n", msg)
	fmt.Fprint(conn, msg)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	log.Printf("< %s\n", string(buf[:n]))
	return nil
}
