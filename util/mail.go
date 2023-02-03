package util

import (
	"bytes"
	"net/smtp"
	"strconv"
)

type MailSender interface {
	Send(contentType string, subject string, body string, to ...MailUser) error
	SendText(subject string, text string, to ...MailUser) error
	SendHTML(subject string, text string, to ...MailUser) error
}

func NewSender(host string, port int, name, username, password string) MailSender {
	ms := new(Mail)
	ms.From = &sender{name, username}
	ms.Auth = smtp.PlainAuth("", username, password, host)
	ms.Addr = host + ":" + strconv.Itoa(port)
	return ms
}

type MailUser interface {
	Name() string
	Mail() string
}

type sender struct {
	name, mail string
}

func (s *sender) Name() string { return s.name }
func (s *sender) Mail() string { return s.mail }

type Mail struct {
	From MailUser
	Addr string
	Auth smtp.Auth
}

func (m *Mail) msg(contentType string, subject, body string, to []MailUser) []byte {
	var buf = new(bytes.Buffer)
	buf.WriteString("From: " + m.From.Name() + " <" + m.From.Mail() + ">" + "\r\n")
	buf.WriteString("To: " + mapRecv(to) + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("Content-Type: " + contentType + "; charset=UTF-8\r\n")
	buf.WriteString("\r\n" + body + "\r\n")
	return buf.Bytes()
}

func (m *Mail) Send(contentType string, subject, body string, to ...MailUser) error {
	msg := m.msg(contentType, subject, body, to)
	return smtp.SendMail(m.Addr, m.Auth, m.From.Mail(), mapAddr(to), msg)
}

func (m *Mail) SendText(subject string, text string, to ...MailUser) error {
	msg := m.msg("text/plain", subject, text, to)
	return smtp.SendMail(m.Addr, m.Auth, m.From.Mail(), mapAddr(to), msg)
}

func (m *Mail) SendHTML(subject string, html string, to ...MailUser) error {
	msg := m.msg("text/html", subject, html, to)
	return smtp.SendMail(m.Addr, m.Auth, m.From.Mail(), mapAddr(to), msg)
}

func mapAddr(recv []MailUser) (addrs []string) {
	addrs = make([]string, len(recv))
	for i, r := range recv {
		addrs[i] = r.Mail()
	}
	return addrs
}

func mapRecv(recv []MailUser) (strRecv string) {
	if len(recv) == 0 {
		return ""
	}

	strRecv = recv[0].Name() + " <" + recv[0].Mail() + ">"
	if len(recv) == 1 {
		return strRecv
	}

	for _, r := range recv[1:] {
		strRecv += "; " + r.Name() + " <" + r.Mail() + ">"
	}

	return strRecv
}
