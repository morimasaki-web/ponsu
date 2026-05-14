package http

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates/*.html
var templatesFS embed.FS

var htmlTemplates = template.Must(template.New("root").ParseFS(templatesFS, "templates/*.html"))

func renderHTML(w http.ResponseWriter, templateName string, data any, status int) {
	var buf bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&buf, templateName, data); err != nil {
		slog.Default().Error("template render failed", "template", templateName, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
