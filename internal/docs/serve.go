package docs

import (
    "embed"
    "net/http"
)

//go:embed openapi.yaml
var content embed.FS

// Handler serves the embedded OpenAPI spec from / (e.g., /openapi.yaml)
func Handler() http.Handler {
    return http.FileServer(http.FS(content))
}

