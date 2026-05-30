package handler

import (
	"fmt"
	imgpkg "imagemanager/internal/imaging"
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

func (a *App) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "Method not allowed")
		return
	}

	var body struct {
		ImageID     string  `json:"image_id"`
		PresetID    string  `json:"preset_id"`
		VariantID   *string `json:"variant_id"`
		ComponentID *string `json:"component_id"`
	}
	if err := readJSON(r, &body); err != nil {
		jsonErr(w, 400, "Invalid request body")
		return
	}

	if body.ImageID == "" || body.PresetID == "" {
		jsonErr(w, 400, "Missing image_id or preset_id")
		return
	}

	imgRepo := repo.NewImageRepo(a.DB)
	cropRepo := repo.NewCropRepo(a.DB)

	image, err := imgRepo.FindByID(body.ImageID)
	if err != nil || image == nil {
		jsonErr(w, 404, "Image not found")
		return
	}

	presets, err := config.LoadPresets(a.AppRoot)
	if err != nil {
		jsonErr(w, 500, "Failed to load presets")
		return
	}
	preset := config.FindPreset(presets, body.PresetID)
	if preset == nil {
		jsonErr(w, 400, "Unknown preset")
		return
	}

	cfg, err := config.LoadAppConfig(a.AppRoot)
	if err != nil {
		jsonErr(w, 500, "Failed to load config")
		return
	}

	if cfg.OutputFolder == "" {
		jsonErr(w, 400, "Output folder not configured")
		return
	}
	if _, err := os.Stat(cfg.OutputFolder); os.IsNotExist(err) {
		jsonErr(w, 400, "Output folder does not exist")
		return
	}

	imageSlug := config.Slugify(image.Name)
	componentID := ""
	if body.ComponentID != nil {
		componentID = *body.ComponentID
	}

	var subDir string
	if componentID != "" {
		subDir = filepath.Join(cfg.OutputFolder, componentID, imageSlug)
	} else {
		subDir = filepath.Join(cfg.OutputFolder, imageSlug)
	}

	if err := os.MkdirAll(subDir, 0755); err != nil {
		jsonErr(w, 500, "Failed to create output directory")
		return
	}

	// Load source image once
	src, err := loadSourceImage(a.AppRoot, image.Filename, image.MimeType)
	if err != nil {
		jsonErr(w, 500, "Failed to load source image")
		return
	}

	// Select variants to process
	variants := preset.Variants
	if body.VariantID != nil && *body.VariantID != "" {
		var onlyVariants []model.Variant
		for _, v := range preset.Variants {
			if v.ID == *body.VariantID {
				onlyVariants = append(onlyVariants, v)
			}
		}
		variants = onlyVariants
	}

	type generatedEntry struct {
		VariantID string            `json:"variant_id"`
		Files     map[string]string `json:"files"`
		Folder    string            `json:"folder"`
	}

	var generated []generatedEntry

	for _, variant := range variants {
		// Get or auto-center crop
		savedCrop, _ := cropRepo.Find(body.ImageID, body.PresetID, variant.ID)

		var cropX, cropY, cropW, cropH int
		if savedCrop != nil {
			cropX = savedCrop.CropX
			cropY = savedCrop.CropY
			cropW = savedCrop.CropW
			cropH = savedCrop.CropH
		} else {
			ratio := float64(variant.Width) / float64(variant.Height)
			srcW := float64(image.Width)
			srcH := float64(image.Height)

			if srcW/srcH > ratio {
				cropH = image.Height
				cropW = int(math.Round(srcH * ratio))
			} else {
				cropW = image.Width
				cropH = int(math.Round(srcW / ratio))
			}
			cropX = int(math.Round((srcW - float64(cropW)) / 2))
			cropY = int(math.Round((srcH - float64(cropH)) / 2))
		}

		format := variant.Format
		if format == "" {
			format = preset.Format
		}
		if format == "" {
			format = cfg.DefaultFormat
		}

		ext := imgpkg.OutputExt(format)
		base := body.PresetID + "_" + variant.ID
		files := make(map[string]string)

		for _, dpr := range []int{1, 2} {
			targetW := variant.Width * dpr
			targetH := variant.Height * dpr
			suffix := fmt.Sprintf("%dx", dpr)
			outName := base + "_" + suffix + "." + ext
			outPath := filepath.Join(subDir, outName)

			out := imgpkg.CropAndResize(src, cropX, cropY, cropW, cropH, targetW, targetH)

			outFile, err := os.Create(outPath)
			if err != nil {
				jsonErr(w, 500, "Failed to create output file")
				return
			}
			if err := imgpkg.EncodeImage(outFile, out, format); err != nil {
				outFile.Close()
				jsonErr(w, 500, "Failed to encode output image")
				return
			}
			outFile.Close()

			files[suffix] = outName
		}

		cropRepo.Upsert(body.ImageID, body.PresetID, variant.ID, cropX, cropY, cropW, cropH)
		cropRepo.MarkGenerated(body.ImageID, body.PresetID, variant.ID)

		generated = append(generated, generatedEntry{
			VariantID: variant.ID,
			Files:     files,
			Folder:    subDir,
		})
	}

	jsonOK(w, map[string]any{"generated": generated})
}

