# SynX — Full System Architecture & Backend Specification

**Version:** 1.0  
**Target:** Cross-platform desktop application  
**Primary stack:** Go + Wails  
**Platforms:** Windows, Linux, macOS  
**Network model:** Peer-to-peer over local network, with optional future remote transport  
**Current product direction:** Private, direct device-to-device file transfer first; synchronization/versioning is the next major layer.

---

## 1. Product Definition

SynX is a cross-platform desktop application for transferring files between trusted devices without requiring a central cloud-storage service.

The intended user flow is:

```text
Install SynX
   ↓
Application starts
   ↓
Discover nearby SynX devices
   ↓
Pair / trust a device
   ↓
Select files or folders
   ↓
Select destination device
   ↓
Create transfer
   ↓
Chunk / stream data
   ↓
Verify integrity
   ↓
Commit file
   ↓
Show completion + history
```

The system should be designed so that the same transfer engine can later support:

- folder synchronization
- version history
- conflict resolution
- scheduled synchronization
- selective sync
- background synchronization
- remote/WAN transfer
- device-to-device backup

---

# 2. Architecture Principles

SynX should follow these principles:

1. **Peer-to-peer by default**
2. **No central file-storage server**
3. **The desktop client is both a client and a server**
4. **Control plane and data plane are separated**
5. **Files are transferred in chunks**
6. **Transfers are resumable**
7. **Every completed file is integrity-verified**
8. **Untrusted network requests cannot escape the shared directory**
9. **Paired devices have explicit trust records**
10. **UI must never own transfer logic**
11. **Long-running operations run asynchronously**
12. **Transfer state survives application restarts**
13. **The protocol must be versioned**
14. **Transport should be replaceable without rewriting the sync engine**

---

# 3. High-Level Architecture

```text
                         SYNX DEVICE A
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│  Wails Desktop UI                                              │
│       │                                                        │
│       │ App API / Events                                       │
│       ▼                                                        │
│  Application Service                                           │
│       │                                                        │
│  ┌────┴───────────────┬────────────────┬──────────────────┐    │
│  │                    │                │                  │    │
│  ▼                    ▼                ▼                  ▼    │
│ Peer Manager      Transfer Manager  Device Manager    Settings │
│  │                    │                │                  │    │
│  ▼                    ▼                ▼                  │    │
│ Discovery          Sync Engine       Trust Store          │    │
│  │                    │                │                  │    │
│  └──────────────┬─────┴──────────────┴──────────────────┘    │
│                 │                                             │
│                 ▼                                             │
│          Transport Manager                                    │
│          ┌───────────────┐                                    │
│          │ LAN HTTP/QUIC │                                    │
│          └───────┬───────┘                                    │
│                  │                                             │
│          Chunk / Stream Engine                                 │
│                  │                                             │
│          Hash / Integrity                                      │
│                  │                                             │
│          Local Storage                                         │
│                                                                │
└──────────────────┬─────────────────────────────────────────────┘
                   │
                   │ encrypted / authenticated P2P
                   ▼
                         SYNX DEVICE B
```

Each SynX installation contains the same major services.

There is no permanent master/slave relationship.

A device can simultaneously:

- send
- receive
- advertise itself
- discover other devices
- accept incoming transfers
- synchronize folders
- maintain transfer history

---

# 4. Control Plane vs Data Plane

This separation is important.

## Control Plane

Responsible for:

- discovery
- pairing
- authentication
- device metadata
- transfer negotiation
- capability negotiation
- file metadata
- transfer creation
- resume negotiation
- cancellation
- pause/resume commands
- status events

Example:

```text
Device A → Device B

"Do you support chunked upload?"
"Yes."
"File is 8.2 GB."
"SHA-256 = ..."
"Chunk size = 8 MB."
"Existing bytes = 4.1 GB."
"Resume from offset 4.1 GB."
```

## Data Plane

Responsible only for:

- file bytes
- chunks
- streaming
- checksums
- retransmission
- flow control

The data plane should not contain UI concepts.

---

# 5. Backend Services

The backend should be divided into the following logical services.

```text
internal/
├── app/
├── config/
├── discovery/
├── identity/
├── pairing/
├── peers/
├── protocol/
├── transport/
├── transfer/
├── chunk/
├── integrity/
├── storage/
├── filesystem/
├── history/
├── events/
├── security/
├── sync/
├── conflict/
├── database/
└── observability/
```

These are Go packages/services, not necessarily separate processes.

---

# 6. Application Service

### Package

```text
internal/app
```

### Responsibility

The application service is the orchestration layer.

It coordinates:

