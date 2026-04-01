package main

import (
	"log"

	"jarvis/clients"
	"jarvis/config"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

func NewMemoryService(cfg config.Config) service.MemoryService {
	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}
	store, err := service.NewPGMemoryService(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to init postgres store: %v", err)
	}
	log.Println("Storage backend: PostgreSQL")
	return store
}

func NewFinanceService(sheetsClient *clients.SheetsClient, sheetName string) service.FinanceService {
	if sheetsClient == nil {
		return nil
	}
	return service.NewSheetsFinanceService(sheetsClient, sheetName)
}

func NewEmbedder(ai domain.AIProvider) service.Embedder {
	inner := service.NewAIEmbedder(ai)
	return service.NewCachedEmbedder(inner, 500)
}
