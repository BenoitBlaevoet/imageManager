package server

import (
	"database/sql"
	"html/template"
	"imagemanager/internal/config"
	"imagemanager/internal/handler"
	"imagemanager/internal/repo"
	"io/fs"
	"net/http"
)

func New(db *sql.DB, appRoot string, webFS fs.FS) *http.ServeMux {
	app := &handler.App{DB: db, AppRoot: appRoot}
	mux := http.NewServeMux()

	// Static assets
	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("GET /assets/", http.FileServerFS(sub))

	// Pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveFile(w, sub, "index.html", "text/html")
	})

	mux.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		serveEditor(w, r, app, sub)
	})

	// API routes
	mux.HandleFunc("/api/upload", app.Upload)
	mux.HandleFunc("/api/images", app.Images)
	mux.HandleFunc("/api/crops", app.Crops)
	mux.HandleFunc("/api/generate", app.Generate)
	mux.HandleFunc("/api/image-serve", app.ImageServe)
	mux.HandleFunc("/api/components", app.Components)
	mux.HandleFunc("/api/component-images", app.ComponentImages)
	mux.HandleFunc("/api/presets", app.Presets)
	mux.HandleFunc("/api/tags", app.Tags)
	mux.HandleFunc("/api/config", app.Config)
	mux.HandleFunc("/api/open-folder", app.OpenFolder)

	return mux
}

func serveFile(w http.ResponseWriter, fsys fs.FS, name, contentType string) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

type editorData struct {
	Title       string
	CompLabel   string
	OriginalName string
	Width       int
	Height      int
	ImageID     string
	Filename    string
	ComponentID string
}

func serveEditor(w http.ResponseWriter, r *http.Request, app *handler.App, fsys fs.FS) {
	imageID := r.URL.Query().Get("image_id")
	if imageID == "" {
		imageID = r.URL.Query().Get("id")
	}
	componentID := r.URL.Query().Get("component_id")

	imgRepo := repo.NewImageRepo(app.DB)
	image, err := imgRepo.FindByID(imageID)
	if err != nil || image == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	compLabel := ""
	if componentID != "" {
		components, _ := config.LoadComponents(app.AppRoot)
		comp := config.FindComponent(components, componentID)
		if comp != nil {
			compLabel = comp.Label
		}
	}

	title := "Edit"
	if compLabel != "" {
		title += " — " + compLabel
	}
	title += " — " + image.OriginalName

	tmplData, err := fs.ReadFile(fsys, "editor.html")
	if err != nil {
		http.Error(w, "Template not found", 500)
		return
	}

	tmpl, err := template.New("editor").Parse(string(tmplData))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, editorData{
		Title:        title,
		CompLabel:    compLabel,
		OriginalName: image.OriginalName,
		Width:        image.Width,
		Height:       image.Height,
		ImageID:      imageID,
		Filename:     image.Filename,
		ComponentID:  componentID,
	})
}
