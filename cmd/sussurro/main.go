package main

import (
	"log"

	"codeberg.org/dbus/sussurro/internal/ai/groq"
	"codeberg.org/dbus/sussurro/internal/audio/malgo"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/ui/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	recorder, err := malgo.NewRecorder(16000)
	if err != nil {
		log.Fatal("Error initializing recorder: ", err)
	}
	defer recorder.Close()

	processor, err := groq.NewProcessor(cfg.GroqAPIKey)
	if err != nil {
		log.Fatal("Error initializing processor: ", err)
	}

	engine := core.NewEngine(recorder, processor)

	if err := tui.Start(engine); err != nil {
		log.Fatal("Error starting TUI: ", err)
	}
}
