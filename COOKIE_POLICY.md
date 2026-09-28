# Cookie & Local Storage Policy for SynX

**Last Updated:** September 28, 2026  
**Effective Date:** September 28, 2026

This Cookie and Local Storage Policy explains how **SynX** treats cookies and client-side storage technologies.

---

## 1. Zero HTTP / Tracking Cookies

**SynX does NOT use HTTP tracking cookies, third-party analytics cookies, or advertising beacons.**

SynX runs as a cross-platform desktop application built with Go, Wails v2, and Next.js, or as a standalone local CLI daemon serving a web dashboard on your local network (e.g., `http://localhost:8787`). Because SynX is not connected to commercial advertising networks, profiling platforms, or telemetry vendors, it never creates, reads, or transmits HTTP tracking cookies over the public internet.

---

## 2. Local Storage Technologies Used

To remember your interface preferences between app restarts without requiring cloud synchronization, SynX uses standard browser **`localStorage`** and transient in-memory state:

| Storage Type | Key / Identifier | Purpose | Lifetime | Storage Location |
| :--- | :--- | :--- | :--- | :--- |
| **`localStorage`** | `synx_ui_theme` | Remembers your active theme (OLED Void, Cyberpunk Neon, High Contrast). | Persistent until cleared | Local device only |
| **`localStorage`** | `synx_radar_prefs` | Remembers radar animation speed, range filtering, and audio blip preferences. | Persistent until cleared | Local device only |
| **`localStorage`** | `synx_consent_acknowledged` | Records that you have reviewed and acknowledged the local data and offline policy notice. | Persistent until cleared | Local device only |
| **In-Memory** | Live Socket Buffers | Holds active chunk streams, current bandwidth metrics, and dynamic radar peer blips. | Duration of app session | RAM only |

**Important:** None of the entries stored in `localStorage` or session memory are ever sent across the internet or shared with external parties.

---

## 3. How to Inspect or Clear Local Storage

Because SynX respects your device sovereignty:
- You can clear your browser storage or webview cache at any time via your browser's Developer Tools (`Application` → `Local Storage` → `Clear All`).
- To reset all desktop application state, delete the application data directory:
  - **Windows:** `%APPDATA%\SynX`
  - **macOS / Linux:** `~/.config/synx`

---

## 4. Contact & Inquiries

For questions regarding local storage practices:
- **GitHub Issues:** [https://github.com/kunal-live/SynX/issues](https://github.com/kunal-live/SynX/issues)
- **Repository:** [https://github.com/kunal-live/SynX](https://github.com/kunal-live/SynX)
