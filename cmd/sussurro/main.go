package main

import (
	"log"

	"github.com/joho/godotenv"

	"codeberg.org/dbus/sussurro/internal/ai/factory"
	"codeberg.org/dbus/sussurro/internal/audio/malgo"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/ui/tui"
)

func main() {
	// We still load .env for backward compatibility in TUI
	// while the desktop app will strictly use the UI to manage config.json.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	recorder, err := malgo.NewRecorder(16000)
	if err != nil {
		log.Fatal("Error initializing recorder: ", err)
	}
	defer recorder.Close()

	// The AI Factory builds our transcriber & refiner implementations based on the JSON Config
	processor, err := factory.NewFromConfig(cfg)
	if err != nil {
		log.Fatal("Error initializing AI factory: ", err)
	}

	engine := core.NewEngine(recorder, processor)

	if err := tui.Start(engine); err != nil {
		log.Fatal("Error starting TUI: ", err)
	}
}
