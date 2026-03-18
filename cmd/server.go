package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"asistente/boot"
	"asistente/config"
	"asistente/internal/hooks"
	"asistente/internal/skills"
	"asistente/pkg/service"
	"asistente/pkg/usecase"
)

type App struct {
	memorySvc service.MemoryService
	scheduler *usecase.Scheduler
	server    boot.Gin
}

func NewApp(cfg config.Config) *App {
	cl := NewClients(cfg)
	memorySvc := NewMemoryService(cfg)
	hooksRegistry := hooks.NewRegistry()

	scheduler := NewScheduler(cl, cfg, memorySvc, hooksRegistry)

	ctrls := NewControllers(cl, cfg, memorySvc,
		NewFinanceService(cl.Sheets, cfg.SheetsSheetName),
		NewEmbedder(cl.AILight),
		skills.NewCachedLoader(skills.NewLoader(cfg.SkillsDir)),
		hooksRegistry,
		scheduler,
	)

	scheduler.Start()

	return &App{
		memorySvc: memorySvc,
		scheduler: scheduler,
		server:    boot.NewGin(middlewareMapper(cfg.WebhookSecret), setupRoutes(ctrls)),
	}
}

func (a *App) Run() {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %s, shutting down...", sig)
		if err := a.server.Shutdown(); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := a.server.Run(); err != nil {
		log.Printf("server stopped: %v", err)
	}
}

func (a *App) Close() {
	a.scheduler.Stop()
	a.memorySvc.Close()
}
