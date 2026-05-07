package handler

import (
	"html/template"
	"net/http"
)

// HomeHandler 首页处理器
type HomeHandler struct {
	Tmpl *template.Template
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.Tmpl.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
