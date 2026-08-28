package mailer

import (
	"fmt"
	"net/smtp"
	"os"
)

type Mailer struct {
	Host     string
	Port     string
	Username string
	Password string
}

func NewMailer() *Mailer {
	return &Mailer{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
	}
}

func (m *Mailer) SendEmail(to string, subject string, body string) error {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.Username, to, subject, body)

	err := smtp.SendMail(addr, auth, m.Username, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
