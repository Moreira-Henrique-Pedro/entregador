package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
)

func TestCreateUser(t *testing.T) {
	validBody := `{"email":"ana@condominio.com","password":"12345678","name":"Ana","role":"doorman"}`
	wantInput := &entities.User{Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}
	created := &entities.User{ID: "u1", Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}

	tests := []struct {
		name       string
		body       string
		expect     bool
		useCaseErr error
		wantStatus int
		wantError  string
	}{
		{name: "creates the user", body: validBody, expect: true, wantStatus: http.StatusCreated},
		{name: "invalid email", body: `{"email":"ana","password":"12345678","name":"Ana","role":"doorman"}`, wantStatus: http.StatusBadRequest, wantError: "email must be a valid email"},
		{name: "short password", body: `{"email":"ana@condominio.com","password":"123","name":"Ana","role":"doorman"}`, wantStatus: http.StatusBadRequest, wantError: "password must have at least 8 characters"},
		{name: "invalid role", body: `{"email":"ana@condominio.com","password":"12345678","name":"Ana","role":"resident"}`, wantStatus: http.StatusBadRequest, wantError: "role must be one of: admin, doorman"},
		{name: "already exists", body: validBody, expect: true, useCaseErr: fmt.Errorf("email: %w", entities.ErrUserAlreadyExists), wantStatus: http.StatusConflict, wantError: "user already exists"},
		{name: "unexpected error", body: validBody, expect: true, useCaseErr: errors.New("firebase down"), wantStatus: http.StatusInternalServerError, wantError: "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			create := mocks.NewCreateUser(t)
			if tt.expect {
				var out *entities.User
				if tt.useCaseErr == nil {
					out = created
				}
				create.EXPECT().Execute(mock.Anything, wantInput, "12345678").Return(out, tt.useCaseErr).Once()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(tt.body))
			serveTo(rec, NewUsersController(UsersControllerDependencies{Authorizer: testAuthorizer{}, CreateUser: create}), req)

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantError != "" {
				var body response.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.wantError, body.Error)
				return
			}

			var body response.User
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, response.User{UserID: "u1", Email: "ana@condominio.com", Name: "Ana", Role: "doorman"}, body)
			assert.NotContains(t, rec.Body.String(), "password")
		})
	}
}
