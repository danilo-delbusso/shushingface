<p align="center">
  <img src="build/icon.svg" width="80" height="80" alt="shushing face">
</p>

<h1 align="center">shushing face</h1>

<p align="center">speak naturally, get polished text.</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-in%20development-orange" alt="in development">
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/license-AGPL--3.0-blue" alt="AGPL-3.0"></a>
</p>

> ⚠️ Early-stage project. Expect rough edges, breaking changes, and missing features. Not recommended for production use yet.

<p align="center">
shushing face helps you work more efficiently by leveraging the fact that speaking is generally faster than typing. It operates in the background, allowing you to speak freely, and with a simple press or shortcut, it transcribes and cleans up your spoken words into clear, professional text, making it ideal for emails, reports, and other workplace communication.
</p>

<p align="center">
  <img src="images/home.png" width="720" alt="main window">
</p>

## Supported Platforms

| OS | Version | Desktop | Status |
|----|---------|---------|--------|
| Pop!_OS | 24.04 LTS | COSMIC (Wayland) | Tested |
| Ubuntu | 24.04 LTS | GNOME (Wayland/X11) | Expected to work |
| Ubuntu | 22.04 LTS | GNOME (X11/Wayland) | Expected to work |
| Fedora | 40+ | GNOME (Wayland) | Expected to work |
| Arch Linux | Rolling | Any | Expected to work |
| Windows | 10 / 11 | — | In development |

macOS support is planned.

## Install

### From a release artifact

Download the latest build from [releases](https://codeberg.org/dbus/shushingface/releases):

- **Pop!_OS / Ubuntu / Debian**: `sudo dpkg -i shushingface_*.deb`
- **Other Linux**: extract `shushingface-*.tar.gz` and copy `shushingface` onto your `PATH`
- **Windows**: run the NSIS installer

### From source

```bash
just doctor      # report missing dependencies
just bootstrap   # install missing dependencies (use --yes to skip prompts)
just dev         # run in dev mode
just install     # build + install for the current user
```

The same three commands work on Linux and Windows. `just install` puts
the binary under `$HOME/.local/bin/` (`%USERPROFILE%\.local\bin\` on
Windows). Override with `PREFIX=/path/to/wherever just install`.

### Packaging (maintainers)

`just package` produces installable artifacts for the current OS:

- Linux: `dist/shushingface-<version>-linux-amd64.tar.gz` and `dist/shushingface_<version>_amd64.deb`
- Windows: NSIS installer in `dist/`

## Usage

1. Launch shushing face
2. Set up an AI provider (Groq is free and fast)
3. Bind `shushingface --toggle` to a keyboard shortcut
4. Press the shortcut to start recording, press again to stop
5. Refined text is typed where your cursor is

## Screenshots

**Onboarding** — connect a provider, pick a style, bind a shortcut.

<table>
  <tr>
    <td width="33%"><img src="images/chooseapiprovider.png" alt="connect a provider"></td>
    <td width="33%"><img src="images/choosestyle.png" alt="choose a style"></td>
    <td width="33%"><img src="images/shortcut.png" alt="bind a shortcut"></td>
  </tr>
</table>

**Configure** — pick default models for transcription and refinement, then tweak prompts per style.

<table>
  <tr>
    <td width="50%"><img src="images/choosemodels.png" alt="default models"></td>
    <td width="50%"><img src="images/editstylesandprompts.png" alt="edit styles and prompts"></td>
  </tr>
</table>

**Use** — speak, get polished text wherever your cursor is.

<p align="center">
  <img src="images/exampletranscription.png" width="720" alt="example transcription">
</p>

## License

[AGPL-3.0](LICENSE.md)
