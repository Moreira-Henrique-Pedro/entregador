package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const maxRequestBodyBytes = 1 << 20

var registerFieldNames sync.Once

func bindJSON[T any](ctx *gin.Context) (*T, bool) {
	var request T
	if err := decodeJSON(ctx, &request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: "invalid JSON body: " + err.Error()})
		return nil, false
	}
	if err := validate(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: validationMessage(err)})
		return nil, false
	}
	return &request, true
}

func bindQuery[T any](ctx *gin.Context) (*T, bool) {
	registerFieldNames.Do(useTagFieldNames)

	var query T
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: validationMessage(err)})
		return nil, false
	}
	return &query, true
}

func decodeJSON(ctx *gin.Context, target any) error {
	decoder := limitedDecoder(ctx)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func decodeLenientJSON(ctx *gin.Context, target any) error {
	return limitedDecoder(ctx).Decode(target)
}

func limitedDecoder(ctx *gin.Context) *json.Decoder {
	return json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxRequestBodyBytes))
}

func validate(target any) error {
	registerFieldNames.Do(useTagFieldNames)
	return binding.Validator.ValidateStruct(target)
}

func useTagFieldNames() {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	engine.RegisterTagNameFunc(func(field reflect.StructField) string {
		for _, tag := range []string{"json", "form"} {
			if name := strings.Split(field.Tag.Get(tag), ",")[0]; name != "" && name != "-" {
				return name
			}
		}
		return field.Name
	})
}

func validationMessage(err error) string {
	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return err.Error()
	}

	messages := make([]string, 0, len(fieldErrors))
	for _, fieldError := range fieldErrors {
		messages = append(messages, fieldMessage(fieldError))
	}
	return strings.Join(messages, "; ")
}

func fieldMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	fields := append([]string{field}, strings.Fields(strings.ToLower(fieldError.Param()))...)

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "required_without", "required_without_all":
		return fmt.Sprintf("%s is required", joinAlternatives(fields))
	case "excluded_with":
		return fmt.Sprintf("use either %s, not both", joinAlternatives(fields))
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must have at least %s characters", field, fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(fieldError.Param(), " ", ", "))
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func joinAlternatives(fields []string) string {
	last := len(fields) - 1
	return strings.Join(fields[:last], ", ") + " or " + fields[last]
}
