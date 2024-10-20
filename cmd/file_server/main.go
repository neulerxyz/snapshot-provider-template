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
	logrus.Printf("Starting server on :%d...", cfg.ServerPort)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerPort), r))
}

func listFilesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		clients := []string{cfg.GethClientName, cfg.CosmosClientName}

		// Loop over both clients (geth/cosmos)
		for _, client := range clients {
			clientDir := filepath.Join(cfg.SnapshotDir, client)

			// Read files from client directory
			files, err := os.ReadDir(clientDir)
			if err != nil {
				fmt.Fprintf(w, "Error reading directory %s: %v<br>", clientDir, err)
				continue
			}

			// Display client name (Geth or Cosmos)
			fmt.Fprintf(w, "<h2>%s</h2>\n", client)

			// List all files in the directory
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(client, file.Name())
					fmt.Fprintf(w, `<a href="/snapshots/%s">%s</a><br>`, filePath, file.Name())
				}
			}

			// Add the latest snapshot link (e.g., geth_pruned_latest.tar.lz4)
			latestFile := fmt.Sprintf("%s_%s_latest.tar.lz4", client, cfg.GethSnapshotType)
			fmt.Fprintf(w, `<a href="/snapshots/%s">Latest %s Snapshot (%s)</a><br>`, latestFile, client, cfg.GethSnapshotType)
		}
	}
}
