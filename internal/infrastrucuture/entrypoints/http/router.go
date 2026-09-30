package http

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

func NewRouter(residentHandler *ResidentHandler, deliveryHandler *DeliveryHandler, log logger.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v1/residents", residentHandler.ListResidents)
	mux.HandleFunc("GET /v1/deliveries", deliveryHandler.ListDeliveries)

	return withLogger(mux, log)
}

func withLogger(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger := log.With("method", r.Method, "path", r.URL.Path)
		ctx := requestLogger.AddToContext(r.Context(), requestLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
