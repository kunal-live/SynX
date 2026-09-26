# SynX Desktop

Cross-platform LAN file transfer desktop application built with Go + Wails.

## Features in this desktop build

- Native desktop shell for Windows, Linux and macOS
- LAN peer discovery over UDP
- Token-based peer pairing
- Local HTTP transfer service
- Shared-folder selection
- Peer dashboard
- Transfer dashboard foundation
- Settings view
- Drag-and-drop transfer area foundation
- Path traversal protection

## Development

Install Wails v2.15.0, then:

```bash
wails doctor
wails dev
```

Build:

```bash
wails build
```

The generated executable will be placed in `build/bin/`.

## Notes

The current desktop shell preserves the MVP transfer protocol. The next implementation layer should connect the drag-and-drop queue to the remote `/api/upload` endpoint and add real progress/resume tracking.
