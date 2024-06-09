package main

import "fmt"

type FmtEmail struct {
	From        string
	To          string
	Subject     string
	ContentType string
	Body        string
}

func NewFmtEmail(from, to, subj, ct, body string) *FmtEmail {
	return &FmtEmail{from, to, subj, ct, body}
}

func (e *FmtEmail) Format() string {
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\nContent-Type: %s; charset=UTF-8\n\n%s", e.From, e.To, e.Subject, e.ContentType, e.Body)
	return msg
}
