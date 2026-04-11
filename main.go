package main

import (
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"codeberg.org/dbus/sussurro/internal/ai/factory"
	"codeberg.org/dbus/sussurro/internal/audio/malgo"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/history"
	"codeberg.org/dbus/sussurro/internal/ipc"
	"codeberg.org/dbus/sussurro/internal/ui/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// CLI mode: send toggle signal to running instance and exit
	if len(os.Args) > 1 && os.Args[1] == "--toggle" {
		if err := ipc.SendToggle(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	hist, err := history.NewManager()
	if err != nil {
		slog.Warn("failed to initialize history", "error", err)
	} else {
		defer hist.Close()
	}

	recorder, err := malgo.NewRecorder(16000)
	if err != nil {
		slog.Error("failed to initialize recorder", "error", err)
		return
	}
	defer recorder.Close()

	processor, err := factory.NewFromConfig(cfg)
	if err != nil {
		slog.Error("failed to initialize AI factory", "error", err)
		return
	}

	prompt := cfg.SystemPrompt
	if prompt == "" {
		prompt = config.DefaultSystemPrompt
	}
	engine := core.NewEngine(recorder, processor, prompt)
	app := desktop.NewApp(engine, cfg, hist)

	logPath, _ := config.GetLogPath()
	appLogger := logger.NewFileLogger(logPath)

	err = wails.Run(&options.App{
		Title:  "Sussurro",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:        app.Startup,
		OnShutdown:        app.Shutdown,
		Bind:              []interface{}{app},
		Logger:            appLogger,
		LogLevel:          logger.INFO,
		HideWindowOnClose: true,
	})

	if err != nil {
		slog.Error("wails application error", "error", err)
	}
}
