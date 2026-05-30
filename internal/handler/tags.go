package handler

import (
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"net/http"
	"strconv"
)

func (a *App) Tags(w http.ResponseWriter, r *http.Request) {
	tags := repo.NewTagRepo(a.DB)

	switch r.Method {
	case http.MethodGet:
		all, err := tags.All()
		if err != nil {
			jsonErr(w, 500, "Failed to load tags")
			return
		}
		if all == nil {
			jsonOK(w, map[string]any{"tags": []model.Tag{}})
			return
		}
		jsonOK(w, map[string]any{"tags": all})

	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &body); err != nil || body.Name == "" {
			jsonErr(w, 400, "Tag name is required")
			return
		}
		existing, err := tags.FindByName(body.Name)
		if err != nil {
			jsonErr(w, 500, "DB error")
			return
		}
		if existing != nil {
			jsonOK(w, map[string]any{"id": existing.ID, "name": existing.Name})
			return
		}
		id, err := tags.Create(body.Name)
		if err != nil {
			jsonErr(w, 500, "Failed to create tag")
			return
		}
		jsonOK(w, map[string]any{"id": id, "name": body.Name})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id == 0 {
			jsonErr(w, 400, "Missing id")
			return
		}
		if err := tags.Delete(id); err != nil {
			jsonErr(w, 500, "Failed to delete tag")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
