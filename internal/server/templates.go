// Package server internal server Templates
package server

import (
	"html/template"
	"io"
	"os"
	"path"

	"github.com/labstack/echo/v4"
)

const (
	internalDirName  = "internal"
	templatesDirName = "templates"
	htmlRegex        = "*.html"
)

// Template represents a struct that holds a collection of parsed templates.
type Template struct {
	templates *template.Template
}

// Render renders the specified template with the provided data and writes the output to the given writer.
// It uses the ExecuteTemplate method of the template package to render the template.
//
// Parameters:
// - w: The io.Writer to which the rendered template will be written.
// - name: The name of the template to render.
// - data: The data to be passed to the template for rendering.
// - c: The echo.Context for the current request (not used in this method but required by the interface).
//
// Returns:
// - An error if the template execution fails, otherwise nil.
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// ConfigureRenderer configures the server's renderer by parsing all HTML templates in the specified directory
// and setting the server's router renderer to the parsed templates.
//
// This function assumes that the templates are located in a directory named "templates" within the "internal" directory,
// and that the templates have the ".html" extension.
//
// Example Directory Structure:
// - internal/
//   - templates/
//   - index.html
//   - about.html
//
// The function uses the template.ParseGlob method to parse all matching templates and stores them in a Template struct.
// The Template struct is then set as the renderer for the server's router.
func (s *Server) ConfigureRenderer() {
	rootDir, _ := os.Getwd()
	t := &Template{
		templates: template.Must(template.ParseGlob(
			path.Join(
				rootDir,
				internalDirName,
				templatesDirName,
				htmlRegex,
			),
		)),
	}
	s.router.Renderer = t
}
