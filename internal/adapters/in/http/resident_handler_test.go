package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListResidents(t *testing.T) {
	resident := &domain.Resident{
		ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: domain.ResidentTypePrimary, Status: domain.ResidentStatusCreated,
	}

	tests := []struct {
		name          string
		query         string
		wantApartment string
		wantPhone     string
		readerOut     []*domain.Resident
		readerErr     error
		wantStatus    int
		wantCount     int
	}{
		{name: "by apartment", query: "?apartment=101", wantApartment: "101", readerOut: []*domain.Resident{resident}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "by phone", query: "?phone=11999999999", wantPhone: "11999999999", readerOut: []*domain.Resident{resident}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "empty result returns empty list", query: "?apartment=999", wantApartment: "999", wantStatus: http.StatusOK, wantCount: 0},
		{name: "missing params", query: "", wantStatus: http.StatusBadRequest},
		{name: "both params", query: "?apartment=101&phone=1", wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?phone=1", wantPhone: "1", readerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byApartment := mocks.NewListResidents(t)
			byPhone := mocks.NewListResidents(t)
			if tt.wantApartment != "" {
				byApartment.EXPECT().Execute(mock.Anything, tt.wantApartment).Return(tt.readerOut, tt.readerErr).Once()
			}
			if tt.wantPhone != "" {
				byPhone.EXPECT().Execute(mock.Anything, tt.wantPhone).Return(tt.readerOut, tt.readerErr).Once()
			}

			rec := httptest.NewRecorder()
			NewResidentHandler(ResidentHandlerDependencies{ListResidentsByApartment: byApartment, ListResidentsByPhone: byPhone}).ListResidents(rec, httptest.NewRequest(http.MethodGet, "/v1/residents"+tt.query, nil))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				var body errorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body.Error)
				return
			}

			var body []residentResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Len(t, body, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, residentResponse{
					ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
					Type: "resident-primary", Status: "created",
				}, body[0])
			}
		})
	}
}

func TestCreateResident(t *testing.T) {
	created := &domain.Resident{
		ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: domain.ResidentTypePrimary, Status: domain.ResidentStatusCreated,
	}

	tests := []struct {
		name       string
		body       string
		wantInput  *domain.Resident
		writerErr  error
		wantStatus int
	}{
		{
			name:       "creates the resident",
			body:       `{"name":"Ana","apartment":"101","phone":"11999999999"}`,
			wantInput:  &domain.Resident{Name: "Ana", Apartment: "101", Phone: "11999999999"},
			wantStatus: http.StatusCreated,
		},
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"name":"Ana","type":"other"}`, wantStatus: http.StatusBadRequest},
		{
			name:       "validation error",
			body:       `{"name":"Ana"}`,
			wantInput:  &domain.Resident{Name: "Ana"},
			writerErr:  fmt.Errorf("%w: apartment is required", domain.ErrInvalidResident),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "writer error",
			body:       `{"name":"Ana","apartment":"101"}`,
			wantInput:  &domain.Resident{Name: "Ana", Apartment: "101"},
			writerErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			create := mocks.NewCreateResident(t)
			if tt.wantInput != nil {
				out := created
				if tt.writerErr != nil {
					out = nil
				}
				create.EXPECT().Execute(mock.Anything, tt.wantInput).Return(out, tt.writerErr).Once()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/residents", strings.NewReader(tt.body))
			NewResidentHandler(ResidentHandlerDependencies{CreateResident: create}).CreateResident(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusCreated {
				assertErrorBody(t, rec)
				return
			}

			var body residentResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, residentResponse{
				ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
				Type: "resident-primary", Status: "created",
			}, body)
		})
	}
}

func TestUpdateResident(t *testing.T) {
	updated := &domain.Resident{ResidentID: "r1", Name: "Ana Maria", Apartment: "101", Type: domain.ResidentTypePrimary}

	tests := []struct {
		name       string
		body       string
		wantInput  *domain.Resident
		writerErr  error
		wantStatus int
	}{
		{
			name:       "updates the resident",
			body:       `{"name":"Ana Maria"}`,
			wantInput:  &domain.Resident{ResidentID: "r1", Name: "Ana Maria"},
			wantStatus: http.StatusOK,
		},
		{name: "invalid JSON", body: `nope`, wantStatus: http.StatusBadRequest},
		{
			name:       "validation error",
			body:       `{}`,
			wantInput:  &domain.Resident{ResidentID: "r1"},
			writerErr:  domain.ErrInvalidResident,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			body:       `{"name":"x"}`,
			wantInput:  &domain.Resident{ResidentID: "r1", Name: "x"},
			writerErr:  fmt.Errorf("resident r1: %w", domain.ErrEntityNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "other resident",
			body:       `{"name":"x"}`,
			wantInput:  &domain.Resident{ResidentID: "r1", Name: "x"},
			writerErr:  domain.ErrOtherResidentReadOnly,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "writer error",
			body:       `{"name":"x"}`,
			wantInput:  &domain.Resident{ResidentID: "r1", Name: "x"},
			writerErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := mocks.NewUpdateResident(t)
			if tt.wantInput != nil {
				out := updated
				if tt.writerErr != nil {
					out = nil
				}
				update.EXPECT().Execute(mock.Anything, tt.wantInput).Return(out, tt.writerErr).Once()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/v1/residents/r1", strings.NewReader(tt.body))
			req.SetPathValue("resident_id", "r1")
			NewResidentHandler(ResidentHandlerDependencies{UpdateResident: update}).UpdateResident(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				assertErrorBody(t, rec)
				return
			}

			var body residentResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, residentResponse{ResidentID: "r1", Name: "Ana Maria", Apartment: "101", Type: "resident-primary"}, body)
		})
	}
}

func TestDeleteResident(t *testing.T) {
	tests := []struct {
		name       string
		writerErr  error
		wantStatus int
	}{
		{name: "deletes the resident", wantStatus: http.StatusNoContent},
		{name: "not found", writerErr: domain.ErrEntityNotFound, wantStatus: http.StatusNotFound},
		{name: "other resident", writerErr: domain.ErrOtherResidentReadOnly, wantStatus: http.StatusConflict},
		{name: "writer error", writerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			del := mocks.NewDeleteResident(t)
			del.EXPECT().Execute(mock.Anything, "r1").Return(tt.writerErr).Once()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/v1/residents/r1", nil)
			req.SetPathValue("resident_id", "r1")
			NewResidentHandler(ResidentHandlerDependencies{DeleteResident: del}).DeleteResident(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus == http.StatusNoContent {
				assert.Empty(t, rec.Body.String())
				return
			}
			assertErrorBody(t, rec)
		})
	}
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body errorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Error)
}
