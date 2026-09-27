# SynX — Protocol Specification

**Protocol Version:** 1  
**Default TCP Session Port:** 8787  
**Default UDP Discovery Port:** 8788  

---

## 1. Discovery Announcement

Nodes broadcast UDP datagrams containing a JSON announcement:

```json
{
  "type": "hello",
  "protocol": "synx",
  "protocol_version": 1,
  "device_id": "synx_7f9b1c2d3e4f5a6b",
  "device_name": "Kunal-Workstation",
  "platform": "windows",
  "port": 8787,
  "capabilities": [
    "terminal",
    "command",
    "files",
    "clipboard"
  ],
  "public_key": "3a4f...e8",
  "timestamp": 1780000000
}
```

---

## 2. Session Protocol Envelope

All messages transmitted across established TCP connections use a uniform envelope:

```json
{
  "version": 1,
  "message_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "type": "request",
  "capability": "terminal",
  "action": "open",
  "payload": {
    "cols": 80,
    "rows": 24
  },
  "timestamp": 1780000000
}
```

### Message Types
- `request`: Sent from client to peer to invoke a capability action. Expects a `response`.
- `response`: Sent from peer to client in response to a `request`.
- `event`: Asynchronous one-way notification (e.g. streaming terminal output, chunk progress).

---

## 3. Capability Actions

### 3.1 Terminal Capability (`terminal`)
- `terminal.open`: Start interactive shell session.
- `terminal.input`: Send raw stdin keystrokes.
- `terminal.resize`: Adjust terminal dimensions (cols, rows).
- `terminal.close`: Terminate active shell session.

### 3.2 Command Capability (`command`)
- `command.execute`: Run non-interactive command, stream stdout/stderr.
- `command.cancel`: Abort in-flight process execution.

### 3.3 File Capability (`files`)
- `files.negotiate`: Request chunked file reception with SHA-256 verification.
- `files.chunk`: Stream binary file slices.
- `files.commit`: Verify integrity hash and atomically write to disk.

### 3.4 Clipboard Capability (`clipboard`)
- `clipboard.get`: Request current text buffer from peer.
- `clipboard.set`: Push text buffer to peer clipboard.

---

## 4. Standard Error Codes

| Code | Description |
|---|---|
| `INVALID_REQUEST` | Message envelope or payload failed validation |
| `UNAUTHORIZED` | Peer not authenticated or missing required token |
| `FORBIDDEN` | Peer does not possess permission for capability action |
| `PEER_NOT_FOUND` | Target device unreachable or unknown |
| `CAPABILITY_NOT_FOUND`| Peer does not advertise requested capability |
| `ACTION_NOT_SUPPORTED`| Capability does not implement requested action |
| `SESSION_NOT_FOUND` | Active session or terminal session does not exist |
| `TIMEOUT` | Operation timed out waiting for peer |
| `TRANSFER_FAILED` | File chunk transfer or hash verification failed |
| `PROTOCOL_VERSION_UNSUPPORTED` | Major protocol version mismatch |
| `INTERNAL_ERROR` | Internal subsystem failure |