- startup
- shutdown
- service initialization
- service lifecycle
- UI commands
- event publishing

### Responsibilities

```text
Start()
Stop()
GetState()
GetDevices()
GetTransfers()
SendFile()
SendFolder()
CancelTransfer()
PauseTransfer()
ResumeTransfer()
PairDevice()
RemoveDevice()
ChooseSharedFolder()
```

The UI talks to this layer instead of directly accessing storage, networking, or transfer code.

---

# 7. Configuration Service

### Package

```text
internal/config
```

Stores:

```yaml
device:
  name: "Kunal-PC"

network:
  port: 8787
  discovery_port: 8788

storage:
  shared_directory: "~/SynX"
  download_directory: "~/Downloads/SynX"

security:
  require_pairing: true
  auto_accept: false

transfer:
  chunk_size: 8388608
  max_parallel_chunks: 4
  max_parallel_transfers: 2
```

Configuration should be persisted locally.

Recommended locations:

### Windows

```text
%APPDATA%\SynX\
```

### Linux

```text
~/.config/synx/
```

### macOS

```text
~/Library/Application Support/SynX/
```

---

# 8. Device Identity Service

### Package

```text
internal/identity
```

Every installation gets a permanent device identity.

Example:

```json
{
  "device_id": "01JY...XYZ",
  "device_name": "Kunal-PC",
  "platform": "windows",
  "version": "1.0.0"
}
```

Do not use IP address as device identity.

IP addresses change.

The device ID remains stable.

## Identity generation

Use a cryptographically random UUID/ULID.

Future security model:

```text
Device ID
   +
Public Key
   +
Private Key
```

The private key never leaves the device.

---

# 9. Peer Discovery Service

### Package

```text
internal/discovery
```

Purpose:

Find SynX devices on the same network.

## Initial implementation

UDP broadcast or multicast.

Example:

```text
UDP port: 8788
```

Broadcast packet:

```json
{
  "protocol": "synx",
  "version": 1,
  "device_id": "abc123",
  "name": "Kunal-PC",
  "platform": "windows",
  "port": 8787
}
```

A receiving device responds:

```json
{
  "device_id": "xyz789",
  "name": "Kunal-Mac",
  "port": 8787
}
```

## Discovery state

Each discovered peer gets:

```text
DISCOVERED
    ↓
REACHABLE
    ↓
PAIRED
    ↓
TRUSTED
```

An offline device becomes:

```text
ONLINE
  ↓
TIMEOUT
  ↓
OFFLINE
```

Discovery must not automatically imply trust.

---

# 10. Peer Manager

### Package

```text
internal/peers
```

Maintains currently known devices.

```go
type Peer struct {
    ID          string
    Name        string
    Address     string
    Port        int
    Platform    string
    Version     string
    LastSeen    time.Time
    Status      PeerStatus
    Trusted     bool
}
```

Functions:

```text
AddPeer()
UpdatePeer()
RemovePeer()
GetPeer()
ListPeers()
MarkOffline()
MarkOnline()
```

---

# 11. Pairing Service

### Package

```text
internal/pairing
```

Pairing creates trust between two devices.

## Initial pairing

Device A displays:

```text
Pairing Code

842 193
```

Device B enters the code.

Then:

```text
Device A
   ↓
Challenge
   ↓
Device B
   ↓
Response
   ↓
Device A
   ↓
Trusted
```

## Future recommended implementation

Use public-key authentication.

```text
Device A public key
Device B public key
        ↓
Trust record
```

Store:

```text
peer_id
public_key
device_name
created_at
last_seen
```

Do not permanently rely on a short token as the only authentication mechanism.

---

# 12. Trust Store

### Package

```text
internal/security
```

The trust store records paired devices.

Example:

```json
{
  "peer_id": "01J...",
  "public_key": "...",
  "name": "Kunal-Laptop",
  "trusted_at": "2026-09-26T20:00:00+05:30"
}
```

Operations:

```text
TrustPeer()
IsTrusted()
RevokePeer()
ListTrustedPeers()
```

---

# 13. Transport Manager

### Package

```text
internal/transport
```

The transport abstraction prevents the rest of the application from depending directly on HTTP.

Interface:

```go
type Transport interface {
    Connect(ctx context.Context, peer Peer) (Connection, error)
    Send(ctx context.Context, conn Connection, req Request) error
    Receive(ctx context.Context, conn Connection) (Response, error)
    Close() error
}
```

Initial implementation:

```text
HTTPTransport
```

Future:

```text
QUICTransport
WebRTCTransport
RelayTransport
```

This allows SynX to evolve without rewriting the transfer engine.

---

# 14. LAN HTTP Service

