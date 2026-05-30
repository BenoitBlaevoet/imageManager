package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/repo"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func (a *App) OpenFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "Method not allowed")
		return
	}

	cfg, err := config.LoadAppConfig(a.AppRoot)
	if err != nil || cfg.OutputFolder == "" {
		jsonErr(w, 400, "Output folder not configured")
		return
	}
	if _, err := os.Stat(cfg.OutputFolder); os.IsNotExist(err) {
		jsonErr(w, 400, "Output folder does not exist")
		return
	}

	var body struct {
		ComponentID string `json:"component_id"`
		ImageID     string `json:"image_id"`
	}
	readJSON(r, &body) // ignore error; body is optional

	targetFolder := cfg.OutputFolder

	if body.ComponentID != "" && body.ImageID != "" {
		imgRepo := repo.NewImageRepo(a.DB)
		image, _ := imgRepo.FindByID(body.ImageID)
		if image != nil {
			slug := config.Slugify(image.Name)
			sub := filepath.Join(cfg.OutputFolder, body.ComponentID, slug)
			if info, err := os.Stat(sub); err == nil && info.IsDir() {
				targetFolder = sub
			}
		}
	}

	openFolderOS(targetFolder)
	jsonOK(w, map[string]bool{"success": true})
}

func openFolderOS(folder string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("explorer", filepath.FromSlash(strings.ReplaceAll(folder, "/", "\\"))).Start()
	case "darwin":
		exec.Command("open", folder).Start()
	default:
		exec.Command("xdg-open", folder).Start()
	}
}
