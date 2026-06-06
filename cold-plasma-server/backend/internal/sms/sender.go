package sms

import "context"

type Sender interface {
	SendCode(ctx context.Context, phone, code string) error
	SendText(ctx context.Context, phone, text string) error
}
