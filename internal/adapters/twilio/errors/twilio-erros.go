// package erros contem os erros específicos do serviço Twilio.
package errors

import "errors"

// ErrFailedToSendWhatsApp é o erro retornado quando falha ao enviar uma mensagem via WhatsApp.
var ErrFailedToSendWhatsApp = errors.New("failed to send WhatsApp message")
