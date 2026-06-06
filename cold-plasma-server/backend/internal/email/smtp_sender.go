package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string

	fromEmail string
	fromName  string

	implicitTLS          bool
	insecureSkipVerifyTLS bool
}

func NewSMTPSender(host, port, username, password, fromEmail, fromName string, implicitTLS bool, insecureSkipVerifyTLS bool) *SMTPSender {
	return &SMTPSender{
		host:                  strings.TrimSpace(host),
		port:                  strings.TrimSpace(port),
		username:              strings.TrimSpace(username),
		password:              password,
		fromEmail:             strings.TrimSpace(fromEmail),
		fromName:              strings.TrimSpace(fromName),
		implicitTLS:           implicitTLS,
		insecureSkipVerifyTLS: insecureSkipVerifyTLS,
	}
}

func (s *SMTPSender) Send(ctx context.Context, to string, subject string, textBody string, htmlBody string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("email: empty recipient")
	}
	if s.host == "" || s.port == "" {
		return fmt.Errorf("email: SMTP_HOST/SMTP_PORT not set")
	}
	if s.fromEmail == "" {
		s.fromEmail = s.username
	}
	if s.fromEmail == "" {
		return fmt.Errorf("email: SMTP_FROM is empty")
	}

	addr := net.JoinHostPort(s.host, s.port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var conn net.Conn
	var err error
	if s.implicitTLS {
		tlsConfig := &tls.Config{ServerName: s.host, InsecureSkipVerify: s.insecureSkipVerifyTLS}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("email dial: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(25 * time.Second))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("email client: %w", err)
	}
	defer client.Close()

	if !s.implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: s.host, InsecureSkipVerify: s.insecureSkipVerifyTLS}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("email starttls: %w", err)
			}
		}
	}

	if s.username != "" && s.password != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", s.username, s.password, s.host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("email auth: %w", err)
			}
		}
	}

	msg := buildMIMEMessage(s.fromName, s.fromEmail, to, subject, textBody, htmlBody)

	if err := client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("email from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("email rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("email data: %w", err)
	}
	_, wErr := w.Write(msg)
	cErr := w.Close()
	if wErr != nil {
		return fmt.Errorf("email write: %w", wErr)
	}
	if cErr != nil {
		return fmt.Errorf("email close: %w", cErr)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("email quit: %w", err)
	}
	return nil
}

func buildMIMEMessage(fromName, fromEmail, to, subject, textBody, htmlBody string) []byte {
	fromHeader := fromEmail
	if fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", encodeHeaderUTF8(fromName), fromEmail)
	}

	subject = encodeHeaderUTF8(subject)
	boundary := "cp_boundary_" + fmt.Sprint(time.Now().UnixNano())

	var b strings.Builder
	b.WriteString("From: " + fromHeader + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	b.WriteString("\r\n")

	// text/plain
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(textBody + "\r\n")

	// text/html
	if htmlBody != "" {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		b.WriteString("\r\n")
		b.WriteString(htmlBody + "\r\n")
	}

	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}

func sanitizeHeader(v string) string {
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	return strings.TrimSpace(v)
}

func encodeHeaderUTF8(v string) string {
	v = sanitizeHeader(v)
	if v == "" {
		return v
	}
	for i := 0; i < len(v); i++ {
		if v[i] >= 0x80 {
			return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(v)) + "?="
		}
	}
	return v
}
