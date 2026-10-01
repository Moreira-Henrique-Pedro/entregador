package http

import (
	"net/http"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/readers"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeliveryHandler struct {
	getDeliveriesByApartment readers.DeliveriesReaderPort
}

func NewDeliveryHandler(getDeliveriesByApartment readers.DeliveriesReaderPort) *DeliveryHandler {
	return &DeliveryHandler{
		getDeliveriesByApartment: getDeliveriesByApartment,
	}
}

type deliveryResponse struct {
	DeliveryID  string     `json:"delivery_id"`
	Apartment   string     `json:"apartment"`
	ResidentID  string     `json:"resident_id"`
	PackageType string     `json:"package_type"`
	Urgency     string     `json:"urgency"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func (h *DeliveryHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	apartment := r.URL.Query().Get("apartment")
	if apartment == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "apartment query parameter is required"})
		return
	}

	var status *entities.DeliveryStatus
	if raw := r.URL.Query().Get("status"); raw != "" {
		parsed := entities.DeliveryStatus(raw)
		if !parsed.IsValid() {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "status must be pending or deleted"})
			return
		}
		status = &parsed
	}

	deliveries, err := h.getDeliveriesByApartment.Handle(r.Context(), apartment, status)
	if err != nil {
		logger.GetLoggerFromContext(r.Context()).Error("Failed to list deliveries", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	response := make([]deliveryResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		item := deliveryResponse{
			DeliveryID:  delivery.DeliveryID,
			Apartment:   delivery.Apartment,
			ResidentID:  delivery.ResidentID,
			PackageType: delivery.PackageType,
			Urgency:     delivery.Urgency,
			Status:      string(delivery.Status),
			CreatedAt:   delivery.CreatedAt,
			UpdatedAt:   delivery.UpdatedAt,
		}
		if !delivery.DeleteAt.IsZero() {
			deletedAt := delivery.DeleteAt
			item.DeletedAt = &deletedAt
		}
		response = append(response, item)
	}

	writeJSON(w, http.StatusOK, response)
}
