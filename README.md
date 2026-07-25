# VAPEU - Terminal API Client

An offline-first, keyboard-driven terminal API client for testing and debugging HTTP APIs.

## Overview

VAPEU is a cross-platform TUI application built with Go. Designed as a fast, lightweight alternative to Postman and Insomnia, it operates locally and over SSH using a keyboard-first workflow.

## Features

- **Keyboard-Driven TUI**: Full-width dual-panel interface with customizable themes.
- **Telescope Request Finder (`Alt+Space`)**: Real-time fuzzy filtering of saved and historical HTTP requests by name, URL, or method.
- **HTTP Methods & Formats**: Support for GET, POST, PUT, PATCH, DELETE, HEAD, and OPTIONS with headers, query parameters, cookies, and payloads (JSON, XML, Form URL-Encoded, Multipart, and Binary).
- **Interactive Authentication (`Alt+3`)**: Live credentials editor supporting Bearer Tokens, Basic Authentication, and API Keys (`Alt+A`).
- **Dynamic Variable Resolver**: Support for environment variables (`{{BASE_URL}}`, `{{TOKEN}}`) and built-in functions (`{{uuid}}`, `{{timestamp}}`, `{{randomInt}}`, `{{randomString}}`, `{{hostname}}`).
- **Response Metrics & Syntax Highlighting**: Color-coded JSON/XML rendering via Chroma, HTTP status badges, timing metrics (ms), and payload size indicators.
- **Request Management & History**: Save named requests, organize open requests in tabs, and view automatically logged request history (`Ctrl+H`).
- **Import / Export Capabilities**: Import cURL commands, OpenAPI/Swagger specifications, and Postman collections (`Alt+I`).
- **HTTP/2 Transport & Redirect Handling**: Multi-streamed connection pooling with automatic 307/308 redirect resolution.

## Tech Stack

- **Language**: Go 1.22+
- **TUI Framework**: [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling & Layout**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **TUI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Syntax Highlighting**: [Chroma](https://github.com/alecthomas/chroma)
- **CLI Engine**: [Cobra](https://github.com/spf13/cobra)

## Project Structure

```
vapeu/
├── .github/
│   └── workflows/
│       ├── ci.yml          # Continuous Integration workflow
│       └── release.yml     # Multi-platform binary release workflow
├── cmd/
│   └── apicli/
│       └── main.go         # CLI entrypoint
├── internal/
│   ├── api/                # HTTP client and transport engine
│   ├── impexp/             # cURL, OpenAPI, and Postman parsers
│   ├── models/             # Core domain models
│   ├── storage/            # Local configuration and disk storage manager
│   ├── theme/              # Design system tokens and color themes
│   ├── ui/                 # Bubble Tea TUI components, layouts, and modals
│   └── variables/          # Template variable resolver engine
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.22 or higher installed on your system.

### Installation

1. Clone the repository:
```bash
git clone https://github.com/teyorkk/vapeu.git
cd vapeu
```

2. Build the executable binary:
```bash
go build -o apicli ./cmd/apicli
```

3. Launch the interactive terminal UI:
```bash
./apicli
```

### CLI Commands

#### Execute cURL Command Directly:
```bash
./apicli run "curl -X POST https://jsonplaceholder.typicode.com/posts -d '{\"title\":\"test\"}'"
```

#### Import Specification File:
```bash
./apicli import path/to/openapi.yaml
```

## Keyboard Shortcuts Reference

### General & Workspace
| Key Binding | Action |
| --- | --- |
| `Alt+Space` / `Ctrl+F` | Open Telescope Request Finder |
| `Alt+Left` / `Alt+Right` (`Alt+[` / `Alt+]`) | Switch Open Request Tabs |
| `Ctrl+N` / `Ctrl+T` | Create New Request Tab |
| `Ctrl+W` | Close Active Request Tab |
| `Tab` / `Shift+Tab` | Toggle Focus between Request Editor and Response Viewer |
| `Ctrl+1` / `Ctrl+2` | Direct Focus (1: Request Editor, 2: Response Viewer) |
| `Alt+K` | Cycle UI Color Theme (Dark, Light, Cyberpunk, Nord, High Contrast) |
| `?` | Display Help & Keyboard Reference Modal |
| `Ctrl+Q` | Exit Application |

### Request & Execution
| Key Binding | Action |
| --- | --- |
| `Ctrl+R` / `Ctrl+Enter` | Send HTTP Request |
| `Ctrl+X` | Cancel / Abort Active Request |
| `Ctrl+S` | Save Request (Prompts for Name) |
| `Alt+I` / `Ctrl+I` | Open Import Modal (cURL / OpenAPI / Postman) |
| `Ctrl+H` | Open Request History Modal |

### Editor & Sub-Tabs
| Key Binding | Action |
| --- | --- |
| `Alt+1` | Select Query Parameters Sub-Tab |
| `Alt+2` | Select Headers Sub-Tab |
| `Alt+3` | Select Auth Sub-Tab (`Alt+A` cycles None / Bearer / Basic / APIKey) |
| `Alt+4` | Select Request Body Sub-Tab |
| `Alt+5` | Select Cookies Sub-Tab |
| `ESC` / `Alt+U` | Focus URL Input Field (from sub-tab text areas) |

### Response Viewer
| Key Binding | Action |
| --- | --- |
| `Alt+B` | Select Response Body View |
| `Alt+H` | Select Response Headers View |
| `Alt+C` | Select Response Cookies View |
| `Alt+P` | Toggle Pretty / Raw JSON Formatting |

## Configuration

Application data and configuration files are stored locally:
- Configuration: `~/.apicli/config.yaml`
- Collections: `~/.apicli/collections/`
- Environments: `~/.apicli/environments/`
- History Log: `~/.apicli/history.json`
