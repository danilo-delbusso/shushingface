package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"codeberg.org/dbus/sussurro/internal/ai/groq"
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

	// Extract API key for the default Groq provider fallback, or use env
	apiKey := os.Getenv("GROQ_API_KEY")
	if provider, ok := cfg.Providers[cfg.RefinementProviderID]; ok && provider.APIKey != "" {
		apiKey = provider.APIKey
	}

	if apiKey == "" {
		log.Fatal("Error: Groq API key is missing. Please set GROQ_API_KEY or configure it in the desktop app.")
	}

	processor, err := groq.NewProcessor(apiKey)
	if err != nil {
		log.Fatal("Error initializing processor: ", err)
	}

	engine := core.NewEngine(recorder, processor)

	if err := tui.Start(engine); err != nil {
		log.Fatal("Error starting TUI: ", err)
	}
}
