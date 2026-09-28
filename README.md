# SynX — Next-Gen Decentralized LAN File Transfer Mesh

> **SynX** is a cross-platform desktop application for fast, direct device-to-device file sharing over a local network. Built with a focus on simplicity, speed, privacy, and seamless transfers without relying on cloud storage.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Next.js Version](https://img.shields.io/badge/Next.js-16.3-000000?style=flat&logo=next.js)](https://nextjs.org)
[![React Version](https://img.shields.io/badge/React-19.2-61DAFB?style=flat&logo=react)](https://react.dev)
[![Wails Version](https://img.shields.io/badge/Wails-v2.15-DF1A5B?style=flat)](https://wails.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Overview

**SynX** is a cross-platform desktop application for fast, direct device-to-device file sharing over a local network. Built with a focus on simplicity, speed, privacy, and seamless transfers without relying on cloud storage.

The current desktop build provides the foundation for **real peer-to-peer file transfers**, with live transfer progress and resumable downloads planned as the next milestone. There are no third-party servers, no telemetry, no bandwidth caps, and no cloud middlemen. 

Whether transferring multi-gigabyte disk images, video footage, archives, or source code directories, SynX saturates your local wire speed with chunked parallel streaming, streaming SHA-256 integrity verification, and automatic failure recovery.

---

## Key Features

- **⚡ Zero Cloud Dependency**: Transfers occur strictly within your local area network (LAN) over high-speed sockets. Full offline capability.
- **🛰️ Interactive Sonar Radar**: Dynamic 360° rotating radar scanner on an HTML5 canvas plotting active LAN nodes, signal rings, and dynamic peer blips in real time.
- **📦 8MB Chunk-Based Streaming**: Large files are split into 8MB chunks, transferred concurrently with rolling 5s throughput sampling, and committed atomically.
- **🛡️ Strict Chroot Path Containment**: Comprehensive security layer blocking directory traversal (`../`, `..\`), UNC paths, drive letters, and null bytes. Remote uploads never escape your configured folder.
- **🔒 Cryptographic Trust & 6-Digit PIN Pairing**: Pair devices securely using a time-limited 6-digit numeric PIN or Ed25519 token authentication.
- **📡 Automatic UDP Discovery**: Auto-discovers nearby SynX nodes announcing over UDP broadcast (port 8788) every 2.5s.
- **💾 Pure-Go SQLite Persistence**: Uses pure-Go SQLite (`modernc.org/sqlite`) in WAL mode for crash recovery, transfer logs, device trust, and settings.
- **🎨 High-Contrast Cyberpunk UI**: Sleek OLED Void interface with electric cyan and hyper violet neon accents, laser-swept holographic drop zone, and smooth micro-animations.
- **🌐 Dual-Mode Desktop & Web Access**: Operates as a native desktop application via Wails v2 or headlessly via the CLI daemon, accessible from standard web browsers on `:8787`.

---

## Architecture

```
┌────────────────────────────────────────────────────────┐
│                   Next.js 16 UI                        │
│   (React 19, TypeScript, App Router, Canvas Radar)     │
└──────────────┬─────────────────────────┬───────────────┘
               │ (Wails Bindings)        │ (HTTP REST / API)
┌──────────────▼─────────────────────────▼───────────────┐
│                    SynX Core (Go)                      │
├──────────────────┬──────────────────┬──────────────────┤
│  Peer Discovery  │ Transfer Engine  │ Security & Trust │
│  (UDP Port 8788) │ (8MB Chunk Pool) │ (PIN / Ed25519)  │
├──────────────────┼──────────────────┼──────────────────┤
│ Safe File System │ Streaming SHA256 │ SQLite Database  │
│ (Path Sandbox)   │ (1MB Buffers)    │ (modernc WAL)    │
└──────────────────┴──────────────────┴──────────────────┘
```

---

## Getting Started

### Prerequisites

- **Go**: 1.25 or higher
- **Node.js**: 20.x or 24.x and npm
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Running in Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kunal-live/SynX.git
   cd SynX
   ```

2. **Build the Next.js frontend**:
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

3. **Launch the Wails dev environment**:
   ```bash
   wails dev
   ```

   The desktop shell will open automatically, and the application will be accessible in web browsers at:
   - Wails DevServer: `http://localhost:34115`
   - SynX LAN HTTP Server: `http://localhost:8787`

### Building for Production

Compile the standalone, zero-dependency Windows desktop binary:

```bash
# Build desktop executable (embeds Next.js frontend)
wails build

# Or compile the CLI daemon:
go build -o build/bin/synx-cli.exe ./cmd/synx
```

The resulting executable is located in `build/bin/synx.exe`.

---

## Security Model

- **Safe Path Sanitization**: Path inputs from remote peers are strictly checked. Absolute prefixes, drive letters, parent traversals, and non-canonical paths are rejected before filesystem access.
- **Part-File Commits**: Incoming transfers stream to temporary `.synx.part` files. Only upon a verified SHA-256 checksum match is the file committed to its destination.
- **Authentication**: Non-loopback REST endpoints require an `X-SynX-Token` session token or valid pairing PIN.

---

## Privacy & Open-Source Policies

SynX is committed to digital sovereignty, zero external tracking, and full user data transparency:
- [Privacy Policy](PRIVACY.md) — Complete disclosure of local-first data architecture, zero telemetry, and local storage.
- [Terms & Conditions](TERMS.md) — MIT licensing, lawful network usage terms, and disclaimers.
- [Cookie & Local Storage Policy](COOKIE_POLICY.md) — Zero HTTP/tracking cookies; local client-side storage transparency.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
