package docs

import (
	"embed"
)

//go:embed swagger/*
var SwaggerUI embed.FS

//go:embed openapi.yaml
var OpenAPIYaml []byte
