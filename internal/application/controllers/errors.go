package controllers

import (
	"errors"
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/gin-gonic/gin"
)

type errorMapping struct {
	err    error
	status int
}

var errorMappings = []errorMapping{
	{entities.ErrInvalidResident, http.StatusBadRequest},
	{entities.ErrInvalidDelivery, http.StatusBadRequest},
	{entities.ErrInvalidUser, http.StatusBadRequest},
	{entities.ErrUserAlreadyExists, http.StatusConflict},
	{entities.ErrEntityNotFound, http.StatusNotFound},
	{entities.ErrOtherResidentReadOnly, http.StatusConflict},
	{entities.ErrNoResidentInApartment, http.StatusUnprocessableEntity},
}

func writeError(ctx *gin.Context, resource string, err error) {
	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.err) {
			ctx.JSON(mapping.status, response.Error{Error: errorMessage(resource, mapping, err)})
			return
		}
	}

	logger.GetLoggerFromContext(ctx.Request.Context()).Error("Request failed", "resource", resource, "error", err.Error())
	ctx.JSON(http.StatusInternalServerError, response.Error{Error: "internal server error"})
}

func errorMessage(resource string, mapping errorMapping, err error) string {
	switch mapping.status {
	case http.StatusBadRequest:
		return err.Error()
	case http.StatusNotFound:
		return resource + " not found"
	default:
		return mapping.err.Error()
	}
}
