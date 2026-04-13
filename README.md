<p align="center">
  <img src="build/icon.svg" width="80" height="80" alt="shushing face">
</p>

<h1 align="center">shushing face</h1>

<p align="center">speak naturally, get polished text.</p>

<p align="center">
shushing face records your voice, transcribes it with AI, rewrites it into clean text, and types it right where your cursor is.
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

macOS and Windows support is planned.

## Install

### From .deb (Pop!_OS / Ubuntu / Debian)

Download the latest `.deb` from [releases](https://codeberg.org/dbus/shushingface/releases):

```bash
sudo dpkg -i shushingface_*.deb
```

### From tarball

```bash
tar xzf shushingface-*.tar.gz
sudo cp shushingface /usr/local/bin/
```

### From source

Requires Go 1.26+, Bun, and `libwebkit2gtk-4.1-dev`:

```bash
just install
```

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