### Package

```text
internal/server
```

Current MVP already has this basic concept.

Current endpoints include:

```text
GET    /api/info
GET    /api/files
GET    /api/download
POST   /api/upload
DELETE /api/delete
```

The current implementation uses an `X-SynX-Token` header for authentication and supports an upload offset. This is suitable for the MVP but should evolve into signed/authenticated sessions for production.

---

# 15. Production API

Recommended API structure:

```text
/api/v1/device/info
/api/v1/device/capabilities

/api/v1/pair/start
/api/v1/pair/confirm

/api/v1/files/list
/api/v1/files/stat
/api/v1/files/hash

/api/v1/transfers
/api/v1/transfers/{id}
/api/v1/transfers/{id}/pause
/api/v1/transfers/{id}/resume
/api/v1/transfers/{id}/cancel

/api/v1/transfers/{id}/chunks
/api/v1/transfers/{id}/commit

/api/v1/history
```

All APIs should be versioned.

---

# 16. Transfer Manager

### Package

```text
internal/transfer
```

This is the core backend service.

It manages the lifecycle of every transfer.

## Transfer states

```text
CREATED
   ↓
NEGOTIATING
   ↓
QUEUED
   ↓
TRANSFERRING
   ↓
VERIFYING
   ↓
COMMITTING
   ↓
COMPLETED
```

Failure states:

```text
FAILED
CANCELLED
PAUSED
INTERRUPTED
```

## Transfer object

```go
type Transfer struct {
    ID              string
    PeerID          string
    Direction       Direction
    SourcePath      string
    DestinationPath string
    FileName        string
    Size            int64
    BytesTransferred int64
    ChunkSize       int64
    TotalChunks     int
    CompletedChunks int
    Status          Status
    SHA256          string
    CreatedAt       time.Time
    StartedAt       time.Time
    CompletedAt     *time.Time
}
```

---

# 17. Transfer Queue

The UI should never directly execute uploads.

Instead:

```text
UI
 ↓
CreateTransfer()
 ↓
Transfer Queue
 ↓
Worker
 ↓
Transport
```

Example:

```text
Queue
────────────────────────
1. photos.zip      QUEUED
2. video.mp4       QUEUED
3. project.tar.gz  QUEUED
```

Workers process the queue.

Configuration:

```text
max concurrent transfers = 2
```

---

# 18. Chunk Service

### Package

```text
internal/chunk
```

Large files should not be treated as one giant request.

Example:

```text
8 GB file
        ↓
8 MB chunks
        ↓
1024 chunks
```

Each chunk:

```json
{
  "transfer_id": "tx_123",
  "chunk_index": 382,
  "offset": 3204448256,
  "size": 8388608,
  "sha256": "..."
}
```

Recommended default:

```text
8 MB
```

Allow configuration later.

---

# 19. Resumable Transfer

This is a critical SynX feature.

Suppose:

```text
File size: 8 GB
Transferred: 5.4 GB
Connection lost
```

SynX should NOT restart.

Instead:

```text
Reconnect
   ↓
Query transfer state
   ↓
Remote says:
chunks 0–690 complete
   ↓
Continue from chunk 691
```

Use a transfer manifest:

```json
{
  "transfer_id": "tx_123",
  "file_size": 8589934592,
  "chunk_size": 8388608,
  "completed": [
    0,
    1,
    2,
    3
  ]
}
```

---

# 20. Integrity Service

### Package

```text
internal/integrity
```

Use SHA-256 initially.

Calculate:

```text
Source SHA-256
        ↓
Transfer
        ↓
Destination SHA-256
        ↓
Compare
```

If equal:

```text
VERIFIED
```

If different:

```text
INTEGRITY_FAILURE
```

The destination file should not be marked complete until verification succeeds.

---

# 21. Temporary File Strategy

Never directly write a transfer into the final filename.

Instead:

```text
video.mp4.synx.part
```

Transfer:

```text
video.mp4.synx.part
```

Verification:

```text
SHA-256 OK
```

Commit:

```text
video.mp4.synx.part
        ↓
video.mp4
```

This prevents incomplete files from appearing as completed files.

---

# 22. Storage Service

### Package

```text
internal/storage
```

Responsibilities:

- transfer metadata
- manifests
- peer records
- configuration
- history
- interrupted transfers

Recommended database:

```text
SQLite
```

SynX does not need PostgreSQL for the desktop client.

---

# 23. Database Schema

## devices

```sql
CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    platform TEXT,
    version TEXT,
    public_key TEXT,
    trusted INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    last_seen INTEGER
);
```

## transfers

