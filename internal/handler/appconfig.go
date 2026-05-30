package handler

import (
	"imagemanager/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := config.LoadAppConfig(a.AppRoot)
		if err != nil {
			jsonErr(w, 500, "Failed to load config")
			return
		}
		jsonOK(w, cfg)

	case http.MethodPost:
		var body struct {
			OutputFolder  *string `json:"output_folder"`
			DefaultFormat *string `json:"default_format"`
		}
		if err := readJSON(r, &body); err != nil {
			jsonErr(w, 400, "Invalid request body")
			return
		}

		cfg, err := config.LoadAppConfig(a.AppRoot)
		if err != nil {
			jsonErr(w, 500, "Failed to load config")
			return
		}

		if body.OutputFolder != nil {
			folder := strings.ReplaceAll(*body.OutputFolder, "\\", "/")
			folder = strings.TrimRight(folder, "/")
			if folder == "" {
				folder = filepath.ToSlash(filepath.Join(a.AppRoot, "output"))
			}
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				jsonErr(w, 422, "Output folder does not exist: "+folder)
				return
			}
			cfg.OutputFolder = folder
		}

		if body.DefaultFormat != nil {
			f := *body.DefaultFormat
			if f != "webp" && f != "jpg" && f != "png" {
				jsonErr(w, 422, "Invalid format. Allowed: webp, jpg, png")
				return
			}
			cfg.DefaultFormat = f
		}

		if err := config.SaveAppConfig(a.AppRoot, cfg); err != nil {
			jsonErr(w, 500, "Failed to save config")
			return
		}
		jsonOK(w, cfg)

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
