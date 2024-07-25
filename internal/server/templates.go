package server

import (
	"html/template"
	"io"
	"path"

	"github.com/labstack/echo"
)

const (
	internalDirName  = "internal"
	templatesDirName = "templates"
	htmlRegex        = "*.html"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (s *Server) ConfigureRenderer() {
	t := &Template{
		templates: template.Must(template.ParseGlob(
			path.Join(
				internalDirName,
				templatesDirName,
				htmlRegex,
			),
		)),
	}
	s.router.Renderer = t
}
