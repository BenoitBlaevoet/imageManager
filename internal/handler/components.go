package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"net/http"
	"net/url"
)

func (a *App) Components(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		components, err := config.LoadComponents(a.AppRoot)
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}

		compRepo := repo.NewComponentRepo(a.DB)
		presets, _ := config.LoadPresets(a.AppRoot)
		// Build variant map: presetID → []variantID
		variantMap := make(map[string][]string)
		for _, p := range presets {
			var ids []string
			for _, v := range p.Variants {
				ids = append(ids, v.ID)
			}
			variantMap[p.ID] = ids
		}

		var output []any
		for _, c := range components {
			active, _ := compRepo.ActiveImage(c.ID)
			rows, _ := compRepo.ForComponent(c.ID)
			total, generated, _ := compRepo.GenerationStatusRaw(c.ID, c.Presets, variantMap)

			entry := map[string]any{
				"id":                  c.ID,
				"label":               c.Label,
				"description":         c.Description,
				"presets":             c.Presets,
				"image_count":         len(rows),
				"variants_total":      total,
				"variants_generated":  generated,
				"active_image":        nil,
			}

			if active != nil {
			entry["active_image"] = map[string]any{
				"assignment_id": active.AssignmentID,
				"id":            active.ImageID,
				"original_name": active.OriginalName,
				"name":          active.Name,
				"filename":      active.Filename,
				"width":         active.Width,
				"height":        active.Height,
				"thumb_url":     "image-serve?file=" + url.QueryEscape(active.Filename) + "&w=400&h=250&fit=crop",
			}
		}

			output = append(output, entry)
		}

		if output == nil {
			output = []any{}
		}
		jsonOK(w, map[string]any{"components": output})

	case http.MethodPost:
		var body struct {
			Components []model.Component `json:"components"`
		}
		if err := readJSON(r, &body); err != nil {
			jsonErr(w, 400, "Missing or invalid components array")
			return
		}

		presets, _ := config.LoadPresets(a.AppRoot)
		errs := config.ValidateComponents(body.Components, presets)
		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(422)
			jsonOK(w, map[string]any{"error": "Validation failed", "errors": errs})
			return
		}

		if err := config.SaveComponents(a.AppRoot, body.Components); err != nil {
			jsonErr(w, 500, "Failed to save components")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
