package main

import (
	"fmt"
	"log"

	"helm-visualizer/internal/adapters/http"
	"helm-visualizer/internal/adapters/repository"
	"helm-visualizer/internal/core/services"
	"helm-visualizer/pkg/config"
)

func main() {
	fmt.Printf("=== CHART PAPER BACKEND STARTING ===\n")
	fmt.Printf("Version: OCI-only (no directory scanning)\n")

	cfg := config.Load()

	repo, err := repository.NewSQLiteRepository(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	chartService := services.NewChartService(repo)
	registryService := services.NewRegistryService(repo)

	server := http.NewServer(chartService, registryService)

	fmt.Printf("=== SERVER STARTING ON :%s ===\n", cfg.Port)
	fmt.Printf("NO DIRECTORY SCANNING - OCI REGISTRY ONLY\n")
	
	if err := server.Start(cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
