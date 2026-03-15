# 🤖 Standup Bot

> A privacy-first developer standup assistant. Fetches your GitHub commits, summarizes them with a local LLM, and lets you chat with your daily summary — all without sending a single byte of your data to an external service.

Built in 3 hours at Cursor Workshop Istanbul by [Rüştü Kağan Kıvanç](https://github.com/rkkivanc) & [Yakup Kahraman](https://github.com/yakupkahraman).

---

## ✨ Features

- **Automatic Standup Generation** — Fetches your last 24 hours of GitHub commits and structures them into Yesterday / Today / Blockers format
- **Local LLM Inference** — Uses Ollama / LM Studio or any Local LLM running on your machine. No OpenAI, no Anthropic, no cloud
- **Chat with your Standup** — Ask follow-up questions about your daily work via a streaming chat interface
- **Multi-provider Support** — Auto-detects Local LLMs running on your machine
- **Copy-ready Output** — Export your standup as plain text or Markdown in one click
- **Privacy by Design** — GitHub token is used in-memory per request only, never logged or persisted

---

## 🖼️ Screenshots

### Local AI Model Selection
> Auto-detects running local providers (Ollama, LM Studio, LocalAI) and lets you connect with one click.
 
<div align="center">
  <img src="https://github.com/user-attachments/assets/abf9ccf4-59f1-4941-9995-0d27f017e1e0" width="900" />
</div>
 
---
 
### Standup Dashboard & Chat
> Commits automatically categorized into Yesterday / Today / Blockers. Ask follow-up questions via the local LLM-powered chat panel.
 
<div align="center">
  <img src="https://github.com/user-attachments/assets/91d3f917-141f-4dd6-a1f9-861c32831ca2" width="900" />
</div>

> _Enter your repo, generate your standup, ask questions — all locally._

---

## 🏗️ Architecture

```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│                 │  HTTP   │                 │  HTTP   │                 │
│  Next.js 16     │────────▶│   Go Backend    │────────▶│  Local LLM      │
│  (Frontend)     │◀────────│   (Port 8080)   │◀────────│                 │
│  Port 3000      │   JSON  │                 │   SSE   │                 │
└─────────────────┘         └────────┬────────┘         └─────────────────┘
                                     │ HTTPS
                                     ▼
                            ┌─────────────────┐
                            │   GitHub API    │
                            │  (commits only) │
                            └─────────────────┘
```

**Stack:**
- **Frontend:** Next.js 16, React 19, Tailwind CSS 4, TypeScript
- **Backend:** Go (net/http, no framework)
- **LLM:** Ollama · LM Studio · LocalAI (auto-detected)
- **Infra:** Docker · Docker Compose

---

## 🚀 Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- A GitHub Personal Access Token ([create one here](https://github.com/settings/personal-access-tokens)) with `repo` scope
- A local LLM running on your machine (see [LLM Setup](#-llm-setup) below)

### 1. Clone the repo

```bash
git clone https://github.com/rkkivanc/standup-bot.git
cd standup-bot
git checkout main
```

### 2. Run with Docker Compose

```bash
docker compose up --build
```

Or without Docker, use the convenience script:

```bash
chmod +x run.sh
./run.sh
```

### 4. Open the app

Navigate to [http://localhost:3000](http://localhost:3000).

---

## 🧠 LLM Setup

Standup Bot auto-detects any of these local AI providers on startup:

| Provider | Default Port | Notes |
|----------|-------------|-------|
| **Ollama** | `11434` | `ollama pull gemma3:1b` to get started |
| **LM Studio** | `1234` | Enable local server in the app |
| **LocalAI** | `8081` | Self-hosted, runs on consumer hardware |

Once a provider is running, open the **Local AI** modal in the app, click **Connect**, and you're ready.

---

## 🗂️ Project Structure

```
cursor-workshop/
├── backend/
│   ├── cmd/
│   │   └── main.go                     # Entry point
│   └── internal/
│       ├── controllers/
│       │   ├── chat_controller.go       # POST /api/chat
│       │   ├── commits_controller.go    # POST /api/commits
│       │   ├── llm_discovery_controller.go  # LLM provider endpoints
│       │   └── standup_controller.go    # POST /api/standup
│       ├── routes/
│       │   └── routes.go               # Route registration
│       └── services/
│           ├── github_service.go        # GitHub API client
│           ├── llm_discovery_service.go # Local LLM auto-detection
│           └── standup_service.go       # LLM + keyword-based summarizer
├── frontend/
│   ├── app/
│   │   ├── layout.tsx
│   │   └── page.tsx                    # Main dashboard
│   └── components/
│       └── LocalAISelector.tsx          # Local AI modal
├── docker-compose.yml
├── run.sh                              # Local dev convenience script
```
---

## 🔒 Privacy

This project was built with privacy as a first-class concern:

- **GitHub token** is passed through in-memory only, never written to any database or log
- **Commit data** is processed per-request and discarded immediately after
- **LLM inference** runs entirely on your local machine — no data leaves your network
- **No telemetry**, no analytics, no tracking of any kind

---

## 🏆 Origin

Built at **Cursor Workshop Istanbul** organized by [Gürkan Fikret Günak](https://www.linkedin.com/in/gurkanfikretgunak/) (Cursor Ambassador Turkey).

~100 developers, 40+ projects, 3 hours. This was one of them.

---

## 🤝 Contributing

PRs are welcome. For major changes, open an issue first to discuss what you'd like to change.

```bash
# Backend
cd backend && go run ./cmd/main.go

# Frontend
cd frontend && npm install && npm run dev
```
