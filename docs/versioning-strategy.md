# Versioning, Migration & Release Strategy

## Overview

This document covers the complete versioning, migration, auto-update, and release strategy for shushingface. It is designed to be implemented incrementally — each section is independent and can be built in any order.

---

## 1. Version Numbers

### Format: SemVer (MAJOR.MINOR.PATCH)

- Start at `v0.1.0` (pre-1.0 = "config may change between minors")
- Move to `v1.0.0` when core features are stable and config format is committed to
- Pre-release: `v0.2.0-rc.1`, `v0.2.0-beta.1`

### Single Source of Truth: Git Tags

No version constant in source code. The git tag is canonical. Everything derives from it at build time.

**Flow:**

```
git tag v0.1.0
  → CI reads the tag
  → passes to Go via -ldflags
  → binary knows its version at runtime
  → frontend reads it via Wails binding
```

### Implementation

**`internal/version/version.go`:**

```go
package version

// Set at build time via -ldflags
// "-X codeberg.org/dbus/shushingface/internal/version.version=v0.1.0"
var version = "dev"

func Version() string { return version }
```

**Build command:**

```bash
VERSION=$(git describe --tags --always --dirty)
wails build -tags webkit2_41 \
  -ldflags "-X codeberg.org/dbus/shushingface/internal/version.version=$VERSION"
```

`git describe` produces:
- `v0.1.0` — exactly on a tag
- `v0.1.0-3-gabcdef` — 3 commits after the tag
- `v0.1.0-3-gabcdef-dirty` — uncommitted changes

**Expose to frontend:**

```go
func (a *App) GetVersion() string {
    return version.Version()
}
```

Show in sidebar footer, About section, or settings.

**justfile integration:**

```just
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

build:
    wails build -tags webkit2_41 \
      -ldflags "-X codeberg.org/dbus/shushingface/internal/version.version={{VERSION}}"
```

---

## 2. Config Migration System

### Problem

The current config has no version field. Migrations are ad-hoc `if` blocks in `Load()` that detect the *shape* of the data. This works with 3 migrations but becomes fragile as the config evolves.

### Design

Add a `configVersion` integer to the JSON. Migrations are ordered functions that transform raw JSON from version N to N+1.

