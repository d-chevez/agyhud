# agyhud

> Modern, ultra-fast HUD and statusline engine for the Google Antigravity CLI (`agy`), built in Go with an interactive configuration TUI.

[![Go Version](https://img.shields.io/github/go-mod/go-version/d-chevez/agyhud)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Overview

`agyhud` is a dedicated Heads-Up Display (HUD) and statusline for Google Antigravity CLI (`agy`). It surfaces critical session telemetry—active model, agent lifecycle state, context window consumption, quota countdowns, and Git workspace status—at a glance directly inside your terminal with sub-10ms render latency.

### Highlights (In Active Development)

- ⚡ **Zero-Lag Engine:** Single compiled Go binary with sub-10ms render latency on every agent cycle.
- 🎨 **Interactive TUI:** Built-in terminal UI (powered by Bubbletea) for real-time live preview, row layout adjustments, and granular color customizations.
- 🧩 **Modular Row Layout:** Flexible 2-line HUD or single-line views that automatically adapt to terminal width.
- 🚀 **One-Click Automated Setup:** Safe, zero-friction integration with Antigravity settings (`agyhud install` / `agyhud uninstall`).
- 🌲 **Smart Git Telemetry:** Direct `.git/HEAD` inspection with ephemeral TTL caching to prevent disk thrashing.
- 🔤 **Font Flexibility:** First-class support for both Nerd Fonts and classic Unicode/ASCII terminals.

---

## Roadmap

- [ ] **v0.1.0-alpha:** Core engine, JSON payload parser, modular Lipgloss widget pipeline, and CLI `render` command.
- [ ] **v0.2.0:** Safe integration lifecycle manager (`install` / `uninstall`).
- [ ] **v0.3.0:** Interactive configuration TUI with live preview and color editor.
- [ ] **v1.0.0:** Production release with GoReleaser matrix, GitHub Actions CI/CD, and one-liner installer scripts.

---

## License

This project is licensed under the MIT License.
