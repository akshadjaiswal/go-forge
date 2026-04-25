package generator

import (
	"regexp"
	"runtime"
	"strings"
)

// ProjectOptions holds everything the user configured via prompts or flags.
type ProjectOptions struct {
	Name       string // "my-api" — directory name and project identifier
	ModulePath string // "github.com/username/my-api" — go.mod module path
	WithAuth   bool   // include JWT auth (handlers, middleware, model, validator)
	WithDocker bool   // include Dockerfile + docker-compose
	GoVersion  string // detected from runtime: "1.22"
}

// TemplateData is the value passed into every .tmpl file during rendering.
type TemplateData struct {
	ProjectName string // "my-api"
	ModulePath  string // "github.com/username/my-api"
	PackageName string // "myapi" — valid Go identifier (no hyphens)
	GoVersion   string // "1.22"
	DBName      string // "my_api_db" — Postgres database name
	WithAuth    bool
	WithDocker  bool
}

// BuildTemplateData converts ProjectOptions into TemplateData for rendering.
func BuildTemplateData(opts ProjectOptions) TemplateData {
	return TemplateData{
		ProjectName: opts.Name,
		ModulePath:  opts.ModulePath,
		PackageName: sanitizePackageName(opts.Name),
		GoVersion:   opts.GoVersion,
		DBName:      sanitizeDBName(opts.Name),
		WithAuth:    opts.WithAuth,
		WithDocker:  opts.WithDocker,
	}
}

// DetectGoVersion returns the major.minor Go version from runtime (e.g. "1.22").
func DetectGoVersion() string {
	v := runtime.Version() // "go1.22.3"
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

// sanitizePackageName converts "my-api" → "myapi" (valid Go identifier).
func sanitizePackageName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "")
}

// sanitizeDBName converts "my-api" → "my_api_db".
func sanitizeDBName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "_") + "_db"
}