**Key principles:**
- Migrations operate on `map[string]any` (raw JSON), NOT the Settings struct
- This decouples old migrations from the current struct shape
- Forward-only — no "down" migrations (desktop apps don't roll back configs)
- If `configVersion > currentConfigVersion`, refuse to load (user should update the app)

**`internal/config/migrate.go`:**

```go
package config

import "fmt"

type migration struct {
    version     int
    description string
    up          func(data map[string]any) error
}

var migrations = []migration{
    {
        version:     1,
        description: "multi-provider map to single provider fields",
        up: func(data map[string]any) error {
            // ... port existing migration logic ...
            return nil
        },
    },
    {
        version:     2,
        description: "single provider to connections slice",
        up: func(data map[string]any) error {
            // ...
            return nil
        },
    },
    // ... more migrations ...
}

const currentConfigVersion = 2

func migrateConfig(data map[string]any) (bool, error) {
    v, _ := data["configVersion"].(float64) // JSON numbers are float64
    current := int(v)

    if current > currentConfigVersion {
        return false, fmt.Errorf(
            "config version %d is newer than this app supports (%d); please update",
            current, currentConfigVersion,
        )
    }

    if current == currentConfigVersion {
        return false, nil
    }

    for _, m := range migrations {
        if m.version > current {
            if err := m.up(data); err != nil {
                return false, fmt.Errorf("migration v%d (%s): %w", m.version, m.description, err)
            }
            data["configVersion"] = float64(m.version)
        }
    }

    return true, nil
}
```

**Revised `Load()`:**

```go
func Load() (*Settings, error) {
    data, err := os.ReadFile(configFile)
    // ...

    var raw map[string]any
    json.Unmarshal(data, &raw)

    migrated, err := migrateConfig(raw)
    if err != nil {
        return nil, err
    }

    // Re-marshal migrated map into the struct
    data, _ = json.Marshal(raw)
    var settings Settings
    json.Unmarshal(data, &settings)

    applyDefaults(&settings)

    if migrated {
        Save(&settings)
    }
    return &settings, nil
}
```

**Bootstrapping existing users:** Users with no `configVersion` field are treated as version 0. All migrations run sequentially.

**Adding a new migration:**

1. Append to the `migrations` slice with the next version number
2. Increment `currentConfigVersion`
3. Write a test with a synthetic `map[string]any` input
4. Update the `Settings` struct if new fields are needed

**Testing:**

```go
func TestMigration_V1(t *testing.T) {
    data := map[string]any{
        "providers": map[string]any{
            "groq-default": map[string]any{
                "name": "groq", "apiKey": "gsk_test",
            },
        },
        "transcriptionProviderId": "groq-default",
    }
    migrated, err := migrateConfig(data)
    assert(migrated == true)
    assert(err == nil)
    assert(data["configVersion"] == float64(1))
    // ... assert the migrated shape ...
}
```

**Benefits over current approach:**
- Legacy fields can be removed from the Settings struct once their migration exists
- This also cleans up the generated TypeScript bindings
- Each migration is independently testable
- Clear error messages on failure ("migration v3 (add connections) failed: ...")
- Future-version protection prevents data corruption on downgrade

---

## 3. Database Migration System

### Problem

The current SQLite setup uses `CREATE TABLE IF NOT EXISTS` + a bare `ALTER TABLE` that silently fails if the column exists. This doesn't track which migrations have run.

### Design

A `schema_migrations` table tracks applied versions. Each migration runs in its own transaction.

**`internal/history/migrate.go`:**

```go
package history

import (
    "database/sql"
    "fmt"
)

type dbMigration struct {
    version     int
    description string
    up          func(tx *sql.Tx) error
}

var dbMigrations = []dbMigration{
    {
        version:     1,
        description: "create transcriptions table",
        up: func(tx *sql.Tx) error {
            _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS transcriptions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                raw_transcript TEXT,
                refined_message TEXT,
                active_app TEXT
            )`)
            return err
        },
    },
    {
        version:     2,
        description: "add error column",
        up: func(tx *sql.Tx) error {
            _, err := tx.Exec(`ALTER TABLE transcriptions ADD COLUMN error TEXT DEFAULT ''`)
            return err
        },
    },
}

func runMigrations(db *sql.DB) error {
    db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version INTEGER PRIMARY KEY,
        applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)

    var current int
    db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current)

    for _, m := range dbMigrations {
        if m.version <= current {
            continue
        }
        tx, err := db.Begin()
        if err != nil {
            return err
        }
        if err := m.up(tx); err != nil {
            tx.Rollback()
            return fmt.Errorf("migration %d (%s): %w", m.version, m.description, err)
        }
        tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, m.version)
        if err := tx.Commit(); err != nil {
            return err
        }
    }
    return nil
}
```

**Bootstrapping existing databases:**

Users upgrading from before the migration system have the table + error column but no `schema_migrations`. Detect this and seed the tracking table:

```go
func bootstrapExistingDB(db *sql.DB) {
    // If schema_migrations exists, already bootstrapped
    var name string
    if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&name) == nil {
        return
    }
    // If transcriptions table exists, seed migrations as already applied
    if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='transcriptions'`).Scan(&name) != nil {
        return // Fresh database
    }
    // Seed based on what columns exist
    db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
    db.Exec(`INSERT INTO schema_migrations (version) VALUES (1)`)
    // Check for error column via PRAGMA
    rows, _ := db.Query(`PRAGMA table_info(transcriptions)`)
    defer rows.Close()
    for rows.Next() {
        var cid int; var n, t string; var nn int; var d sql.NullString; var pk int
        rows.Scan(&cid, &n, &t, &nn, &d, &pk)
        if n == "error" {
            db.Exec(`INSERT INTO schema_migrations (version) VALUES (2)`)
        }
    }
}
```

