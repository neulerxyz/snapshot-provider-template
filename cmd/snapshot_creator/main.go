package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"snapshot_service/config"
	"snapshot_service/internal/node"
	"snapshot_service/internal/snapshot"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "./config", "Path to config directory")
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

	runSnapshotProcess(cfg)
	ticker := time.NewTicker(cfg.SnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runSnapshotProcess(cfg)
		}
	}
}

func runSnapshotProcess(cfg *config.Config) {
	logrus.Info("Starting snapshot process...")
	gethNode := node.NewGethNode(cfg.GethRPCURL)
	cosmosNode := node.NewCosmosNode(cfg.CosmosRPCURL)

	gethSynced, err := gethNode.IsSynced()
	if err != nil {
		logrus.Errorf("Error checking Geth sync status: %v", err)
		return
	}

	cosmosSynced, err := cosmosNode.IsSynced()
	if err != nil {
		logrus.Errorf("Error checking Cosmos sync status: %v", err)
		return
	}

	if gethSynced && cosmosSynced {
		logrus.Info("Both nodes are synced. Proceeding with snapshot.")
		logrus.Info("Stopping Geth and Cosmos nodes...")
		if err := snapshot.StopService(cfg.GethServiceName); err != nil {
			logrus.Errorf("Error stopping Geth service: %v", err)
			return
		}
		if err := snapshot.StopService(cfg.CosmosServiceName); err != nil {
			logrus.Errorf("Error stopping Cosmos service: %v", err)
			return
		}

		snapshot.Sleep(10)

		gethSnapshotDir := filepath.Join(cfg.SnapshotDir, cfg.GethClientName)
		cosmosSnapshotDir := filepath.Join(cfg.SnapshotDir, cfg.CosmosClientName)
		err = os.MkdirAll(gethSnapshotDir, os.ModePerm)
		if err != nil {
			logrus.Errorf("Error creating geth snapshot directory: %v", err)
			return
		}
		err = os.MkdirAll(cosmosSnapshotDir, os.ModePerm)
		if err != nil {
			logrus.Errorf("Error creating cosmos snapshot directory: %v", err)
			return
		}

		timestamp := time.Now().Format("200601021504")
		gethSnapshotFile := filepath.Join(gethSnapshotDir, fmt.Sprintf("%s_%s_%s.tar.lz4", cfg.GethClientName, cfg.GethSnapshotType, timestamp))
		cosmosSnapshotFile := filepath.Join(cosmosSnapshotDir, fmt.Sprintf("%s_%s_%s.tar.lz4", cfg.CosmosClientName, cfg.CosmosSnapshotType, timestamp))

		logrus.Infof("Creating %s snapshot...", cfg.GethClientName)
		if err := snapshot.CompressDirectory(cfg.GethDataDir, gethSnapshotFile); err != nil {
			logrus.Errorf("Error compressing %s data directory: %v", cfg.GethClientName, err)
		} else {
			logrus.Infof("%s snapshot created at %s", cfg.GethClientName, gethSnapshotFile)
		}

		logrus.Infof("Creating %s snapshot...", cfg.CosmosClientName)
		if err := snapshot.CompressDirectory(cfg.CosmosDataDir, cosmosSnapshotFile); err != nil {
			logrus.Errorf("Error compressing %s data directory: %v", cfg.CosmosClientName, err)
		} else {
			logrus.Infof("%s snapshot created at %s", cfg.CosmosClientName, cosmosSnapshotFile)
		}

		logrus.Infof("Updating latest %s snapshot...", cfg.GethClientName)
		// Update latest link for Geth snapshot using the config value for client name
		err := updateLatestSymlink(cfg.SnapshotDir, cfg.GethClientName, gethSnapshotFile, cfg.GethSnapshotType)
		if err != nil {
			log.Printf("Error updating symlink for Geth: %v", err)
		}

		logrus.Infof("Updating latest %s snapshot...", cfg.CosmosClientName)
		// Update latest link for Cosmos snapshot using the config value for client name
		err = updateLatestSymlink(cfg.SnapshotDir, cfg.CosmosClientName, cosmosSnapshotFile, cfg.CosmosSnapshotType)
		if err != nil {
			log.Printf("Error updating symlink for Cosmos: %v", err)
		}

		logrus.Infof("Restarting %s and %s nodes...", cfg.GethClientName, cfg.CosmosClientName)
		if err := snapshot.StartService(cfg.GethServiceName); err != nil {
			logrus.Errorf("Error starting %s service: %v", cfg.GethClientName, err)
		}
		if err := snapshot.StartService(cfg.CosmosServiceName); err != nil {
			logrus.Errorf("Error starting %s service: %v", cfg.CosmosClientName, err)
		}

		// Retain only the latest snapshots
		retainSnapshots(gethSnapshotDir, fmt.Sprintf("%s_%s_", cfg.GethClientName, cfg.GethSnapshotType), 2)
		retainSnapshots(cosmosSnapshotDir, fmt.Sprintf("%s_%s_", cfg.CosmosClientName, cfg.CosmosSnapshotType), 2)

		logrus.Info("Snapshot process completed successfully.")
	} else {
		logrus.Info("One or both nodes are not synced. Snapshot process aborted.")
	}
}

func retainSnapshots(snapshotDir, prefix string, maxSnapshots int) {
	// Find all snapshot files with the given prefix
	files, err := filepath.Glob(filepath.Join(snapshotDir, prefix+"*.tar.lz4"))
	if err != nil {
		logrus.Errorf("Error listing snapshots with prefix %s: %v", prefix, err)
		return
	}

	// Sort the files by modification time (most recent first)
	sort.Slice(files, func(i, j int) bool {
		fi1, _ := os.Stat(files[i])
		fi2, _ := os.Stat(files[j])
		return fi1.ModTime().After(fi2.ModTime())
	})

	// If there are more snapshots than the allowed max, delete the oldest ones
	if len(files) > maxSnapshots {
		for _, file := range files[maxSnapshots:] {
			logrus.Infof("Deleting old snapshot: %s", file)
			if err := os.Remove(file); err != nil {
				logrus.Errorf("Error deleting snapshot %s: %v", file, err)
			}
		}
	}
}

// Function to update the symlink for the latest snapshot
func updateLatestSymlink(snapshotDir, clientName, snapshotFileName, snapshotType string) error {
	// Create the symlink name (e.g., geth_pruned_latest.tar.lz4)
	latestSymlink := filepath.Join(snapshotDir, fmt.Sprintf("%s_%s_latest.tar.lz4", clientName, snapshotType))
	// Remove the existing symlink if it exists
	err := os.Remove(latestSymlink)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create a new symlink pointing to the latest snapshot file
	relativeSnapshotFile, err := filepath.Rel(snapshotDir, snapshotFileName) // Get relative path
	if err != nil {
		return err
	}

	err = os.Symlink(relativeSnapshotFile, latestSymlink)
	if err != nil {
		return err
	}

	log.Printf("Updated symlink: %s -> %s", latestSymlink, relativeSnapshotFile)
	return nil
}
