package providers

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/api"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/middlewares"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/server"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

var ginParam = regexp.MustCompile(`:([a-zA-Z_]+)`)

func TestOpenAPISpecIsValid(t *testing.T) {
	spec := loadSpec(t)

	require.NoError(t, spec.Validate(context.Background()))
}

func TestOpenAPISpecMatchesRoutes(t *testing.T) {
	documented := specOperations(loadSpec(t))
	registered := registeredRoutes(t)

	assert.Equal(t, documented, registered, "api/openapi.yaml and the registered routes differ: update the spec")
}

func TestDocsAreNotServedInProduction(t *testing.T) {
	production := &config.Environment{}
	production.App.Env = config.EnvironmentProduction

	for _, route := range routesOf(t, production) {
		assert.NotContains(t, route, "/swagger")
	}
}

func loadSpec(t *testing.T) *openapi3.T {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromData(api.OpenAPISpec)
	require.NoError(t, err)
	return spec
}

func specOperations(spec *openapi3.T) []string {
	var operations []string
	for path, item := range spec.Paths.Map() {
		for method := range item.Operations() {
			operations = append(operations, method+" "+path)
		}
	}
	sort.Strings(operations)
	return operations
}

func registeredRoutes(t *testing.T) []string {
	var routes []string
	for _, route := range routesOf(t, &config.Environment{}) {
		if !strings.Contains(route, "/swagger") {
			routes = append(routes, route)
		}
	}
	return routes
}

func routesOf(t *testing.T, env *config.Environment) []string {
	t.Helper()
	controllers := newControllers(env, apiDependencies{authorizer: middlewares.NewOpenAuthorizer()})
	engine, ok := server.New(logger.NewNoopLogger(), nil, controllers...).(*gin.Engine)
	require.True(t, ok)

	var routes []string
	for _, route := range engine.Routes() {
		routes = append(routes, route.Method+" "+ginParam.ReplaceAllString(route.Path, "{$1}"))
	}
	sort.Strings(routes)
	return routes
}