```sql
CREATE TABLE transfers (
    id TEXT PRIMARY KEY,
    peer_id TEXT NOT NULL,
    direction TEXT NOT NULL,
    file_name TEXT NOT NULL,
    source_path TEXT,
    destination_path TEXT,
    size INTEGER NOT NULL,
    bytes_transferred INTEGER NOT NULL DEFAULT 0,
    sha256 TEXT,
    chunk_size INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    started_at INTEGER,
    completed_at INTEGER
);
```

## chunks

```sql
CREATE TABLE chunks (
    transfer_id TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    offset INTEGER NOT NULL,
    size INTEGER NOT NULL,
    sha256 TEXT,
    status TEXT NOT NULL,
    PRIMARY KEY (transfer_id, chunk_index)
);
```

## settings

```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

## transfer_history

```sql
CREATE TABLE transfer_history (
    id TEXT PRIMARY KEY,
    transfer_id TEXT,
    peer_id TEXT,
    file_name TEXT,
    size INTEGER,
    direction TEXT,
    status TEXT,
    completed_at INTEGER
);
```

---

# 24. Filesystem Service

### Package

```text
internal/filesystem
```

Never allow the HTTP/API layer to manipulate arbitrary filesystem paths.

Use:

```text
requested relative path
        ↓
sanitize
        ↓
resolve inside allowed root
        ↓
verify root containment
        ↓
filesystem operation
```

Protection required against:

```text
../
..\ 
absolute paths
symlinks
UNC paths
drive traversal
```

Important:

```text
C:\Users\Kunal\file.txt
```

must never be accepted as a remote relative path.

---

# 25. File Scanner

### Package

```text
internal/filesystem/scanner
```

Used to:

- enumerate directories
- calculate sizes
- calculate hashes
- create manifests
- detect changes

For large directories, scanning must be asynchronous.

---

# 26. File Watcher

### Package

```text
internal/sync/watcher
```

For the future synchronization engine, use OS filesystem events.

Recommended Go library:

```text
fsnotify
```

Events:

```text
CREATE
WRITE
RENAME
REMOVE
CHMOD
```

Pipeline:

```text
Filesystem event
      ↓
Debounce
      ↓
Normalize path
      ↓
Hash file
      ↓
Create change event
      ↓
Sync engine
```

Do not immediately transfer on every write event.

Applications can produce dozens of events for one save operation.

---

# 27. Change Journal

### Package

```text
internal/sync/journal
```

Every filesystem change gets recorded.

Example:

```json
{
  "event_id": "evt_123",
  "path": "Documents/report.pdf",
  "operation": "modified",
  "hash": "abc...",
  "size": 1293812,
  "timestamp": 178...",
  "device_id": "device_a"
}
```

This makes synchronization restartable.

---

# 28. Sync Engine

### Package

```text
internal/sync
```

The sync engine compares local and remote manifests.

Example:

```text
Local:
A.txt
B.txt
C.txt

Remote:
A.txt
B.txt
D.txt
```

Result:

```text
A.txt → unchanged
B.txt → unchanged
C.txt → upload
D.txt → download
```

---

# 29. Manifest

Each synchronized directory should have a manifest.

Example:

```json
{
  "version": 1,
  "device_id": "device_a",
  "files": [
    {
      "path": "photos/a.jpg",
      "size": 123456,
      "modified": 1780000000,
      "sha256": "..."
    }
  ]
}
```

Comparison should primarily use content hashes, not timestamps.

---

# 30. Conflict Detection

Example:

```text
Device A edits report.txt
Device B edits report.txt
        ↓
Both changes occur before synchronization
        ↓
CONFLICT
```

Do not silently overwrite one version.

Possible policy:

```text
report.txt
report (Kunal-PC conflict 2026-09-26).txt
```

Future UI:

```text
Conflict detected

report.txt

○ Keep this device
○ Keep remote version
○ Keep both
○ Compare
```

---

# 31. Event Bus

### Package

```text
internal/events
```

The backend should publish events to the Wails UI.

Examples:

```text
peer.discovered
peer.connected
peer.disconnected

transfer.created
transfer.started
transfer.progress
transfer.paused
transfer.resumed
transfer.completed
transfer.failed

sync.started
sync.progress
sync.completed

conflict.detected
```

Example event:

```json
{
  "type": "transfer.progress",
  "transfer_id": "tx_123",
  "bytes": 419430400,
  "total": 838860800,
  "speed": 18432000,
  "eta": 23
}
```

---

# 32. Wails Bridge

The frontend should communicate with Go through Wails bindings.

Example:

```go
func (a *App) GetDevices() []Peer
func (a *App) CreateTransfer(req TransferRequest) error
func (a *App) CancelTransfer(id string) error
func (a *App) PauseTransfer(id string) error
func (a *App) ResumeTransfer(id string) error
func (a *App) GetTransfer(id string) Transfer
func (a *App) ChooseFolder() string
```

For continuous updates:

```text
Go Event Bus
      ↓
