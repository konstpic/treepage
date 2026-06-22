package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type Email struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewEmailFromEnv() *Email {
	host := strings.TrimSpace(os.Getenv("NOTIFY_SMTP_HOST"))
	if host == "" {
		return nil
	}
	port := strings.TrimSpace(os.Getenv("NOTIFY_SMTP_PORT"))
	if port == "" {
		port = "587"
	}
	from := strings.TrimSpace(os.Getenv("NOTIFY_SMTP_FROM"))
	if from == "" {
		from = "treepage@localhost"
	}
	return &Email{
		host: host,
		port: port,
		user: os.Getenv("NOTIFY_SMTP_USER"),
		pass: os.Getenv("NOTIFY_SMTP_PASSWORD"),
		from: from,
	}
}

func (e *Email) Notify(ctx context.Context, userEmail string, p Payload) {
	if e == nil || strings.TrimSpace(userEmail) == "" {
		return
	}
	to := []string{userEmail}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", userEmail, p.Title, p.Body))
	addr := e.host + ":" + e.port
	var auth smtp.Auth
	if e.user != "" {
		auth = smtp.PlainAuth("", e.user, e.pass, e.host)
	}
	// Best-effort; ignore errors (notification is auxiliary).
	if e.port == "465" {
		_ = sendMailTLS(addr, auth, e.from, to, msg)
		return
	}
	_ = smtp.SendMail(addr, auth, e.from, to, msg)
}

func sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host := strings.Split(addr, ":")[0]
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}
