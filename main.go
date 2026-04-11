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

	"codeberg.org/dbus/shushingface/internal/ai/factory"
	"codeberg.org/dbus/shushingface/internal/audio/malgo"
	"codeberg.org/dbus/shushingface/internal/config"
	"codeberg.org/dbus/shushingface/internal/core"
	"codeberg.org/dbus/shushingface/internal/history"
	"codeberg.org/dbus/shushingface/internal/ipc"
	"codeberg.org/dbus/shushingface/internal/ui/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// CLI commands — talk to running instance and exit
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--toggle":
			if err := ipc.SendToggle(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "--show":
			if err := ipc.SendShow(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "--quit":
			if err := ipc.Send("QUIT"); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	// Single instance: if already running, bring to front and exit
	if ipc.IsRunning() {
		slog.Info("already running, showing existing window")
		ipc.SendShow()
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
		Title:  "shushing face",
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