Wails runtime.EventsEmit()
      ↓
Frontend listener
      ↓
React/Vue/etc state
      ↓
UI update
```

The frontend should not poll every second for transfer progress.

---

# 33. Transfer Progress

The backend should calculate:

```text
bytes transferred
total bytes
percentage
instant speed
average speed
ETA
chunks completed
chunks remaining
```

Example:

```text
File: ubuntu.iso
Size: 5.2 GB

████████████████░░░░ 78%

4.06 GB / 5.2 GB
Speed: 84.2 MB/s
ETA: 14 seconds
```

Use a rolling speed window rather than calculating speed from application startup.

---

# 34. Parallel Transfer

For large files:

```text
File
 ↓
Chunk 0 ──────┐
Chunk 1 ──────┤
Chunk 2 ──────┤ → network
Chunk 3 ──────┤
Chunk 4 ──────┘
```

But parallelism must be bounded.

Example:

```text
max_parallel_chunks = 4
```

Too much concurrency can:

- saturate the network
- increase memory usage
- increase disk contention
- reduce performance on low-end devices

---

# 35. Bandwidth Manager

### Package

```text
internal/transport/ratelimit
```

Future settings:

```text
Unlimited
10 MB/s
25 MB/s
50 MB/s
100 MB/s
Custom
```

This prevents SynX from consuming the entire LAN bandwidth.

---

# 36. Security Architecture

Security should evolve in phases.

## MVP

```text
Pairing token
+
Trusted peer
+
HTTPS/TLS or authenticated transport
```

## Production

```text
Device identity
       +
Public/private key
       +
Mutual authentication
       +
Encrypted transport
       +
Per-session authorization
```

Do not rely solely on:

```text
IP address
```

or:

```text
device name
```

for authorization.

---

# 37. Session Security

When a transfer begins:

```text
Peer authentication
       ↓
Session creation
       ↓
Session ID
       ↓
Transfer authorization
       ↓
Data transfer
```

A session should have:

```text
session_id
peer_id
created_at
expires_at
capabilities
```

This prevents an old token from being valid indefinitely.

---

# 38. Capability Negotiation

Before transfer:

```json
{
  "protocol_version": 1,
  "transports": [
    "http",
    "quic"
  ],
  "max_chunk_size": 16777216,
  "compression": false,
  "resume": true,
  "parallel_chunks": true,
  "folder_transfer": true
}
```

The two devices select a compatible feature set.

---

# 39. Transfer Negotiation

Example:

```text
A → B

TRANSFER_REQUEST
file = movie.mkv
size = 8,923,421,123
sha256 = abc...
chunk_size = 8MB

B → A

TRANSFER_ACCEPT
resume = true
existing_bytes = 4,194,304,000
```

Then:

```text
A → B
CHUNK 501

A → B
CHUNK 502

...
```

Finally:

```text
A → B
TRANSFER_COMMIT

B:
hash verified

B → A
TRANSFER_COMPLETE
```

---

# 40. Folder Transfer

A folder should not be transferred as a single opaque archive by default.

Instead:

```text
Folder
 ↓
Manifest
 ↓
Files
 ↓
Per-file transfer
 ↓
Directory reconstruction
```

Example:

```text
Photos/
├── 2025/
│   ├── img1.jpg
│   └── img2.jpg
└── 2026/
    └── img3.jpg
