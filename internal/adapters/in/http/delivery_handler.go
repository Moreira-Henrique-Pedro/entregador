package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeliveryHandler struct {
	listDeliveries   in.ListDeliveries
	registerDelivery in.RegisterDelivery
	deleteDelivery   in.DeleteDelivery
}

type DeliveryHandlerDependencies struct {
	ListDeliveries   in.ListDeliveries
	RegisterDelivery in.RegisterDelivery
	DeleteDelivery   in.DeleteDelivery
}

func NewDeliveryHandler(deps DeliveryHandlerDependencies) *DeliveryHandler {
	return &DeliveryHandler{
		listDeliveries:   deps.ListDeliveries,
		registerDelivery: deps.RegisterDelivery,
		deleteDelivery:   deps.DeleteDelivery,
	}
}

type deliveryRequest struct {
	Apartment   string `json:"apartment"`
	ResidentID  string `json:"resident_id"`
	PackageType string `json:"package_type"`
	Urgency     string `json:"urgency"`
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

	var status *domain.DeliveryStatus
	if raw := r.URL.Query().Get("status"); raw != "" {
		parsed := domain.DeliveryStatus(raw)
		if !parsed.IsValid() {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "status must be pending or deleted"})
			return
		}
		status = &parsed
	}

	deliveries, err := h.listDeliveries.Execute(r.Context(), apartment, status)
	if err != nil {
		logger.GetLoggerFromContext(r.Context()).Error("Failed to list deliveries", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	response := make([]deliveryResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		response = append(response, toDeliveryResponse(delivery))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *DeliveryHandler) RegisterDelivery(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeJSON[deliveryRequest](w, r)
	if !ok {
		return
	}

	delivery, err := h.registerDelivery.Execute(r.Context(), in.RegisterDeliveryInput{
		Apartment:   request.Apartment,
		ResidentID:  request.ResidentID,
		PackageType: request.PackageType,
		Urgency:     request.Urgency,
	})
	if err != nil {
		writeDeliveryError(w, r, "Failed to register delivery", err)
		return
	}

	writeJSON(w, http.StatusCreated, toDeliveryResponse(delivery))
}

func (h *DeliveryHandler) DeleteDelivery(w http.ResponseWriter, r *http.Request) {
	if err := h.deleteDelivery.Execute(r.Context(), r.PathValue("delivery_id")); err != nil {
		writeDeliveryError(w, r, "Failed to delete delivery", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeDeliveryError(w http.ResponseWriter, r *http.Request, message string, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidDelivery):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrEntityNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "delivery not found"})
	case errors.Is(err, domain.ErrNoResidentInApartment):
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: domain.ErrNoResidentInApartment.Error()})
	default:
		logger.GetLoggerFromContext(r.Context()).Error(message, "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toDeliveryResponse(delivery *domain.Delivery) deliveryResponse {
	response := deliveryResponse{
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
		response.DeletedAt = &deletedAt
	}
	return response
}
