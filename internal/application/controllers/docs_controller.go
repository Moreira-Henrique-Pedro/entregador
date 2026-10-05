package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	docsPath = "/swagger"
	specPath = docsPath + "/openapi.yaml"
)

const swaggerUIPage = `<!doctype html>
<html lang="pt-BR">
  <head>
    <meta charset="utf-8" />
    <title>Entregador API</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({ url: "` + specPath + `", dom_id: "#swagger-ui", persistAuthorization: true })
    </script>
  </body>
</html>`

type DocsController struct {
	spec []byte
}

func NewDocsController(spec []byte) *DocsController {
	return &DocsController{spec: spec}
}

func (c *DocsController) RegisterRoutes(router gin.IRouter) {
	router.GET(docsPath, c.page)
	router.GET(specPath, c.openAPISpec)
}

func (c *DocsController) page(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIPage))
}

func (c *DocsController) openAPISpec(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/yaml", c.spec)
}
