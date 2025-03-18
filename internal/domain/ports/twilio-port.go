package ports

import "context"

type TwilioPort interface {
	SendWhatsAppMessage(ctx context.Context, to string) error
}
