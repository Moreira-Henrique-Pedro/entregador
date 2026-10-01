package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ResidentHandler struct {
	listResidentsByApartment in.ListResidents
	listResidentsByPhone     in.ListResidents
	createResident           in.CreateResident
	updateResident           in.UpdateResident
	deleteResident           in.DeleteResident
}

type ResidentHandlerDependencies struct {
	ListResidentsByApartment in.ListResidents
	ListResidentsByPhone     in.ListResidents
	CreateResident           in.CreateResident
	UpdateResident           in.UpdateResident
	DeleteResident           in.DeleteResident
}

func NewResidentHandler(deps ResidentHandlerDependencies) *ResidentHandler {
	return &ResidentHandler{
		listResidentsByApartment: deps.ListResidentsByApartment,
		listResidentsByPhone:     deps.ListResidentsByPhone,
		createResident:           deps.CreateResident,
		updateResident:           deps.UpdateResident,
		deleteResident:           deps.DeleteResident,
	}
}

type residentRequest struct {
	Name      string `json:"name"`
	Apartment string `json:"apartment"`
	Phone     string `json:"phone"`
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

func (h *ResidentHandler) ListResidents(w http.ResponseWriter, r *http.Request) {
	apartment := r.URL.Query().Get("apartment")
	phone := r.URL.Query().Get("phone")

	var listResidents in.ListResidents
	var value string
	switch {
	case apartment != "" && phone != "":
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "use either apartment or phone, not both"})
		return
	case apartment != "":
		listResidents, value = h.listResidentsByApartment, apartment
	case phone != "":
		listResidents, value = h.listResidentsByPhone, phone
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "apartment or phone query parameter is required"})
		return
	}

	residents, err := listResidents.Execute(r.Context(), value)
	if err != nil {
		logger.GetLoggerFromContext(r.Context()).Error("Failed to list residents", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	response := make([]residentResponse, 0, len(residents))
	for _, resident := range residents {
		response = append(response, toResidentResponse(resident))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ResidentHandler) CreateResident(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeJSON[residentRequest](w, r)
	if !ok {
		return
	}

	resident, err := h.createResident.Execute(r.Context(), &domain.Resident{
		Name:      request.Name,
		Apartment: request.Apartment,
		Phone:     request.Phone,
	})
	if err != nil {
		writeResidentError(w, r, "Failed to create resident", err)
		return
	}

	writeJSON(w, http.StatusCreated, toResidentResponse(resident))
}

func (h *ResidentHandler) UpdateResident(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeJSON[residentRequest](w, r)
	if !ok {
		return
	}

	resident, err := h.updateResident.Execute(r.Context(), &domain.Resident{
		ResidentID: r.PathValue("resident_id"),
		Name:       request.Name,
		Apartment:  request.Apartment,
		Phone:      request.Phone,
	})
	if err != nil {
		writeResidentError(w, r, "Failed to update resident", err)
		return
	}

	writeJSON(w, http.StatusOK, toResidentResponse(resident))
}

func (h *ResidentHandler) DeleteResident(w http.ResponseWriter, r *http.Request) {
	if err := h.deleteResident.Execute(r.Context(), r.PathValue("resident_id")); err != nil {
		writeResidentError(w, r, "Failed to delete resident", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeResidentError(w http.ResponseWriter, r *http.Request, message string, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidResident):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrEntityNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "resident not found"})
	case errors.Is(err, domain.ErrOtherResidentReadOnly):
		writeJSON(w, http.StatusConflict, errorResponse{Error: domain.ErrOtherResidentReadOnly.Error()})
	default:
		logger.GetLoggerFromContext(r.Context()).Error(message, "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toResidentResponse(resident *domain.Resident) residentResponse {
	return residentResponse{
		ResidentID: resident.ResidentID,
		Name:       resident.Name,
		Apartment:  resident.Apartment,
		Phone:      resident.Phone,
		Type:       string(resident.Type),
		Status:     string(resident.Status),
		CreatedAt:  resident.CreatedAt,
		UpdatedAt:  resident.UpdatedAt,
	}
}
