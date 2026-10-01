package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type residentReader interface {
	Handle(ctx context.Context, value string) ([]*entities.Resident, error)
}

type ResidentHandler struct {
	getResidentsByApartment residentReader
	getResidentsByPhone     residentReader
}

func NewResidentHandler(getResidentsByApartment, getResidentsByPhone residentReader) *ResidentHandler {
	return &ResidentHandler{
		getResidentsByApartment: getResidentsByApartment,
		getResidentsByPhone:     getResidentsByPhone,
	}
}

type residentResponse struct {
	ResidentID string    `json:"resident_id"`
	Name       string    `json:"name"`
	Apartment  string    `json:"apartment"`
	Phone      string    `json:"phone"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ListResidents handles GET /v1/residents?apartment={apartment} or GET /v1/residents?phone={phone}.
func (h *ResidentHandler) ListResidents(w http.ResponseWriter, r *http.Request) {
	apartment := r.URL.Query().Get("apartment")
	phone := r.URL.Query().Get("phone")

	var reader residentReader
	var value string
	switch {
	case apartment != "" && phone != "":
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "use either apartment or phone, not both"})
		return
	case apartment != "":
		reader, value = h.getResidentsByApartment, apartment
	case phone != "":
		reader, value = h.getResidentsByPhone, phone
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "apartment or phone query parameter is required"})
		return
	}

	residents, err := reader.Handle(r.Context(), value)
	if err != nil {
		logger.GetLoggerFromContext(r.Context()).Error("Failed to list residents", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	response := make([]residentResponse, 0, len(residents))
	for _, resident := range residents {
		response = append(response, residentResponse{
			ResidentID: resident.ResidentID,
			Name:       resident.Name,
			Apartment:  resident.Apartment,
			Phone:      resident.Phone,
			Type:       string(resident.Type),
			Status:     string(resident.Status),
			CreatedAt:  resident.CreatedAt,
			UpdatedAt:  resident.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
