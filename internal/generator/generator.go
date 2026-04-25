package generator

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templateFS embed.FS

// Generate creates the project directory and all files from templates.
func Generate(opts ProjectOptions) error {
	data := BuildTemplateData(opts)

	// Collect which template directories to include based on options.
	dirs := []string{"templates/base", "templates/requests"}
	if opts.WithAuth {
		dirs = append(dirs, "templates/database")
		dirs = append(dirs, "templates/auth")
	}
	if opts.WithDocker {
		dirs = append(dirs, "templates/docker")
	}

	for _, dir := range dirs {
		if err := renderDir(dir, opts.Name, data); err != nil {
			return fmt.Errorf("rendering %s: %w", dir, err)
		}
	}
	return nil
}

// renderDir walks all .tmpl files in a template directory and writes them
// into the output project directory.
func renderDir(tmplDir, projectName string, data TemplateData) error {
	return fs.WalkDir(templateFS, tmplDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Derive output path:
		// "templates/base/cmd/api/main.go.tmpl" → "my-api/cmd/api/main.go"
		outPath := deriveOutputPath(projectName, tmplDir, path)

		content, err := renderTemplate(path, data)
		if err != nil {
			return fmt.Errorf("rendering template %s: %w", path, err)
		}

		return writeFile(outPath, content)
	})
}

// deriveOutputPath strips the template dir prefix and .tmpl suffix,
// then prepends the project name directory.
//
// Example:
//   tmplDir  = "templates/base"
//   path     = "templates/base/cmd/api/main.go.tmpl"
//   result   = "my-api/cmd/api/main.go"
func deriveOutputPath(projectName, tmplDir, path string) string {
	// Strip the template directory prefix (e.g. "templates/base/")
	rel := strings.TrimPrefix(path, tmplDir+"/")
	// Strip .tmpl extension
	rel = strings.TrimSuffix(rel, ".tmpl")
	return filepath.Join(projectName, rel)
}

// renderTemplate reads a .tmpl file from the embedded FS and renders it
// with [[ ]] delimiters to avoid conflicts with Go's {{ }} syntax.
func renderTemplate(tmplPath string, data TemplateData) ([]byte, error) {
	raw, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return nil, err
	}

	// Use [[ ]] delimiters — avoids collisions with Go struct/map syntax in
	// the generated source files (e.g. map[string]interface{}{}).
	t, err := template.New(filepath.Base(tmplPath)).
		Delims("[[", "]]").
		Parse(string(raw))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeFile creates all parent directories and writes content to outPath.
func writeFile(outPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outPath, content, 0644)
}

