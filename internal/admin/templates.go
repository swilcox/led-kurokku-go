package admin

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/swilcox/led-kurokku-go/config"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var funcMap = template.FuncMap{
	"inc": func(i int) int { return i + 1 },
	"str": func(dt config.DisplayType) string { return string(dt) },
	"derefBool": func(b *bool) bool {
		if b == nil {
			return false
		}
		return *b
	},
	"derefInt": func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	},
	"map": func(pairs ...interface{}) map[string]interface{} {
		m := make(map[string]interface{})
		for i := 0; i+1 < len(pairs); i += 2 {
			m[pairs[i].(string)] = pairs[i+1]
		}
		return m
	},
}

// pageTemplates maps page names to their compiled template (layout + partials + page).
var pageTemplates map[string]*template.Template

// partialTemplates holds shared partials (instance_form, instance_row, widget_form).
var partialTemplates *template.Template

func initTemplates() {
	// Shared partials used in both full pages and htmx responses.
	sharedPartials := []string{
		"templates/instance_row.html",
		"templates/instance_form.html",
		"templates/widget_form.html",
	}

	// Pages that extend layout.html (each defines {{define "content"}}).
	pages := []string{
		"templates/index.html",
		"templates/config_view.html",
		"templates/config_edit.html",
		"templates/config_json.html",
	}

	// Parse partials once for htmx fragment rendering.
	partialTemplates = template.Must(
		template.New("").Funcs(funcMap).ParseFS(templateFS, sharedPartials...),
	)

	// For each page: clone layout+partials, then parse the page template.
	base := template.Must(
		template.New("").Funcs(funcMap).ParseFS(templateFS, append([]string{"templates/layout.html"}, sharedPartials...)...),
	)

	pageTemplates = make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		clone := template.Must(base.Clone())
		template.Must(clone.ParseFS(templateFS, page))
		pageTemplates[page] = clone
	}
}

// renderPage renders a full page (layout + content block).
func renderPage(w io.Writer, page string, data interface{}) error {
	tmpl, ok := pageTemplates[page]
	if !ok {
		return fmt.Errorf("unknown page template: %s", page)
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

// renderPartial renders a named partial template (for htmx fragment responses).
func renderPartial(w io.Writer, name string, data interface{}) error {
	return partialTemplates.ExecuteTemplate(w, name, data)
}

func staticHandler() http.Handler {
	return http.FileServerFS(staticFS)
}
