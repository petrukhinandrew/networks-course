package main

import (
	"errors"
	"flag"
	"io"
	"log"
)

func main() {
	modeFlag := flag.String("mode", "std", "std or sock(et)")
	fromFlag := flag.String("from", "dartmol2300@gmail.com", "from")
	pwdFlag := flag.String("pwd", "", "password for google smtp")
	toFlag := flag.String("to", "st094948@student.spbu.ru", "to")
	subjFlag := flag.String("subject", "sample subject", "subject")
	bodyFlag := flag.String("body", "sample body", "body")
	ctypeFlag := flag.String("ctype", "plain", "plain or html")
	flag.Parse()

	ctype := "text/" + *ctypeFlag
	var err error = nil
	switch *modeFlag {
	case "std":
		err = SendStd(*fromFlag, *pwdFlag, *toFlag, *subjFlag, *bodyFlag, ctype)
	case "sock":
		err = SendSock(*fromFlag, *pwdFlag, *toFlag, *subjFlag, *bodyFlag, ctype)
	default:
		err = errors.New("unsupported mode " + *modeFlag)
	}
	if err != nil && err != io.EOF {
		log.Println(err.Error())
	}
}
