package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	appdb "imagemanager/internal/db"
	"imagemanager/internal/server"
)

func main() {
	port := flag.String("port", "8080", "HTTP port")
	dataDir := flag.String("data-dir", "", "App data directory (default: executable directory)")
	flag.Parse()

	appRoot := *dataDir
	if appRoot == "" {
		appRoot = resolveAppRoot()
	}

	db, err := appdb.Open(appRoot)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	mux := server.New(db, appRoot, webFS)

	addr := ":" + *port
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		time.Sleep(200 * time.Millisecond)
		openBrowser("http://localhost:" + *port)
	}()

	log.Printf("ImageManager running at http://localhost:%s", *port)
	log.Printf("Data directory: %s", appRoot)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down...")
}

func resolveAppRoot() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	// During `go run`, executable is in a temp go-build path
	if strings.Contains(exe, "go-build") || strings.Contains(exe, "go\\tmp") {
		return "."
	}
	return filepath.Dir(exe)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

