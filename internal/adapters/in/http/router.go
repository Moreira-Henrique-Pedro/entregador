package http

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

// PubSubPushPath receives the Pub/Sub push subscription of the notifications topic.
const PubSubPushPath = "/internal/pubsub/notifications"

type Handlers struct {
	Residents  *ResidentHandler
	Deliveries *DeliveryHandler
	// PubSubPush is optional: only mounted when the messaging provider is Pub/Sub.
	PubSubPush http.Handler
}

func NewRouter(handlers Handlers, log logger.Logger) http.Handler {
	mux := http.NewServeMux()
	residentHandler, deliveryHandler := handlers.Residents, handlers.Deliveries

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v1/residents", residentHandler.ListResidents)
	mux.HandleFunc("POST /v1/residents", residentHandler.CreateResident)
	mux.HandleFunc("PATCH /v1/residents/{resident_id}", residentHandler.UpdateResident)
	mux.HandleFunc("DELETE /v1/residents/{resident_id}", residentHandler.DeleteResident)
	mux.HandleFunc("GET /v1/deliveries", deliveryHandler.ListDeliveries)
	mux.HandleFunc("POST /v1/deliveries", deliveryHandler.RegisterDelivery)
	mux.HandleFunc("DELETE /v1/deliveries/{delivery_id}", deliveryHandler.DeleteDelivery)
	if handlers.PubSubPush != nil {
		mux.Handle("POST "+PubSubPushPath, handlers.PubSubPush)
	}

	return withLogger(mux, log)
}

func withLogger(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := log.With("method", r.Method, "path", r.URL.Path)
		ctx := requestLogger.AddToContext(r.Context(), requestLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
