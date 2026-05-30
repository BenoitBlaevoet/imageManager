package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/repo"
	"net/http"
	"net/url"
	"strconv"
)

func (a *App) ComponentImages(w http.ResponseWriter, r *http.Request) {
	compRepo := repo.NewComponentRepo(a.DB)

	switch r.Method {
	case http.MethodGet:
		componentID := r.URL.Query().Get("component_id")
		if componentID == "" {
			jsonErr(w, 400, "Missing component_id")
			return
		}

		rows, err := compRepo.ForComponent(componentID)
		if err != nil {
			jsonErr(w, 500, "DB error")
			return
		}

		tagRepo := repo.NewTagRepo(a.DB)
		var imageIDs []string
		for _, r := range rows {
			imageIDs = append(imageIDs, r.ImageID)
		}
		tagMap, _ := tagRepo.ForImages(imageIDs)

		images := make([]map[string]any, len(rows))
		for i, ci := range rows {
			images[i] = map[string]any{
				"assignment_id": ci.AssignmentID,
				"is_active":     ci.IsActive,
				"assigned_at":   ci.AssignedAt,
				"id":            ci.ImageID,
				"original_name": ci.OriginalName,
				"name":          ci.Name,
				"filename":      ci.Filename,
				"width":         ci.Width,
				"height":        ci.Height,
				"file_size":     ci.FileSize,
				"tags":          tagMap[ci.ImageID],
				"thumb_url":     "image-serve?file=" + url.QueryEscape(ci.Filename) + "&w=300&h=200&fit=crop",
			}
			if images[i]["tags"] == nil {
				images[i]["tags"] = []interface{}{}
			}
		}
		jsonOK(w, map[string]any{"images": images})

	case http.MethodPost:
		var body struct {
			ComponentID string `json:"component_id"`
			ImageID     string `json:"image_id"`
		}
		if err := readJSON(r, &body); err != nil || body.ComponentID == "" || body.ImageID == "" {
			jsonErr(w, 400, "Missing component_id or image_id")
			return
		}

		components, _ := config.LoadComponents(a.AppRoot)
		if config.FindComponent(components, body.ComponentID) == nil {
			jsonErr(w, 404, "Component not found")
			return
		}

		imgRepo := repo.NewImageRepo(a.DB)
		if img, _ := imgRepo.FindByID(body.ImageID); img == nil {
			jsonErr(w, 404, "Image not found")
			return
		}

		if err := compRepo.Assign(body.ComponentID, body.ImageID); err != nil {
			jsonErr(w, 500, "Failed to assign image")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	case http.MethodPatch:
		var body struct {
			AssignmentID int64 `json:"assignment_id"`
		}
		if err := readJSON(r, &body); err != nil || body.AssignmentID == 0 {
			jsonErr(w, 400, "Missing assignment_id")
			return
		}
		if err := compRepo.SetActive(body.AssignmentID); err != nil {
			jsonErr(w, 500, "DB error")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id == 0 {
			jsonErr(w, 400, "Missing id")
			return
		}
		if err := compRepo.RemoveAssignment(id); err != nil {
			jsonErr(w, 500, "DB error")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
