# Platform Compatibility

## Supported

| OS | Version | Desktop | Display | Architecture | Status |
|----|---------|---------|---------|-------------|--------|
| Pop!_OS | 24.04 LTS | COSMIC | Wayland | x86_64 | Tested |

## Expected to Work (untested)

| OS | Version | Desktop | Display | Architecture | Notes |
|----|---------|---------|---------|-------------|-------|
| Pop!_OS | 24.04 LTS | COSMIC | Wayland | x86_64 | Primary target |
| Pop!_OS | 22.04 LTS | GNOME | X11/Wayland | x86_64 | Older Pop, GNOME-based |
| Ubuntu | 24.04 LTS | GNOME | Wayland | x86_64 | Same base as Pop!_OS |
| Ubuntu | 22.04 LTS | GNOME | X11/Wayland | x86_64 | Needs webkit2gtk-4.1 |
| Fedora | 40+ | GNOME | Wayland | x86_64 | webkit2gtk available |
| Arch Linux | Rolling | Any | Any | x86_64 | webkit2gtk available |

## Not Yet Supported

| OS | Notes |
|----|-------|
| macOS | Wails supports it, needs testing + packaging (.dmg) |
| Windows | Wails supports it, needs testing + packaging (NSIS) |
| Linux ARM64 | Cross-compilation needed, no runner available |

## Runtime Dependencies

The app requires `libwebkit2gtk-4.1-0` at runtime. This is preinstalled on:
- Pop!_OS 24.04 (COSMIC)
- Ubuntu 24.04 (GNOME)
- Most GNOME-based distros

For distros without it preinstalled, the .deb package declares the dependency.
