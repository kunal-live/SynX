# SynX — System Architecture

**Version:** 1.0  
**Focus:** Developer-First Local Device Connectivity Platform  
**Target:** Cross-platform (Windows, macOS, Linux)  
**Primary Stack:** Go + Wails v2 + Next.js 16  

---

## 1. Architectural Overview

SynX connects developer workstations, laptops, and servers across a local network (Wi-Fi / Ethernet) without requiring central cloud infrastructure.

Rather than being a consumer file drop tool, SynX provides a **developer device connectivity substrate** where capabilities (interactive terminal, remote command execution, clipboard synchronization, chunked file transfer, and automation events) run on top of an authenticated, capability-negotiated peer mesh.

```text
                         SYNX PLATFORM
                               |
                    +----------+----------+
                    |      SynX Core      |
                    +----------+----------+
                               |
           +-------------------+-------------------+
           |                   |                   |
      Discovery             Identity            Security
           |                   |                   |
           +-------------------+-------------------+
                               |
                        Connection Layer
                               |
                         Protocol Layer
                               |
                      Capability Registry
                               |
         +----------+----------+----------+----------+
         |          |          |          |          |
      Terminal   Commands    Files    Clipboard    Events
         |          |          |          |          |
         +----------+----------+----------+----------+
                               |
                         Developer API
                               |
                        Automation Mesh
```

---

## 2. Core Subsystems

### 2.1 Identity Service (`internal/identity`)
- **Device ID:** Cryptographically random identifier generated on first launch (`synx_<hex>`), persisted in `identity.json`.
- **Key Pair:** Ed25519 256-bit asymmetric key pair for peer authentication and message signing.
- **Node Metadata:** Device name, platform (OS/Arch), version, and capability list.

### 2.2 Discovery Layer (`internal/discovery`)
- **Transport:** UDP Multicast & Broadcast on port `8788`.
- **Announcement Interval:** 2.5 seconds.
- **Privacy:** Discovery broadcasts strictly non-sensitive identity and capability advertisements.

### 2.3 Peer Manager (`internal/peer`)
- **Lifecycle States:** `DISCOVERED` → `AVAILABLE` → `CONNECTING` → `CONNECTED` → `DISCONNECTED`.
- **Heartbeat & Expiration:** Peers not seen for >10s automatically transition to `OFFLINE`.
- **Trust Records:** Paired device public keys persisted in local SQLite WAL database.

### 2.4 Connection & Transport (`internal/network`)
- **Interface Abstraction:** Pluggable `Connection` interface (`Send`, `Receive`, `Close`).
- **Framing:** Length-prefixed binary/JSON framing over TCP.
- **Encryption:** AES-256-GCM session keys negotiated via authenticated Ed25519 handshake.

### 2.5 Protocol Engine (`internal/protocol`)
- Standardized envelope for `request`, `response`, and `event`.
- Request-response correlation via unique UUID `message_id`.
- Structured error codes (`INVALID_REQUEST`, `UNAUTHORIZED`, `CAPABILITY_NOT_FOUND`, etc.).

### 2.6 Capability Registry (`internal/capability`)
- Pluggable capability registration.
- Each peer advertises supported capabilities:
  - `terminal` (interactive pseudo-terminal streaming)
  - `command` (audited, authenticated remote command execution)
  - `files` (high-throughput parallel chunk streaming with SHA-256 validation)
  - `clipboard` (cross-device clipboard sharing with explicit sync)
  - `events` (developer automation hooks)
