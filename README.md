# Only4BMS Server

> The multiplayer backend for [Only4BMS](https://github.com/minwook-shin/Only4BMS). Host your own rhythm battles — no central server required.

## What is this?

Only4BMS Server is a lightweight, self-hosted multiplayer server for the Only4BMS rhythm game. Think Minecraft servers, but for 4-key BMS battles.

Spin up a server, share the address with friends, and play together. It's that simple.

## ✨ Features

- **Real-time 1v1 (and beyond)**: WebSocket-based score sync with configurable player count
- **Daily Courses**: Procedurally generated daily challenges with leaderboards (daily / weekly / monthly)
- **Song Hosting**: Serve BMS files directly — clients download what they need
- **Zero Dependencies Runtime**: Single binary, SQLite for storage, no external services needed
- **ARM-Ready**: Cross-compiles for Oracle Cloud Free Tier (or any ARM box you have lying around)

## 🚀 Quick Start

### Prerequisites

- Go 1.23+ installed
- BMS song files to serve

### Build & Run

```bash
go build -o only4bms-server .
./only4bms-server
```

Server starts on `http://localhost:8080` by default.

### Adding Songs

Drop your BMS songs into the `songs/` directory:

```text
songs/
├── my_song/
│   ├── metadata.json    # {"id", "title", "artist", "level", "bpm"}
│   ├── chart.bms
│   ├── kick.wav
│   ├── snare.wav
│   └── bg.jpg
└── another_song/
    └── ...
```

## ⚙️ Configuration

All settings can be passed as environment variables or CLI flags:

| Env Variable | CLI Flag | Default | Description |
|---|---|---|---|
| `PORT` | `--port` | `8080` | Server port |
| `SONGS_DIR` | `--songs-dir` | `./songs` | Path to song data |
| `DB_PATH` | `--db-path` | `./rankings.db` | SQLite database path |
| `SERVER_PASSWORD` | `--password` | *(none)* | Lobby password (optional) |
| `MAX_PLAYERS` | `--max-players` | `2` | Max players per lobby (min 2) |
| `DAILY_RESET_HOUR` | `--daily-reset-hour` | `0` | Daily course reset hour (UTC) |

Example with custom settings:

```bash
MAX_PLAYERS=4 SERVER_PASSWORD=secret ./only4bms-server --port 9090
```

## 🌐 API Overview

### HTTP Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/songs` | List all available songs |
| `GET` | `/api/songs/{id}` | Get file manifest for a song |
| `GET` | `/api/songs/{id}/download/{filename}` | Download a song file |
| `GET` | `/api/daily/today` | Today's daily course (seed + difficulty) |
| `POST` | `/api/daily/submit` | Submit a daily course score |
| `GET` | `/api/daily/ranking?period=daily\|weekly\|monthly` | Leaderboard |
| `GET` | `/api/daily/history?days=7` | Recent daily courses |
| `GET` | `/health` | Health check |

### WebSocket

Connect to `ws://host:port/ws?name=YourName` for real-time multiplayer.

See [MULTIPLAYER_API.md](documents/MULTIPLAYER_API.md) for the full protocol spec.

## 🏗️ Project Structure

```
only4bms-server/
├── main.go              # Entrypoint
├── config.go            # Env vars / CLI flags
├── internal/
│   ├── ws/              # WebSocket hub, client, message types
│   ├── lobby/           # Lobby state machine, manager
│   ├── api/             # HTTP handlers (songs, daily)
│   ├── daily/           # Daily course seed, ranking
│   └── db/              # SQLite init, queries
└── songs/               # Song data directory
```

## 🛠️ Development

```bash
# Run tests
go test ./...

# Build for ARM64 (Oracle Cloud, Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o only4bms-server-arm64 .
```

## 📋 Roadmap

- [x] **Phase 1**: Core server — HTTP + WebSocket + lobby + score relay
- [x] **Phase 2**: Daily courses — SQLite + seed generation + ranking
- [ ] **Phase 3**: Deployment — Docker + ARM cross-compile
- [ ] **Phase 4**: Discord OAuth2 authentication

## 🤝 Contributing

PRs and issues welcome! Whether it's bug fixes, new features, or just telling us your server setup — we'd love to hear from you.

## 📜 License

MIT License
