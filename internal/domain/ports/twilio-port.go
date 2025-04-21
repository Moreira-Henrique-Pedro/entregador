// package ports contem interfaces que definem os métodos para interagir com diferentes serviços e repositórios.
package ports

import "context"

// TwilioPort é a interface que define os métodos do serviço Twilio
type TwilioPort interface {
	SendWhatsAppMessage(ctx context.Context, to string, message string) error
}
