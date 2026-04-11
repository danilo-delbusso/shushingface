package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"codeberg.org/dbus/shushingface/internal/ai/factory"
	"codeberg.org/dbus/shushingface/internal/audio/malgo"
	"codeberg.org/dbus/shushingface/internal/config"
	"codeberg.org/dbus/shushingface/internal/core"
	"codeberg.org/dbus/shushingface/internal/ui/tui"
)

func main() {
	closeLog := config.InitLogger()
	defer closeLog()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	recorder, err := malgo.NewRecorder(16000)
	if err != nil {
		slog.Error("failed to initialize recorder", "error", err)
		os.Exit(1)
	}
	defer recorder.Close()

	processor, err := factory.NewFromConfig(cfg)
	if err != nil {
		slog.Error("failed to initialize AI factory", "error", err)
		os.Exit(1)
	}

	prompt := cfg.SystemPrompt
	if prompt == "" {
		prompt = config.DefaultSystemPrompt
	}
	engine := core.NewEngine(recorder, processor, prompt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := tui.Start(engine, ctx); err != nil {
		slog.Error("TUI error", "error", err)
		os.Exit(1)
	}
}
