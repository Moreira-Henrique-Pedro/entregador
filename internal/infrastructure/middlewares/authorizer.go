package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type Authorizer struct {
	verifier services.TokenVerifier
}

func NewAuthorizer(verifier services.TokenVerifier) *Authorizer {
	return &Authorizer{verifier: verifier}
}

func (a *Authorizer) Require(roles ...entities.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log := logger.GetLoggerFromContext(ctx.Request.Context())

		user, err := a.authenticate(ctx)
		if err != nil {
			log.Warn("Rejected unauthenticated request", "error", err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Error{Error: "unauthenticated"})
			return
		}

		log = log.With("user_id", user.ID, "role", string(user.Role))
		if !user.HasAnyRole(roles...) {
			log.Warn("Rejected request without the required role")
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.Error{Error: "forbidden"})
			return
		}

		ctx.Request = ctx.Request.WithContext(log.AddToContext(ctx.Request.Context(), log))
		ctx.Next()
	}
}

func (a *Authorizer) authenticate(ctx *gin.Context) (*entities.User, error) {
	token, ok := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if !ok || token == "" {
		return nil, errors.New("missing bearer token")
	}
	return a.verifier.VerifyToken(ctx.Request.Context(), token)
}
