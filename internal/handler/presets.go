package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"net/http"
)

func (a *App) Presets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		presets, err := config.LoadPresets(a.AppRoot)
		if err != nil {
			jsonErr(w, 500, "Failed to load presets")
			return
		}
		jsonOK(w, map[string]any{"presets": presets})

	case http.MethodPost:
		var body struct {
			Presets []model.Preset `json:"presets"`
		}
		if err := readJSON(r, &body); err != nil {
			jsonErr(w, 400, "Missing or invalid presets array")
			return
		}

		errs := config.ValidatePresets(body.Presets)
		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(422)
			jsonOK(w, map[string]any{"error": "Validation failed", "errors": errs})
			return
		}

		if err := config.SavePresets(a.AppRoot, body.Presets); err != nil {
			jsonErr(w, 500, "Failed to save presets")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
