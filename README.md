# 🤫 shushing face

speak naturally, get polished text.

shushing face is a desktop speech-to-text app that records your voice, transcribes it, and rewrites it into clean text using AI — then pastes it right where your cursor is.

## how it works

1. press your shortcut (or click the mic)
2. speak
3. press again — refined text appears in your focused app

no clipboard pollution. the text is typed directly into whatever you're working in.

## features

- **refinement profiles** — casual, professional, concise, or create your own
- **auto-paste** — types refined text into the focused app via `wtype` (wayland) or `xdotool` (x11)
- **panel indicator** — mic icon in your system bar, changes when recording
- **desktop notifications** — optional recording/processing alerts
- **history** — browse past transcriptions with collapsible raw transcripts
- **test playground** — try your prompts against sample text before recording
- **welcome wizard** — guided first-time setup
- **dark/light/system theme** — with a warm yellow accent
- **single instance** — second launch brings existing window to front
- **IPC toggle** — `shushingface --toggle` from any script or shortcut

## install

requires [go 1.26+](https://go.dev), [wails v2](https://wails.io), [bun](https://bun.sh), and on linux: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`.

```bash
git clone https://codeberg.org/dbus/shushingface
cd shushingface
just install
```

this builds the app, installs to `~/.local/bin/`, sets up the desktop entry and icon, and registers `Super+Ctrl+B` as a shortcut (COSMIC/GNOME).

for auto-paste, install `wtype` (wayland) or `xdotool` (x11):

```bash
# debian/ubuntu/pop_os
sudo apt install wtype

# fedora
sudo dnf install wtype

# arch
sudo pacman -S wtype
```

## develop

```bash
just dev          # wails dev with hot reload
just build        # production build
just lint         # go + biome linting
just format       # go + biome formatting
```

## commands

```bash
shushingface              # launch the app
shushingface --toggle     # toggle recording (for keyboard shortcuts)
shushingface --show       # bring window to front
shushingface --quit       # fully exit
```

## architecture

```
internal/
├── ai/            # processor interface + groq implementation
├── audio/         # recorder interface + malgo implementation + wav encoder
├── config/        # settings, refinement profiles, migration
├── core/          # engine (record → transcribe → refine)
├── history/       # sqlite transcription history
├── indicator/     # panel icon via StatusNotifierItem (linux) 
├── ipc/           # unix socket single-instance + CLI commands
├── notify/        # desktop notifications via D-Bus (linux)
├── osutil/        # active window detection (stub)
├── paste/         # auto-type via wtype/xdotool (linux)
├── platform/      # OS, display server, desktop, package manager detection
└── ui/desktop/    # wails bridge (app lifecycle, settings, recording)

frontend/          # react 19 + vite + tailwind v4 + shadcn/ui
├── components/    # sidebar, views, wizard, dialogs
└── lib/           # hooks, utils
```

every platform-specific package uses build tags with matching function signatures across `_linux.go` / `_other.go` / `_windows.go` files. the core pipeline (record → transcribe → refine) is fully cross-platform.

## configuration

stored at `~/.config/shushingface/config.json`. includes API keys, refinement profiles, preferences.

history stored at `~/.config/shushingface/history.db` (sqlite).

## license

MIT
