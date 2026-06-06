package sms

import (
	"context"
	"log"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) SendCode(ctx context.Context, phone, code string) error {
	log.Printf("phone verification code for %s: %s", phone, code)
	return nil
}

func (s *LogSender) SendText(ctx context.Context, phone, text string) error {
	log.Printf("sms to %s: %s", phone, text)
	return nil
}
