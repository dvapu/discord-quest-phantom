<div align="center">

# 👻 Discord Quest Phantom v1.1.0
### *Autonomous Cross-Platform Quest Completer for Windows & Linux (x64 / ARM64)*

[![Latest Release: v1.1.0](https://img.shields.io/badge/Release-v1.1.0%20(Stable)-7289da?style=flat-square&logo=github)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.0)
[![Nightly Build: v1.1.1](https://img.shields.io/badge/Nightly-v1.1.1%20(Pre--release)-f39c12?style=flat-square&logo=github)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.1)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages:**  
[🇺🇸 English](README.md) | [🇻🇳 Tiếng Việt](docs/translations/README.vi.md) | [🇨🇳 简体中文](docs/translations/README.zh.md) | [🇰🇷 한국어](docs/translations/README.ko.md) | [🇯🇵 日本語](docs/translations/README.ja.md) | [🇮🇳 हिन्दी](docs/translations/README.hi.md)

---

</div>

> ### ⚠️ DISCLAIMER
> **USE AT YOUR OWN RISK.**
> This software is developed purely for educational, network security, and reverse engineering research purposes. The use of automation tools may violate Discord's Terms of Service (ToS). The authors accept no responsibility or liability for any account warnings, restrictions, or bans resulting from the use of this program.

---

```
          ┌─────────────────────────────────────────────────────────────┐
          │                    DISCORD QUEST PHANTOM                    │
          │             v1.1.0 (Stable)  ·  v1.1.1 (Nightly)            │
          │    Multi-Region Sweep · Parallel Engine · Captcha Portal    │
          └──────────────────────────────┬──────────────────────────────┘
                                         │
                              Core System Orchestrator
                                         │
             ┌───────────────────────────┼───────────────────────────┐
             ▼                           ▼                           ▼
  [1. MULTI-REGION SWEEP]      [2. PARALLEL RUNNER]     [3. LOCAL CAPTCHA PORTAL]
  - Probes US, JP, VN zones   - Runs 2-5 quests parallel - Auto LAN IP & port scan
  - Spoofs client properties  - Staggered request jitter - Mobile 1-tap web solver
  - Unlocks hidden rewards    - 5 games in 15m (not 75m) - Zero SSH/headless locks
             │                           │                           │
             └───────────────────────────┴───────────────────────────┘
                                         │
                            Dual-Engine Execution Layer
                                         │
                    ┌────────────────────┬────────────────────┐
                    ▼                                         ▼
    [ENGINE 1: AUTONOMOUS API RUNNER]          [ENGINE 2: OS PROCESS SPOOFER]
  - 100% Headless (Linux, VPS, Armbian)    - For users running Discord Desktop app
  - Zero browser or Discord app required   - Concurrently spawns dummy game stubs
  - Auto video progress & stream heartbeat - Broadcasts native Gateway status
```

---

## 🚀 Key Highlights & Capabilities (v1.1.0 Stable · v1.1.1 Nightly)

1. **🌍 Golden Trio Multi-Region Auto-Scan (`--region all`)**:
   - Discord region-locks exclusive avatar decorations and quests by evaluating `client_locale` / `system_locale` (why changing iPhone locale unlocks hidden quests).
   - Phantom queries the Golden Trio (`en-US`, `ja-JP`, `vi-VN`) in parallel (~2.5s) to reveal 100% of global quests, avatar frames, and rewards directly inside your Discord app without proxy/VPN!
2. **⚡ Parallel / Concurrent Quest Execution (`-concurrency 5`)**:
   - Complete 2–5 quests concurrently in parallel lanes with staggered request jitter.
   - Finish 5 games in **15 minutes** instead of 75 minutes!
3. **📱 Local Captcha Web Portal (`-portal` & `-portal-port`)**:
   - Running headless on a 24/7 Linux server/Armbian (e.g. `192.168.1.200`) via SSH?
   - When Discord challenges an enrollment, Phantom spins up a local web portal on port `8080` (with auto port-conflict resolution).
   - Just tap the local LAN link on your phone (same WiFi) to solve the captcha in 3 seconds, and the headless daemon instantly resumes!
4. **🔄 Auto-Update Notification**:
   - Built-in updater queries GitHub releases at startup and alerts you when a new build is released.

---

## 🎯 Discord Quest Classification Architecture

Discord Quests come in multiple task structures. Discord Quest Phantom categorizes and handles each gracefully:

| Quest Category | Task Type Identifier | Technical Mechanism & Automation Handling | Automated? |
|---|---|---|:---:|
| **1. 🖥️ Pure PC Game** | `PLAY_ON_DESKTOP` (PC only) | Auto-heartbeats periodic stream frames (`call:0:<pid>`) every 20s or emulates Win32 process (`-spoofer`). | ✅ 100% Automated |
| **2. 🎮 Cross-Platform Game** | `PLAY_ON_DESKTOP` + `XBOX` / `PS` | Desktop app forces a "Choose Platform to start" popup. **Phantom bypasses this UI barrier entirely**, reporting stream progress straight into PC playtime! | ✅ 100% Automated |
| **3. 📺 Desktop Video** | `WATCH_VIDEO` | Progresses video playback timestamps smoothly via `/quests/{id}/video-progress`. Finishes in 10–30s. | ✅ 100% Automated |
| **4. 📱 Mobile Video** | `WATCH_VIDEO_ON_MOBILE` | Targeted at mobile users. The backend API is identical: **Phantom finishes it without needing any mobile device**! | ✅ 100% Automated |
| **5. 🏆 In-Game Achievement** | `ACHIEVEMENT_IN_GAME` / `ACHIEVEMENT_IN_ACTIVITY` | Missions like *VALORANT Aces* require real game servers (Riot/EA) to send webhooks to Discord. **Client tools cannot fake in-game kills**; Phantom automatically detects and cleanly skips these. | ⚠️ Skipped (Play manually) |

---

## 🔑 1. How to Retrieve Your Discord Token (F12 DevTools)

### Method 1: Console Tab (Fastest — 1 Line Command)
1. Open your web browser, navigate to [https://discord.com/app](https://discord.com/app), and log into your account.
2. Press `F12` (or `Ctrl + Shift + I`) to open Developer Tools → Switch to the **Console** tab.
   > 💡 **Console Paste Protection (Self-XSS)**: If your browser blocks pasting with a warning, type `allow pasting` into the Console and press `Enter` first to unlock pasting.
3. Paste the snippet below and press `Enter`:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. Your token string will immediately appear (e.g. `NTk5...` or `mfa....`). Copy this string (do not copy the enclosing quotation marks).

### Method 2: Network Tab
1. In DevTools (`F12`), switch to the **Network** tab and filter by `/api`.
2. Click on any request (`messages`, `users`, `science`).
3. In the right panel, scroll down to **Request Headers** and copy the string value of **`authorization:`**.

---

## ⚡ 2. Setup & Execution

### For Windows:
1. Download `discord-quest-phantom-windows-amd64.zip` from [Releases (v1.1.0 Stable)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.0) or [Nightly Build (v1.1.1)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.1).
2. Extract the archive into any folder.
3. Double-click `run.cmd` or `discord-quest-phantom.exe`:
   - If `.token` is missing, the program will interactively prompt you to paste your token, save it automatically, and start immediately!
   - No Python installation needed, no dependencies required, no Discord client needed!

### For Linux / VPS / Armbian / Raspberry Pi (24/7 Background Daemon):
Run this single command to download the standalone binary (v1.1.0 Stable) and start 24/7 background execution:
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.1.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "YOUR_DISCORD_TOKEN_HERE" > .token && \
nohup ./discord-quest-phantom -daemon -poll 15m -portal=false > quest.log 2>&1 &
```
> 💡 *Note: To test bleeding-edge Nightly features, replace `v1.1.0` with `v1.1.1` in the curl command above.*

#### Running as a Systemd Service (Auto-start on Boot):
Create `/etc/systemd/system/discord-quest-phantom.service`:
```ini
[Unit]
Description=Discord Quest Phantom Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/discord-quest-phantom
ExecStart=/opt/discord-quest-phantom/discord-quest-phantom -daemon -poll 15m -portal=false -concurrency 5 -region all
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```
Then enable and start the service:
```bash
systemctl daemon-reload && systemctl enable --now discord-quest-phantom
```

---

## ⚙️ CLI Options & Configuration Flags

| Flag | Default | Description |
|---|---|---|
| `-region` | `all` | Regional probe scope: `all` (Golden Trio: US+JP+VN), `us`, `jp`, `vn` |
| `-concurrency` | `5` | Maximum concurrent quests running in parallel (1–5) |
| `-daemon` | `false` | Run continuously as background daemon/service, re-checking quests periodically |
| `-poll` | `60s` | Polling interval between quest checks when running in daemon mode (e.g. `15m`, `30m`) |
| `-portal` | `true` | Enable local LAN captcha web portal (`false` for completely silent headless background run) |
| `-portal-port` | `8080` | Port for captcha web portal (auto-increments if port is busy) |
| `-spoofer` | `false` | Enable Win32 OS process emulation (requires Discord Desktop running) |
| `-lang` | `auto` | Language override: `auto`, `en`, `vi` |
| `-dry-run` | `false` | Discover and display active quests without taking any actions |
| `-token` | `""` | Provide Discord token directly via command-line argument |

---

## 🛡️ Anti-Ban Security Analysis

* **Never Auto-Claims Rewards**: Phantom **NEVER** calls the claim reward endpoint. It strictly stops at 100% progress so you can open Discord and claim your rewards manually, guaranteeing maximum safety.
* **Realistic X-Super-Properties**: Base64 encoded client fingerprint matching genuine Discord Desktop x64 builds (`os`, `browser`, `os_version`, `client_version`).
* **Dynamic CDN Build Number**: Dynamically scrapes the active `client_build_number` from Discord CDN at startup. Never sends outdated build tags.
* **Randomized Jitter**: Randomized interval delays between heartbeats and video progress events prevent pattern matching.

---

## 📂 Project Structure

```
discord-quest-phantom/
├── cmd/
│   └── completer/main.go       # Core Dual-Engine Orchestrator (Parallel Runner + Spoofer)
├── pkg/
│   ├── api/                    # REST Client, CDN Build Scraper, Video & Heartbeat
│   ├── captcha/                # Local LAN Captcha Web Portal (auto port resolution)
│   ├── config/                 # Token resolution & CLI parsing (interactive prompt)
│   ├── i18n/                   # Zero-dependency bilingual translation catalogs (EN, VI)
│   ├── scanner/                # Quest classification, lifecycle coordinator & expiry filter
│   ├── spoofer/                # Win32 GUI message pump & Linux process emulation
│   └── updater/                # GitHub release checker & update notification engine
├── docs/
│   ├── translations/           # Multilingual documentation (VI, ZH, KO, JA, HI)
│   ├── ARCHITECTURE.md         # In-depth architectural review & risk matrix
│   └── QUICKSTART.txt          # Minimal setup guide
├── scripts/
│   ├── diagnostics/            # Diagnostic and inspection scripts
│   ├── windows_helpers/        # Windows batch shortcuts
│   ├── build.ps1               # Windows PowerShell build script
│   └── build.sh                # Linux Bash build script
├── main.py                     # Standalone Python runner (Cross-platform)
├── run.cmd                     # Windows 1-click launcher
├── setup_token.cmd             # Windows token setup assistant
├── .gitignore                  # Security filter (blocks token leaks)
├── LICENSE                     # MIT License
└── README.md                   # Multilingual project documentation
```

---

## 📜 License

Distributed under the [MIT License](LICENSE). For educational and network automation research purposes only.
