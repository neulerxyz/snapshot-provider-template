package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"snapshot_service/config"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "./", "Path to config directory")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logfile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logfile.Close()
	logrus.SetOutput(logfile)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	r := mux.NewRouter()
	r.HandleFunc("/", listFilesHandler(cfg))
	r.PathPrefix("/snapshots/").Handler(http.StripPrefix("/snapshots/", http.FileServer(http.Dir(cfg.SnapshotDir))))
	log.Printf("Starting server on :%d...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerPort), r))
}

func listFilesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		dirs := []string{
			filepath.Join(cfg.SnapshotDir, cfg.GethClientName),
			filepath.Join(cfg.SnapshotDir, cfg.CosmosClientName),
		}

		for _, dir := range dirs {
			files, err := os.ReadDir(dir)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading directory %s: %v", dir, err), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "<h2>%s</h2>\n", filepath.Base(dir))

			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(filepath.Base(dir), file.Name())
					fmt.Fprintf(w, `<a href="/snapshots/%s">%s</a><br>`, filePath, file.Name())
				}
			}
		}
	}
}