**Adding a new migration:**

1. Append to `dbMigrations` with the next version number
2. Write the SQL in the `up` function
3. Test with `:memory:` SQLite database

---

## 4. Auto-Update System

### Phase 1: Check + Notify (implement now)

Simple, non-invasive, works on every platform.

**`internal/update/check.go`:**

```go
package update

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type Release struct {
    TagName string `json:"tag_name"`
    HTMLURL string `json:"html_url"`
    Body    string `json:"body"`
}

func Check(ctx context.Context, currentVersion string) (*Release, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET",
        "https://codeberg.org/api/v1/repos/dbus/shushingface/releases/latest", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var rel Release
    json.NewDecoder(resp.Body).Decode(&rel)

    if isNewer(currentVersion, rel.TagName) {
        return &rel, nil
    }
    return nil, nil
}
```

**In `app.Startup()`:**

```go
go func() {
    time.Sleep(5 * time.Second)
    rel, err := update.Check(ctx, version.Version())
    if err == nil && rel != nil {
        wailsRuntime.EventsEmit(ctx, "update-available", rel)
    }
}()
```

**Frontend:** listen for `"update-available"` event, show a toast with a link to the release page.

**User preference:** add `CheckForUpdates bool` to Settings (default true). Respect it before calling `Check()`.

### Phase 2: In-Place Binary Replacement (future)

The pattern used by VS Code, Obsidian, etc. on Linux:
1. Download new binary to temp file
2. Verify checksum against a signed manifest
3. Rename current binary to `.old`
4. Move new binary into place
5. Prompt for restart

