package handler

import (
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func (a *App) Images(w http.ResponseWriter, r *http.Request) {
	imgRepo := repo.NewImageRepo(a.DB)

	switch r.Method {
	case http.MethodGet:
		search := r.URL.Query().Get("search")
		var tagIDs []int64
		for _, s := range r.URL.Query()["tags[]"] {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil && id > 0 {
				tagIDs = append(tagIDs, id)
			}
		}

		images, err := imgRepo.FindAll(search, tagIDs)
		if err != nil {
			jsonErr(w, 500, "Failed to load images")
			return
		}

		tagRepo := repo.NewTagRepo(a.DB)
		var imageIDs []string
		for _, img := range images {
			imageIDs = append(imageIDs, img.ID)
		}
		tagMap, _ := tagRepo.ForImages(imageIDs)

		type imageOut struct {
			ID             string      `json:"id"`
			OriginalName   string      `json:"original_name"`
			Name           string      `json:"name"`
			Filename       string      `json:"filename"`
			MimeType       string      `json:"mime_type"`
			Width          int         `json:"width"`
			Height         int         `json:"height"`
			FileSize       int64       `json:"file_size"`
			CreatedAt      string      `json:"created_at"`
			GeneratedCount int         `json:"generated_count"`
			CropCount      int         `json:"crop_count"`
			Tags           interface{} `json:"tags"`
			ThumbURL       string      `json:"thumb_url"`
		}

		out := make([]imageOut, len(images))
		for i, img := range images {
			tags := tagMap[img.ID]
			var tagsVal interface{} = tags
			if tags == nil {
				tagsVal = []model.Tag{}
			}
			out[i] = imageOut{
				ID:             img.ID,
				OriginalName:   img.OriginalName,
				Name:           img.Name,
				Filename:       img.Filename,
				MimeType:       img.MimeType,
				Width:          img.Width,
				Height:         img.Height,
				FileSize:       img.FileSize,
				CreatedAt:      img.CreatedAt,
				GeneratedCount: img.GeneratedCount,
				CropCount:      img.CropCount,
				Tags:           tagsVal,
				ThumbURL:       "image-serve?file=" + url.QueryEscape(img.Filename) + "&w=300&h=200&fit=crop",
			}
		}
		jsonOK(w, out)

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			jsonErr(w, 400, "Missing id")
			return
		}

		image, err := imgRepo.FindByID(id)
		if err != nil || image == nil {
			jsonErr(w, 404, "Image not found")
			return
		}

		// Delete source file
		sourcePath := filepath.Join(a.AppRoot, "storage", "source", image.Filename)
		os.Remove(sourcePath)

		// Delete output folders per component
		cfg, _ := config.LoadAppConfig(a.AppRoot)
		if cfg.OutputFolder != "" {
			slug := config.Slugify(image.Name)
			compRepo := repo.NewComponentRepo(a.DB)
			assignments, _ := compRepo.AllForImage(id)
			for _, ass := range assignments {
				dir := filepath.Join(cfg.OutputFolder, ass.ComponentID, slug)
				if info, err := os.Stat(dir); err == nil && info.IsDir() {
					entries, _ := os.ReadDir(dir)
					for _, e := range entries {
						os.Remove(filepath.Join(dir, e.Name()))
					}
					os.Remove(dir)
				}
			}
			// Legacy flat files
			matches, _ := filepath.Glob(filepath.Join(cfg.OutputFolder, id+"_*"))
			for _, f := range matches {
				info, err := os.Stat(f)
				if err == nil && !info.IsDir() {
					os.Remove(f)
				}
			}
		}

		// Delete thumbnail cache
		cacheDir := filepath.Join(a.AppRoot, "storage", "cache", image.Filename)
		if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(cacheDir)
			for _, e := range entries {
				os.Remove(filepath.Join(cacheDir, e.Name()))
			}
			os.Remove(cacheDir)
		}

		imgRepo.Delete(id)
		jsonOK(w, map[string]bool{"success": true})

	default:
		jsonErr(w, 405, "Method not allowed")
	}
}