```

This allows:

- resume
- deduplication
- partial retry
- progress by file
- conflict handling

---

# 41. History Service

### Package

```text
internal/history
```

Records:

```text
sent
received
failed
cancelled
```

History entry:

```json
{
  "file": "project.zip",
  "size": 4294967296,
  "peer": "MacBook",
  "direction": "sent",
  "status": "completed",
  "timestamp": 1780000000
}
```

---

# 42. Notification Service

### Package

```text
internal/notifications
```

Examples:

```text
Transfer complete
Transfer failed
New device discovered
Pairing request received
Conflict detected
```

Desktop notifications should be triggered from backend events.

---

# 43. Observability

### Package

```text
internal/observability
```

Log levels:

```text
DEBUG
INFO
WARN
ERROR
```

Example:

```text
INFO transfer started
INFO peer connected
INFO chunk 123 completed
WARN peer connection unstable
ERROR integrity verification failed
```

Never log:

- private keys
- authentication tokens
- file contents
- sensitive paths unnecessarily

---

# 44. Recommended Project Structure

```text
SynX/
│
├── cmd/
│   └── synx/
│       └── main.go
│
├── desktop/
│   ├── frontend/
│   └── bindings/
│
├── internal/
│   │
│   ├── app/
│   │   ├── app.go
│   │   ├── lifecycle.go
│   │   └── commands.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── defaults.go
│   │
│   ├── identity/
│   │   ├── identity.go
│   │   └── keys.go
│   │
│   ├── discovery/
│   │   ├── discovery.go
│   │   ├── broadcaster.go
│   │   └── listener.go
│   │
│   ├── peers/
│   │   ├── manager.go
│   │   └── peer.go
│   │
│   ├── pairing/
│   │   ├── pairing.go
│   │   └── challenge.go
│   │
│   ├── security/
│   │   ├── trust.go
│   │   ├── session.go
│   │   └── crypto.go
│   │
│   ├── protocol/
│   │   ├── version.go
│   │   ├── messages.go
│   │   └── capabilities.go
│   │
│   ├── transport/
│   │   ├── transport.go
│   │   ├── http/
│   │   └── quic/
│   │
│   ├── server/
│   │   ├── server.go
│   │   ├── middleware.go
│   │   └── handlers/
│   │
│   ├── transfer/
│   │   ├── manager.go
│   │   ├── queue.go
│   │   ├── worker.go
│   │   └── state.go
│   │
│   ├── chunk/
│   │   ├── chunk.go
│   │   ├── reader.go
│   │   └── writer.go
│   │
│   ├── integrity/
│   │   ├── hash.go
│   │   └── verify.go
│   │
│   ├── filesystem/
│   │   ├── filesystem.go
│   │   ├── scanner.go
│   │   └── safe_path.go
│   │
│   ├── storage/
│   │   ├── database.go
│   │   ├── migrations/
│   │   └── repositories/
│   │
│   ├── history/
│   │   └── history.go
│   │
│   ├── events/
│   │   └── bus.go
│   │
│   ├── sync/
│   │   ├── engine.go
│   │   ├── watcher/
│   │   ├── journal/
│   │   └── manifest/
│   │
│   ├── conflict/
│   │   └── resolver.go
│   │
│   └── observability/
│       ├── logger.go
│       └── metrics.go
│
├── migrations/
├── build/
├── scripts/
├── go.mod
└── wails.json
```

---

# 45. Complete Transfer Flow

## Sender

```text
User selects file
        ↓
UI calls CreateTransfer()
        ↓
Transfer Manager
        ↓
Validate peer
        ↓
Validate file
        ↓
Calculate metadata
        ↓
Create transfer record
        ↓
Queue transfer
        ↓
Worker starts
        ↓
Connect to peer
        ↓
Authenticate session
        ↓
Negotiate capabilities
        ↓
Send transfer request
        ↓
Receive resume state
        ↓
Read chunk
        ↓
Hash chunk
        ↓
Send chunk
        ↓
Update progress
        ↓
Repeat
        ↓
Send commit
        ↓
Remote verifies
        ↓
Receive completion
        ↓
Mark transfer complete
        ↓
Write history
        ↓
Emit UI event
```

---

# 46. Receiver Flow

```text
Incoming request
        ↓
Authenticate peer
        ↓
Check trusted device
        ↓
Validate destination
        ↓
Create .synx.part
        ↓
Check existing manifest
        ↓
Return resume state
        ↓
Receive chunk
        ↓
Verify chunk
        ↓
Write chunk
        ↓
Update manifest
        ↓
Repeat
        ↓
Receive COMMIT
        ↓
Calculate final SHA-256
        ↓
Compare expected hash
        ↓
Rename .synx.part
        ↓
Write history
        ↓
Return COMPLETE
```

---

# 47. Failure Recovery

## Network disconnect

```text
TRANSFERING
     ↓
CONNECTION_LOST
     ↓
RETRY
     ↓
RECONNECT
     ↓
QUERY_RESUME_STATE
     ↓
CONTINUE
```

Recommended retry:

```text
1s
2s
4s
8s
16s
30s maximum
```

Stop after a configurable number of attempts.

---

# 48. Disk Failure

If writing fails:

```text
TRANSFER_FAILED
```

Keep the partial file if recovery is possible.

Do not report success.

---

# 49. Hash Mismatch

```text
Destination hash != source hash
```

Then:

```text
Mark chunk/file invalid
       ↓
Retry affected chunks
       ↓
Recalculate final hash
```

After repeated failures:

```text
INTEGRITY_FAILED
```

---

# 50. Application Crash Recovery

On startup:

```text
Read database
   ↓
