package web

import (
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
)

const (
	templatesBasePath = "web/templates/"
	templatesExt      = ".tmpl"
)

func (w *Web) parseTemplate(name, path string) {
	if path == "" {
		path = name
	}

	if _, ok := w.Templates[name]; ok {
		log.Panic().Str("template", name).Msg("template already parsed once")
		return
	}

	w.Templates[name] = template.Must(template.ParseFiles(templatesBasePath+name+templatesExt, templatesBasePath+"base"+templatesExt))
}

func (w *Web) templateGet(name string) *template.Template {
	if _, ok := w.Templates[name]; !ok {
		log.Error().Str("name", name).Msg("Trying to get a template that does not exists, returning a 404 page")
		return w.Templates["404.tmpl"]
	}

	return w.Templates[name]
}

func (w *Web) templateExec(rw http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := w.templateGet(name).ExecuteTemplate(rw, "base", data); err != nil {
		log.Error().Err(err).Str("name", name).Interface("data", data).Msg("failed to view template")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
