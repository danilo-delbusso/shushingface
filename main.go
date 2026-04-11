package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"codeberg.org/dbus/sussurro/internal/ai/factory"
	"codeberg.org/dbus/sussurro/internal/audio/malgo"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/history"
	"codeberg.org/dbus/sussurro/internal/ui/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// 1. Initialize History (SQLite)
	var hist *history.Manager
	if cfg.EnableHistory {
		hist, err = history.NewManager()
		if err != nil {
			log.Printf("Warning: failed to initialize history manager: %v", err)
		} else {
			defer hist.Close()
		}
	}

	// 2. Initialize Audio Recorder
	recorder, err := malgo.NewRecorder(16000)
	if err != nil {
		log.Fatal("Error initializing recorder: ", err)
	}
	defer recorder.Close()

	// 3. Initialize AI Processor (via Dynamic Factory)
	processor, err := factory.NewFromConfig(cfg)
	if err != nil {
		log.Fatal("Error initializing AI factory: ", err)
	}

	// 4. Instantiate Core Orchestration Engine
	engine := core.NewEngine(recorder, processor)

	// 5. Build Desktop Bridge Context
	app := desktop.NewApp(engine, cfg, hist)

	// 6. Set up File Logger
	logPath, _ := config.GetLogPath()
	appLogger := logger.NewFileLogger(logPath)

	// 7. Run Wails Application
	err = wails.Run(&options.App{
		Title:  "Sussurro",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
		Logger:            appLogger,
		LogLevel:          logger.INFO,
		HideWindowOnClose: true, // Key component of the "Always-On" background behavior
	})

	if err != nil {
		log.Fatal("Error starting Wails application: ", err)
	}
}