Find TRANSFERRING records
   ↓
Mark INTERRUPTED
   ↓
Check .synx.part files
   ↓
Resume when user/network allows
```

This is why transfer state must be persisted.

---

# 51. Concurrency Model

Recommended Go model:

```text
Application
    │
    ├── Discovery goroutine
    │
    ├── Server goroutine
    │
    ├── Transfer queue
    │      ├── Worker 1
    │      ├── Worker 2
    │      └── Worker N
    │
    ├── Event bus
    │
    └── File watcher
```

Use:

```text
context.Context
sync.Mutex / RWMutex
channels
errgroup
WaitGroup
```

Avoid uncontrolled goroutine creation.

---

# 52. UI ↔ Backend Contract

The UI should know about:

```text
Device
Transfer
TransferProgress
HistoryEntry
Settings
SyncStatus
```

The UI should NOT know about:

```text
HTTP requests
filesystem internals
chunk writers
SQLite queries
cryptographic implementation
network sockets
```

This keeps the frontend replaceable.

---

# 53. UI Pages and Backend Dependencies

## Dashboard

Uses:

```text
Peer Manager
Transfer Manager
Event Bus
```

## Devices

Uses:

```text
Discovery
Peer Manager
Pairing
Trust Store
```

## Transfer Center

Uses:

```text
Transfer Manager
Event Bus
History
```

## History

Uses:

```text
History Service
Storage
```

## Settings

Uses:

```text
Configuration
Device Identity
Network Service
Storage
```

## Sync

Uses:

```text
File Watcher
Change Journal
Manifest
Sync Engine
Conflict Resolver
Transfer Manager
```

---

# 54. API Example — Create Transfer

```http
POST /api/v1/transfers
X-SynX-Session: session_123
Content-Type: application/json
```

```json
{
  "source": "Photos/2026/vacation.zip",
  "destination": "Downloads/vacation.zip",
  "size": 829423412,
  "sha256": "..."
}
```

Response:

```json
{
  "transfer_id": "tx_01J...",
  "accepted": true,
  "chunk_size": 8388608,
  "resume_offset": 0
}
```

---

# 55. API Example — Progress

```http
GET /api/v1/transfers/tx_01J...
```

```json
{
  "id": "tx_01J...",
  "status": "transferring",
  "bytes_transferred": 419430400,
  "total_bytes": 838860800,
  "speed": 84200000,
  "eta_seconds": 5
}
```

For the desktop UI, prefer backend events instead of polling this endpoint.

---

# 56. API Example — Chunk

```http
PUT /api/v1/transfers/tx_01J.../chunks/52
X-Chunk-Offset: 436207616
X-Chunk-Hash: sha256...
```

Body:

```text
raw binary bytes
```

Response:

```json
{
  "chunk": 52,
  "accepted": true
}
```

---

# 57. API Example — Commit

```http
POST /api/v1/transfers/tx_01J.../commit
```

```json
{
  "sha256": "..."
}
```

Response:

```json
{
  "status": "completed",
  "verified": true
}
```

---

# 58. Protocol Versioning

Every request should include:

```text
SynX-Protocol-Version: 1
```

Future:

```text
v1
v2
v3
```

The application should reject unsupported protocol versions cleanly.

---

# 59. Performance Targets

These are engineering targets, not guarantees.

### LAN

Aim for:

```text
Gigabit LAN:
100+ MB/s when hardware/network permits
```

### CPU

Hashing should use streaming I/O.

### Memory

Do not load an entire file into memory.

Bad:

```go
data, _ := os.ReadFile(file)
```

Good:

```go
io.Copy(writer, reader)
```

For chunks:

```text
Read 8 MB
Hash
Send
Release
```

---

# 60. Deduplication — Future

SynX can eventually avoid sending content already present on the destination.

Example:

```text
File A
 ↓
Chunk hashes

Destination already has:
chunk 0
chunk 1
chunk 3

Only send:
chunk 2
chunk 4
...
```

This becomes especially valuable for folder synchronization.

---

# 61. Compression — Future

Do not blindly compress everything.

Good candidates:

```text
TXT
JSON
CSV
source code
logs
```

Poor candidates:

```text
JPEG
PNG
MP4
ZIP
7z
RAR
```

Capability negotiation can determine whether compression is worthwhile.

---

# 62. WAN / Internet Architecture — Future

The first release should focus on LAN.

Later:

```text
Device A
   │
   ├── direct LAN → Device B
   │
   ├── direct WAN → Device B
   │
   └── relay → Device B
