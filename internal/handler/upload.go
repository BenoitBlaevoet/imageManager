package handler

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"imagemanager/internal/config"
	"imagemanager/internal/model"
	"imagemanager/internal/repo"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "golang.org/x/image/webp" // register WebP decode for MIME detection
)

func (a *App) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "Method not allowed")
		return
	}

	const maxSize = 20 << 20 // 20 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonErr(w, 413, "File too large (max 20 MB)")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		jsonErr(w, 400, "No file uploaded")
		return
	}
	defer file.Close()

	if header.Size > maxSize {
		jsonErr(w, 413, "File too large (max 20 MB)")
		return
	}

	// Detect MIME type from first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])

	allowed := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"image/webp": "webp",
	}
	ext, ok := allowed[mimeType]
	if !ok {
		jsonErr(w, 415, "Unsupported file type")
		return
	}

	// Seek back to start
	if _, err := file.Seek(0, 0); err != nil {
		jsonErr(w, 500, "Failed to read file")
		return
	}

	// Generate ID and save
	id := newImageID()
	filename := id + "." + ext
	destDir := filepath.Join(a.AppRoot, "storage", "source")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		jsonErr(w, 500, "Failed to create storage directory")
		return
	}
	dest := filepath.Join(destDir, filename)

	out, err := os.Create(dest)
	if err != nil {
		jsonErr(w, 500, "Failed to save file")
		return
	}
	defer out.Close()

	// Write file
	buf2 := make([]byte, 32*1024)
	if _, err := file.Seek(0, 0); err == nil {
		for {
			nr, er := file.Read(buf2)
			if nr > 0 {
				out.Write(buf2[:nr])
			}
			if er != nil {
				break
			}
		}
	}
	out.Close()

	// Read dimensions
	imgFile, err := os.Open(dest)
	if err != nil {
		jsonErr(w, 500, "Failed to read uploaded file")
		return
	}
	imgCfg, _, err := image.DecodeConfig(imgFile)
	imgFile.Close()
	if err != nil {
		jsonErr(w, 500, "Failed to decode image dimensions")
		return
	}

	// Display name
	rawName := strings.TrimSpace(r.FormValue("name"))
	if rawName == "" {
		rawName = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	img := model.Image{
		ID:           id,
		OriginalName: header.Filename,
		Name:         rawName,
		Filename:     filename,
		MimeType:     mimeType,
		Width:        imgCfg.Width,
		Height:       imgCfg.Height,
		FileSize:     header.Size,
	}

	imgRepo := repo.NewImageRepo(a.DB)
	if err := imgRepo.Insert(img); err != nil {
		jsonErr(w, 500, "Failed to save image record")
		return
	}

	// Tags
	tagRepo := repo.NewTagRepo(a.DB)
	var tagIDs []int64
	if tagsJSON := r.FormValue("tags"); tagsJSON != "" {
		var rawIDs []any
		if err := json.Unmarshal([]byte(tagsJSON), &rawIDs); err == nil {
			for _, v := range rawIDs {
				switch t := v.(type) {
				case float64:
					tagIDs = append(tagIDs, int64(t))
				case string:
					if id, err := strconv.ParseInt(t, 10, 64); err == nil {
						tagIDs = append(tagIDs, id)
					}
				}
			}
		}
	}
	if len(tagIDs) > 0 {
		tagRepo.SetForImage(id, tagIDs)
	}

	// Component assignment
	componentID := r.FormValue("component_id")
	if componentID != "" {
		components, _ := config.LoadComponents(a.AppRoot)
		if config.FindComponent(components, componentID) != nil {
			repo.NewComponentRepo(a.DB).Assign(componentID, id)
		}
	}

	tags, _ := tagRepo.ForImage(id)

	jsonOK(w, map[string]any{
		"id":            id,
		"filename":      filename,
		"original_name": header.Filename,
		"name":          rawName,
		"width":         imgCfg.Width,
		"height":        imgCfg.Height,
		"tags":          tags,
		"component_id":  nilIfEmpty(componentID),
	})
}

func newImageID() string {
	return fmt.Sprintf("img_%x.%08x", time.Now().UnixMicro(), rand.Uint32())
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