Libraries: [minio/selfupdate](https://github.com/minio/selfupdate) handles the atomic replacement. For signing, use Ed25519 — include the public key in the binary, sign releases in CI.

Only implement this if users request it. Check + notify covers 90% of the value.

### Platform-Specific Notes

| Platform | Auto-update approach |
|----------|---------------------|
| Linux (AppImage) | Check + notify → link to download. AppImage supports delta updates via `appimageupdatetool` (future) |
| Linux (Flatpak) | Handled by the Flatpak runtime — `flatpak update` |
| macOS | Sparkle framework is the standard. Or check + notify to start |
| Windows | WinSparkle, or NSIS silent installer. Or check + notify to start |

---

## 5. Release Packaging

### Linux (primary)

**AppImage (recommended primary format):**
- Single file, runs on any distro
- `wails build` → wrap in AppImage with `linuxdeploy`/`appimagetool`
- Include desktop file + icon in the AppImage

**Tarball (always available):**
- `shushingface-v0.1.0-linux-amd64.tar.gz`
- Contains: binary, `.desktop` file, icon
- Users extract and run or use `just install`

**Flatpak (later, post-1.0):**
- Flathub submission for discoverability
- Manifest + runtime dependencies + sandboxing
- Significant one-time setup, low ongoing maintenance

**.deb/.rpm (probably never):**
- Requires hosting a package repo for updates
- Not worth the maintenance for a small project

### macOS (future)

`wails build` → `.app` bundle → wrap in `.dmg` with `create-dmg`. Code signing requires Apple Developer account ($99/year).

### Windows (future)

`wails build -nsis` → NSIS installer. Cross-compile from Linux with MinGW, or build on Windows CI runner.

---

## 6. CI/CD Pipeline

### Current: `.forgejo/workflows/ci.yml`

Runs on every push/PR: Go build + vet + test, frontend typecheck + lint + build. This stays as-is.

### New: `.forgejo/workflows/release.yml`

Triggers on version tags only:

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }

      - uses: actions/setup-go@v6
        with: { go-version: '1.26' }

      - name: Install system deps
        run: sudo apt-get update && sudo apt-get install -y libwebkit2gtk-4.1-dev libgtk-3-dev

      - name: Install frontend deps
        run: cd frontend && npm ci

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          wails build -tags webkit2_41 \
            -ldflags "-X codeberg.org/dbus/shushingface/internal/version.version=$VERSION"

      - name: Package
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          mkdir -p dist
          tar czf "dist/shushingface-${VERSION}-linux-amd64.tar.gz" \
            -C build/bin shushingface

      - name: Generate changelog
        run: |
          PREV=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
          if [ -n "$PREV" ]; then
            git log --pretty=format:"- %s" "${PREV}..HEAD" > CHANGELOG.md
          else
            git log --pretty=format:"- %s" HEAD > CHANGELOG.md
          fi

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          body_path: CHANGELOG.md
          files: dist/*
```

### Release Process

```bash
# 1. Make sure main is clean and CI passes
# 2. Tag the release
git tag v0.1.0
git push origin v0.1.0

# 3. CI builds, packages, and creates a Codeberg release with:
#    - shushingface-v0.1.0-linux-amd64.tar.gz
#    - Auto-generated changelog from commit messages
```

### Build Matrix (future)

When cross-platform support is added:

```yaml
strategy:
  matrix:
    include:
      - os: ubuntu-latest
        platform: linux/amd64
        artifact: shushingface-linux-amd64
      - os: macos-latest
        platform: darwin/amd64
        artifact: shushingface-darwin-amd64
      - os: macos-latest
        platform: darwin/arm64
        artifact: shushingface-darwin-arm64
      - os: windows-latest
        platform: windows/amd64
        artifact: shushingface-windows-amd64.exe
```

Codeberg doesn't offer macOS/Windows runners. Options:
- Self-hosted runners
- Build macOS/Windows locally and upload to the release manually
- Use a separate CI service for non-Linux builds

---

## 7. Implementation Order

Each step is independently shippable:

| Step | What | Depends on | Effort |
|------|------|-----------|--------|
| 1 | `internal/version/version.go` + ldflags in justfile | Nothing | 30min |
| 2 | `GetVersion()` Wails method + frontend display | Step 1 | 30min |
| 3 | `internal/config/migrate.go` + tests | Nothing | 2-3hr |
| 4 | Refactor `config.Load()` to use migration system | Step 3 | 1hr |
| 5 | `internal/history/migrate.go` + bootstrap + tests | Nothing | 2hr |
| 6 | Refactor `history.NewManager()` to use migrations | Step 5 | 30min |
| 7 | `internal/update/check.go` + startup check + frontend toast | Step 1 | 1-2hr |
| 8 | `.forgejo/workflows/release.yml` | Step 1 | 1hr |
| 9 | Tag `v0.1.0` | Steps 1-8 | 5min |

Steps 1-2 and 3-6 can be done in parallel. Step 7 can start as soon as step 1 is done.

---

## 8. Config Migration Testing Checklist

For every release that changes the config format:

- [ ] Write the migration function in `migrate.go`
- [ ] Increment `currentConfigVersion`
- [ ] Add test: version N input → version N+1 output
- [ ] Add test: version 0 (no configVersion) → current version (full chain)
- [ ] Add test: future version → error
- [ ] Manual test: copy a real config from the previous version, run the new binary, verify settings are preserved
- [ ] Document the change in the release changelog

---

## 9. Database Migration Testing Checklist

For every release that changes the schema:

- [ ] Write the migration in `dbMigrations`
- [ ] Add test: fresh `:memory:` database → all migrations run, schema correct
- [ ] Add test: pre-existing database at version N → new migration runs, data preserved
- [ ] Manual test: copy a real `history.db` from the previous version, run the new binary, verify history is intact
- [ ] If the migration is destructive (dropping columns, changing types), add a backup step before the migration
