package internal

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Template wraps the parsed templates
type Template struct {
	tmpl *template.Template
}

// LoadTemplates parses all templates from views folder
func LoadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Render satisfies echo.Renderer interface
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}
