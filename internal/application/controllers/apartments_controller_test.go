package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
)

func TestListApartments(t *testing.T) {
	tests := []struct {
		name       string
		out        []string
		err        error
		wantStatus int
		wantBody   string
	}{
		{name: "apartments", out: []string{"63", "101"}, wantStatus: http.StatusOK, wantBody: `["63","101"]`},
		{name: "no apartments", out: []string{}, wantStatus: http.StatusOK, wantBody: `[]`},
		{name: "use case error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantBody: `{"error":"internal server error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := mocks.NewListApartments(t)
			list.EXPECT().Execute(mock.Anything).Return(tt.out, tt.err).Once()

			rec := httptest.NewRecorder()
			serveTo(rec, NewApartmentsController(ApartmentsControllerDependencies{Authorizer: testAuthorizer{}, ListApartments: list}), httptest.NewRequest(http.MethodGet, "/v1/apartments", nil))

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
