package handlers

import (
	"net/http"
)

// ComponentsExample renders the components example page
func (h *Handler) ComponentsExample(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"title":       "UI Bileşen Örnekleri",
		"description": "KolajAI Uygulaması UI Bileşen Örnekleri",
		"app_name":    "KolajAI",
		"version":     "1.0.0",
	}

	h.RenderTemplate(w, r, "components/example", data)
}
