package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"net/http"
	"os"
	"path/filepath"
)

func (a *App) Crops(w http.ResponseWriter, r *http.Request) {
	imgRepo := repo.NewImageRepo(a.DB)
	cropRepo := repo.NewCropRepo(a.DB)

	switch r.Method {
	case http.MethodGet:
		imageID := r.URL.Query().Get("image_id")
		componentID := r.URL.Query().Get("component_id")

		if imageID == "" {
			jsonErr(w, 400, "Missing image_id")
			return
		}

		image, err := imgRepo.FindByID(imageID)
		if err != nil || image == nil {
			jsonErr(w, 404, "Image not found")
			return
		}

		savedCrops, _ := cropRepo.ForImage(imageID)
		cropMap := make(map[string]map[string]any)
		for _, c := range savedCrops {
			key := c.PresetID + ":" + c.VariantID
			cropMap[key] = map[string]any{
				"crop_x":       c.CropX,
				"crop_y":       c.CropY,
				"crop_w":       c.CropW,
				"crop_h":       c.CropH,
				"generated_at": c.GeneratedAt,
			}
		}

		cfg, _ := config.LoadAppConfig(a.AppRoot)
		allPresets, _ := config.LoadPresets(a.AppRoot)

		// Filter by component if provided
		if componentID != "" {
			components, _ := config.LoadComponents(a.AppRoot)
			comp := config.FindComponent(components, componentID)
			if comp != nil {
				filtered := make([]model.Preset, 0)
				for _, p := range allPresets {
					for _, pid := range comp.Presets {
						if p.ID == pid {
							filtered = append(filtered, p)
							break
						}
					}
				}
				allPresets = filtered
			}
		}

		result := make([]any, 0)
		for _, preset := range allPresets {
			presetFormat := preset.Format
			if presetFormat == "" {
				presetFormat = cfg.DefaultFormat
			}

			variants := make([]any, 0)
			for _, variant := range preset.Variants {
				key := preset.ID + ":" + variant.ID
				saved := cropMap[key]

				format := variant.Format
				if format == "" {
					format = presetFormat
				}
				ext := format

				base := componentID + "_" + variant.ID
				if componentID == "" {
					base = imageID + "_" + preset.ID + "_" + variant.ID
				}
				file1x := base + "_1x." + ext
				file2x := base + "_2x." + ext

				filesExist := false
				if cfg.OutputFolder != "" {
					_, err := os.Stat(filepath.Join(cfg.OutputFolder, file1x))
					filesExist = err == nil
				}

				var cropObj any
				var generatedAt any
				if saved != nil {
					cropObj = map[string]int{
						"x": saved["crop_x"].(int),
						"y": saved["crop_y"].(int),
						"w": saved["crop_w"].(int),
						"h": saved["crop_h"].(int),
					}
					generatedAt = saved["generated_at"]
				}

				variants = append(variants, map[string]any{
					"id":           variant.ID,
					"label":        variant.Label,
					"width":        variant.Width,
					"height":       variant.Height,
					"format":       format,
					"crop":         cropObj,
					"generated_at": generatedAt,
					"files": map[string]string{
						"1x": file1x,
						"2x": file2x,
					},
					"files_exist": filesExist,
				})
			}

			result = append(result, map[string]any{
				"id":       preset.ID,
				"label":    preset.Label,
				"variants": variants,
			})
		}

		jsonOK(w, map[string]any{"presets": result})

	case http.MethodPost:
		var body struct {
			ImageID   string `json:"image_id"`
			PresetID  string `json:"preset_id"`
			VariantID string `json:"variant_id"`
			CropX     int    `json:"crop_x"`
			CropY     int    `json:"crop_y"`
			CropW     int    `json:"crop_w"`
			CropH     int    `json:"crop_h"`
		}
		if err := readJSON(r, &body); err != nil {
			jsonErr(w, 400, "Invalid request body")
			return
		}

		if body.ImageID == "" || body.PresetID == "" || body.VariantID == "" {
			jsonErr(w, 400, "Missing required fields")
			return
		}

		if img, _ := imgRepo.FindByID(body.ImageID); img == nil {
			jsonErr(w, 404, "Image not found")
			return
		}

		presets, _ := config.LoadPresets(a.AppRoot)
		preset := config.FindPreset(presets, body.PresetID)
		if preset == nil || config.FindVariant(preset, body.VariantID) == nil {
			jsonErr(w, 400, "Unknown preset/variant")
			return
		}

		if body.CropW <= 0 || body.CropH <= 0 {
			jsonErr(w, 422, "Invalid crop dimensions")
			return
		}

		if err := cropRepo.Upsert(body.ImageID, body.PresetID, body.VariantID, body.CropX, body.CropY, body.CropW, body.CropH); err != nil {
			jsonErr(w, 500, "Failed to save crop")
			return
		}
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
