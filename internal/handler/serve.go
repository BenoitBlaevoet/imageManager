package handler

import (
	"image"
	imgpkg "imagemanager/internal/imaging"
	"imagemanager/internal/repo"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var filenameRe = regexp.MustCompile(`^img_[a-zA-Z0-9._-]+$`)

func (a *App) ImageServe(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if !filenameRe.MatchString(file) {
		http.Error(w, "Invalid filename", 400)
		return
	}

	imgRepo := repo.NewImageRepo(a.DB)
	imgRec, err := imgRepo.FindByFilename(file)
	if err != nil || imgRec == nil {
		http.Error(w, "Image not found", 404)
		return
	}

	targetW, _ := strconv.Atoi(r.URL.Query().Get("w"))
	targetH, _ := strconv.Atoi(r.URL.Query().Get("h"))
	fit := r.URL.Query().Get("fit")
	if fit == "" {
		fit = "max"
	}

	// Cache key: {w}x{h}_{fit}.jpg
	cacheKey := strconv.Itoa(targetW) + "x" + strconv.Itoa(targetH) + "_" + fit + ".jpg"
	cacheDir := filepath.Join(a.AppRoot, "storage", "cache", file)
	cachePath := filepath.Join(cacheDir, cacheKey)

	if data, err := os.ReadFile(cachePath); err == nil {
		serveJPEG(w, data)
		return
	}

	src, err := loadSourceImage(a.AppRoot, file, imgRec.MimeType)
	if err != nil {
		http.Error(w, "Image not found", 404)
		return
	}

	var out image.Image
	switch fit {
	case "crop":
		out = imgpkg.ThumbnailFill(src, targetW, targetH)
	default:
		out = imgpkg.ThumbnailFit(src, targetW, targetH)
	}

	data, err := imgpkg.EncodeJPEGBytes(out)
	if err != nil {
		http.Error(w, "Failed to encode image", 500)
		return
	}

	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(cachePath, data, 0644)

	serveJPEG(w, data)
}

func loadSourceImage(appRoot, filename, mimeType string) (image.Image, error) {
	f, err := os.Open(filepath.Join(appRoot, "storage", "source", filename))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return imgpkg.DecodeAny(f, mimeType)
}

func serveJPEG(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(data)
}
