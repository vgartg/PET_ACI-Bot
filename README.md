# ACI-Bot

A Telegram bot that controls a Windows host machine over chat. Authorised users can launch and close common applications, open URLs, search YouTube and shut down the host

![CI](https://github.com/vgartg/aci-bot/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/vgartg/aci-bot)](https://goreportcard.com/report/github.com/vgartg/aci-bot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)


## Features

- Key-based access. Users authenticate with a shared secret before any command is executed
- Application launcher for File Explorer, Google Chrome, Steam, Discord and Faceit (with optional Anti-Cheat companion)
- One-shot URL opener and a free-form URL prompt that validates input as `http`/`https`
- YouTube search by free-text query
- Process control via `taskkill` for Chrome, Steam, Discord and Faceit
- Host-only shutdown command guarded by a Telegram chat-id check
- Session manager with per-user state and clean logout

## Project layout

```
.
├── cmd/aci-bot           entry point
├── internal/bot          Telegram client wiring
├── internal/config       environment-based configuration
├── internal/executor     OS-level operations (process control, URL open, shutdown)
├── internal/handler      command router and business logic
└── internal/session      thread-safe in-memory session state
```

The codebase follows the standard Go project layout: only `cmd/` is importable from outside, everything else lives under `internal/`

## Requirements

- Go 1.22 or newer
- Windows host for runtime (the bot uses `taskkill`, `rundll32` and `shutdown`)
- A Telegram bot token from [@BotFather](https://t.me/BotFather)

The repository builds and tests on Linux as well, so CI runs the whole pipeline on `ubuntu-latest`

## Configuration

All configuration is read from environment variables

| Variable | Required | Description |
| --- | --- | --- |
| `ACI_BOT_TOKEN` | yes | Telegram bot API token |
| `ACI_BOT_KEY` | yes | Shared secret users must send to unlock the bot |
| `ACI_HOST_ID` | yes | Telegram chat id that is allowed to issue `/turn_off_pc` |
| `ACI_PATH_CHROME` | no | Path to `chrome.exe` |
| `ACI_PATH_STEAM` | no | Path to `steam.exe` |
| `ACI_PATH_DISCORD` | no | Path to Discord launcher |
| `ACI_PATH_FACEIT` | no | Path to Faceit launcher |
| `ACI_PATH_FACEIT_AC` | no | Path to Faceit Anti-Cheat |
| `ACI_STICKER_DISABLED` | no | Set to a truthy value to disable the sticker easter egg |
| `ACI_STICKER_URL` | no | `fmt`-style URL pattern with a single `%s` for the sticker index |

Example PowerShell session:

```powershell
$env:ACI_BOT_TOKEN  = "123:abc"
$env:ACI_BOT_KEY    = "open-sesame"
$env:ACI_HOST_ID    = "100200300"
$env:ACI_PATH_STEAM = "C:\Program Files (x86)\Steam\steam.exe"
go run ./cmd/aci-bot
```

## Build

```bash
make build
```

The build target cross-compiles a Windows `amd64` binary to `bin/aci-bot.exe` with stripped symbols. To run from source against your local Go toolchain use `go run ./cmd/aci-bot` or `make run`

## Commands

After sending the configured key the bot exposes:

```
/help                 list all commands
/stop                 end the current session
/open_google          open Google Chrome
/open_youtube         open https://youtube.com
/open_video_by_name   search YouTube by query
/open_ya_mus          open Yandex Music
/open_vk              open VKontakte
/open_url             open an arbitrary http/https URL
/close_google         close Google Chrome
/open_steam           open Steam
/close_steam          close Steam
/open_faceit          open Faceit (and Faceit Anti-Cheat if configured)
/close_faceit         close Faceit and its Anti-Cheat
/open_discord         open Discord
/close_discord        close Discord
/open_explorer        open File Explorer
/turn_off_pc          shut down the host machine (host chat only)
```

## Testing

```bash
go test ./...
go test -race ./...
make cover            generate coverage.html
```

The handler is decoupled from Telegram and the operating system through small interfaces (`handler.Sender`, `executor.Executor`), which makes the business logic fully unit-testable without network or process side effects

## Linting

```bash
make lint
```

The repository ships a `.golangci.yml` that enables `govet`, `staticcheck`, `revive`, `errcheck`, `ineffassign`, `misspell`, `unused`, `unconvert`, `gofmt` and `goimports`. The same configuration runs in CI

## Continuous integration

The [`ci` workflow](.github/workflows/ci.yml) runs on every push and pull request to `main`:

1. `go vet` and the full test suite with the race detector
2. `golangci-lint` against the shared configuration
3. Cross-platform `linux/amd64` and `windows/amd64` builds, uploaded as workflow artefacts

## Security notes

- The shared key is the only authentication factor, so deploy the bot only on hosts you fully control and never commit `.env` files
- `/turn_off_pc` is gated by `ACI_HOST_ID`; double-check the chat id before exposing the bot
- The free-form URL prompt rejects anything that is not `http` or `https` to avoid arbitrary protocol handlers
