package apidocs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// specPath is relative to the working directory, which is the repo root under
// `make run` and `make dev`. The file is read on every request so edits show up
// without a rebuild.
const specPath = "docs/openapi.yaml"

const page = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>API Reference</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/docs/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true,
    });
  </script>
</body>
</html>`

type router struct {
	path string
}

// CreateRouter serves the OpenAPI spec and a Swagger UI page. Register it in
// development only.
func CreateRouter(path string) *router {
	return &router{path: path}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
	})
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Type", "application/yaml; charset=utf-8")
		c.File(specPath)
	})
}
