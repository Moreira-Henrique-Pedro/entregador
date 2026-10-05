package controllers

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/request"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/gin-gonic/gin"
)

const PubSubPushPath = "/internal/pubsub/notifications"

type NotificationsController struct {
	notifyDelivery usecases.NotifyDelivery
	auth           gin.HandlerFunc
}

func NewNotificationsController(notifyDelivery usecases.NotifyDelivery, auth gin.HandlerFunc) *NotificationsController {
	return &NotificationsController{
		notifyDelivery: notifyDelivery,
		auth:           auth,
	}
}

func (c *NotificationsController) RegisterRoutes(router gin.IRouter) {
	handlers := []gin.HandlerFunc{c.receivePush}
	if c.auth != nil {
		handlers = append([]gin.HandlerFunc{c.auth}, handlers...)
	}
	router.POST(PubSubPushPath, handlers...)
}

func (c *NotificationsController) receivePush(ctx *gin.Context) {
	push, ok := bindPush(ctx)
	if !ok {
		return
	}

	log := pushLogger(ctx, push)

	if push.EventType() != commands.NotifyDeliveryCommandType {
		log.Warn("Ignoring Pub/Sub message with unknown event type")
		ctx.Status(http.StatusNoContent)
		return
	}

	command, err := decodeNotifyDeliveryCommand(push)
	if err != nil {
		log.Error("Invalid NotifyDelivery message", "error", err.Error(), "data", string(push.Message.Data))
		ctx.Status(http.StatusBadRequest)
		return
	}

	if err := c.notifyDelivery.Execute(log.AddToContext(ctx.Request.Context(), log), command.DeliveryID, command.NotificationType); err != nil {
		log.Error("Failed to notify delivery, Pub/Sub will retry", "error", err.Error())
		ctx.Status(http.StatusInternalServerError)
		return
	}

	log.Info("Pub/Sub message processed successfully")
	ctx.Status(http.StatusNoContent)
}

func bindPush(ctx *gin.Context) (*request.PubSubPush, bool) {
	var push request.PubSubPush
	if err := decodeLenientJSON(ctx, &push); err != nil {
		logger.GetLoggerFromContext(ctx.Request.Context()).Error("Invalid Pub/Sub push body", "error", err.Error())
		ctx.Status(http.StatusBadRequest)
		return nil, false
	}
	return &push, true
}

func decodeNotifyDeliveryCommand(push *request.PubSubPush) (*commands.NotifyDeliveryCommand, error) {
	var command commands.NotifyDeliveryCommand
	if err := push.DecodeData(&command); err != nil {
		return nil, err
	}
	return &command, validate(&command)
}

func pushLogger(ctx *gin.Context, push *request.PubSubPush) logger.Logger {
	return logger.GetLoggerFromContext(ctx.Request.Context()).With(
		"message_id", push.Message.MessageID,
		"event_type", push.EventType(),
		"message_key", push.Key(),
		"delivery_attempt", push.DeliveryAttempt,
	)
}
