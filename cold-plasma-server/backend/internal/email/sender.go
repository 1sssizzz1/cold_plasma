package email

import "context"

type Sender interface {
	Send(ctx context.Context, to string, subject string, textBody string, htmlBody string) error
}

