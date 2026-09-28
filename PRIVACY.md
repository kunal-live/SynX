# Privacy Policy for SynX

**Last Updated:** September 28, 2026  
**Effective Date:** September 28, 2026

At **SynX**, privacy, security, and digital sovereignty are fundamental principles. SynX is designed and engineered from the ground up as a **decentralized, local-first, peer-to-peer (P2P) desktop application** for high-speed file transfers and device connectivity over local area networks (LAN). 

We do not operate tracking servers, cloud relays, telemetry pipelines, or data collection services. This Privacy Policy details what data is handled by SynX, where it resides, and how your privacy is guaranteed.

---

## 1. Core Principles

- **Zero Cloud Middlemen**: File transfers occur directly between devices on your local network over encrypted sockets. Your files never pass through external servers.
- **Zero Telemetry**: SynX does not collect, log, or transmit telemetry, user analytics, crash reports, or device identifiers to any remote service.
- **Zero Third-Party Embeds**: The application UI is self-contained. It loads zero external fonts, scripts, CDNs, or tracking pixels.
- **Local Data Sovereignty**: All pairing keys, transfer logs, and configuration settings are stored exclusively on your local workstation.

---

## 2. What User Data Does SynX Handle?

Because SynX facilitates direct LAN transfers between authorized devices, it manages minimal local configuration and operational state:

### A. Local Node Identity & Preferences
- **What is handled:** Your chosen device display name, randomly generated local node ID, listening port (default: 8787), and user interface theme.
- **Where it is stored:** Saved locally in a pure-Go SQLite database (`synx.db`) within your local application profile:
  - **Windows:** `%APPDATA%\SynX\synx.db`
  - **macOS / Linux:** `~/.config/synx/synx.db`
- **Purpose:** Identifies your device to other SynX nodes on your local Wi-Fi / Ethernet network during UDP peer discovery.

### B. Device Pairing & Trust Records
- **What is handled:** Public keys (Ed25519) of paired devices, paired device labels, trust status (trusted / blocked), and ephemeral 6-digit PIN handshake verification tokens.
- **Where it is stored:** Stored locally in your local `synx.db` SQLite database.
- **Security:** Private keys never leave your device. The 6-digit PIN handshake is conducted solely across your local network.

### C. Transfer Records & Audit History
- **What is handled:** File names, byte sizes, peer node IDs, timestamps, and SHA-256 integrity checksums of transferred items.
- **Where it is stored:** Stored locally in `synx.db`.
- **Purpose:** Allows you to review past incoming and outgoing transfers and verify file integrity.

### D. Chroot Share Directories
- **What is handled:** Directory paths configured by you for incoming or outgoing shared files.
- **Containment:** SynX enforces strict chroot path sandboxing. Remote peers can never access, read, or overwrite files outside your explicitly shared directory.

---

## 3. What User Data We DO NOT Collect

To be completely unequivocal:
- We **DO NOT** collect your personal identity (name, email address, physical location, or phone number).
- We **DO NOT** inspect, scan, index, or store the contents of your transferred files.
- We **DO NOT** route transfers through external relay servers or cloud storage providers.
- We **DO NOT** use Google Analytics, PostHog, Mixpanel, Sentry, or any third-party tracking services.
- We **DO NOT** monetize, share, or sell user data under any circumstances.

---

## 4. Third-Party Embeds and Network Connections

SynX is strictly isolated from unsolicited external internet connections:
- **No Third-Party Scripts or Web Fonts**: The user interface uses local system typography stacks (`Plus Jakarta Sans`, `JetBrains Mono`, `Space Grotesk`, `Segoe UI`, `SF Mono`). No external font servers or CDNs are contacted.
- **Local Network Scope**: Discovery beacons (UDP port 8788) and data streams (TCP port 8787) operate strictly within the bounds of your local subnet (broadcast / multicast / unicast LAN). No outbound connections to the public internet are initiated unless you explicitly configure remote peer tunneling.

---

## 5. Security & Cryptographic Standards

SynX implements industry-standard local security measures:
- **Streaming SHA-256 Verification**: Every chunk transmitted is verified against its cryptographic hash before being written to disk, preventing corrupted or tampered payloads.
- **Strict Path Containment**: Rejects all path traversal attacks (`../`, `..\`), symbolic link escapes, UNC paths, and null-byte injection.
- **Cryptographic Device Authentication**: Peer authorizations use public-key cryptography (Ed25519) and time-limited PIN pairing.

---

## 6. Managing, Purging, and Deleting Your Data

You have complete control over all data stored by SynX:
- **Clear Transfer History**: You can clear transfer logs at any time directly through the application interface.
- **Purge All Application Data**: To permanently delete all SynX records, pairing tokens, and database files:
  - **Windows:** Delete `%APPDATA%\SynX`
  - **macOS / Linux:** Delete `~/.config/synx`
  Deleting this directory removes all device trust relationships, transfer histories, and local keys.

---

## 7. Changes to This Privacy Policy

Any revisions to this Privacy Policy will be committed directly to the official project repository with an updated "Last Updated" timestamp.

---

## 8. Contact & Inquiries

For questions or inquiries regarding SynX security or privacy:
- **GitHub Issues:** [https://github.com/kunal-live/SynX/issues](https://github.com/kunal-live/SynX/issues)
- **Repository:** [https://github.com/kunal-live/SynX](https://github.com/kunal-live/SynX)
