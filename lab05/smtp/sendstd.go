package main

import (
	"net/smtp"
)

func SendStd(from, pwd, to, subj, body, ct string) error {
	msg := NewFmtEmail(from, to, subj, ct, body+"\nsent with net/smtp")
	return smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", from, pwd, "smtp.gmail.com"),
		from,
		[]string{to},
		[]byte(msg.Format()),
	)
}
