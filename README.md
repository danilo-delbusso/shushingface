# 🤫 shushing face

speak naturally, get polished text.

shushing face records your voice, transcribes it with AI, rewrites it into clean text, and types it right where your cursor is. no clipboard pollution.

## features

- **refinement profiles** — casual, professional, concise, or your own custom styles
- **auto-paste** — types refined text directly into the focused app
- **panel indicator** — mic icon in your system bar, changes when recording
- **test playground** — try prompts against sample text before recording
- **history** — browse past transcriptions
- **single instance** — `shushingface --toggle` to record from anywhere
- **dark/light/system theme**

## architecture

```
internal/
├── ai/            # processor interface + groq implementation
├── audio/         # recorder interface + malgo + wav encoder
├── config/        # settings, refinement profiles, migration
├── core/          # engine (record → transcribe → refine)
├── history/       # sqlite history
├── indicator/     # panel icon via D-Bus StatusNotifierItem (linux)
├── ipc/           # unix socket single-instance + CLI commands
├── notify/        # desktop notifications (linux)
├── paste/         # auto-type via wtype/xdotool (linux)
├── platform/      # OS, display server, desktop, package manager detection
└── ui/desktop/    # wails bridge

frontend/          # react 19 + vite + tailwind v4 + shadcn/ui
```

platform-specific code uses build tags. the core pipeline is cross-platform. macOS and Windows stubs are in place.

## license

MIT