```

A future relay server should forward encrypted traffic only.

It should not need access to file contents.

---

# 63. Recommended Technology Stack

```text
Desktop shell       Wails
Backend              Go
Frontend             React/TypeScript or current frontend
Database             SQLite
Discovery            UDP multicast/broadcast
Initial transport    HTTP/1.1
Future transport     QUIC
Hashing              SHA-256
File watching        fsnotify
Serialization        JSON initially
Encryption            TLS / Noise-style authenticated sessions
Logging              structured Go logger
Packaging            Wails native builds
```

---

# 64. What Is Already Implemented

The current desktop code already establishes the foundation for:

- Wails desktop shell
- Go backend
- LAN HTTP server
- token authentication
- shared folder
- file listing
- file download
- file upload
- upload offset/resume foundation
- peer pairing foundation
- peer state
- transfer state model
- LAN discovery foundation
- modern desktop UI

The current README specifically identifies the next backend layer as connecting the drag-and-drop queue to remote upload with real progress/resume tracking.

---

# 65. What Is Still Missing

The production architecture still needs:

### Critical

- Real transfer queue
- Transfer workers
- Real-time progress events
- Persistent transfer state
- Proper chunk protocol
- Resume manifest
- Final integrity verification
- Robust peer authentication
- SQLite persistence
- Error/retry system
- Proper filesystem abstraction

### Next

- Folder transfer
- Transfer history
- Notifications
- Device trust management
- Conflict handling
- File watcher
- Sync engine

### Later

- QUIC
- WAN traversal
- relay
- deduplication
- version history
- selective sync
- multi-device synchronization

---

# 66. Implementation Roadmap

## Phase 1 — Real Transfer Engine

```text
Transfer Manager
      ↓
Queue
      ↓
Workers
      ↓
HTTP transport
      ↓
Chunk upload
      ↓
Progress events
```

## Phase 2 — Persistence

```text
SQLite
 ↓
Transfers
Peers
History
Chunks
Settings
```

## Phase 3 — Production Security

```text
Device identity
 ↓
Public keys
 ↓
Mutual authentication
 ↓
Encrypted sessions
```

## Phase 4 — Better Transport

```text
HTTP
 ↓
QUIC
```

## Phase 5 — Sync

```text
fsnotify
 ↓
Change journal
 ↓
Manifest
 ↓
Sync engine
 ↓
Conflict resolver
```

## Phase 6 — Advanced Features

```text
Deduplication
Version history
Selective sync
WAN
Relay
Mobile clients
```

---

# 67. Final Architecture

The intended final SynX architecture is:

```text
                         ┌─────────────────────┐
                         │     Wails UI        │
                         └──────────┬──────────┘
                                    │
                              Wails Bridge
                                    │
                         ┌──────────▼──────────┐
                         │   Application Core  │
                         └──────────┬──────────┘
                                    │
       ┌────────────────────────────┼────────────────────────────┐
       │                            │                            │
       ▼                            ▼                            ▼
 Discovery                    Transfer Manager             Sync Engine
       │                            │                            │
       ▼                            ▼                            ▼
 Peer Manager                 Transfer Queue              File Watcher
       │                            │                            │
 Pairing                       Workers                    Change Journal
       │                            │                            │
 Trust Store                   Chunk Engine                  Manifest
       │                            │                            │
       └──────────────┬─────────────┴──────────────┬─────────────┘
                      │                            │
                      ▼                            ▼
                Protocol Layer              Integrity Layer
                      │                            │
                      └────────────┬───────────────┘
                                   │
                           Transport Manager
                          ┌────────┴────────┐
                          │                 │
                       HTTP/1.1           QUIC
                          │                 │
                          └────────┬────────┘
                                   │
                              Network
                                   │
                        ┌──────────▼──────────┐
                        │    Remote SynX      │
                        └─────────────────────┘

                 Local persistence:
                 ┌──────────────────────────┐
                 │ SQLite                   │
                 │ - devices                │
                 │ - transfers              │
                 │ - chunks                 │
                 │ - history                │
                 │ - settings               │
                 └──────────────────────────┘
```

---

# 68. Engineering Rule

The most important architectural rule for SynX is:

> **The transfer engine must be independent of the UI and transport.**

That gives SynX this future:

```text
                    Sync Engine
                         │
                  Transfer Manager
                         │
                  Transport Interface
              ┌──────────┼───────────┐
              │          │           │
             LAN        QUIC        Relay
              │          │           │
              └──────────┼───────────┘
                         │
                       Peer
```

This prevents the project from becoming tightly coupled to the current Wails UI or HTTP implementation and makes the later synchronization system much easier to build.
