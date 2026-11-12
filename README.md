<div align="center">

# 👻 Discord Quest Phantom v1.1.3
### *Autonomous Cross-Platform Quest Completer & Badge Maximizer for Windows & Linux (x64 / ARM64)*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
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
                        ┌────────────────────────────────────────────────────────┐
                        │              DISCORD QUEST PHANTOM 👻 v1.1.3           │
                        │   Multi-Region Sweep · Parallel Engine · Badge Maximizer│
                        └───────────────────────────┬────────────────────────────┘
                                                    │
                                     Core System Orchestrator
                                                    │
                ┌───────────────────────────────────┼───────────────────────────────────┐
                ▼                                   ▼                                   ▼
    [1. REGIONAL EGRESS SWEEP]          [2. PARALLEL RUNNER (DEFAULT)]        [3. BADGE & CAPTCHA PORTAL]
  - Local IP first, then proxies     - Concurrently runs 2-5 quests       - Auto LAN IP & port detection
  - Unions quests across regions     - Staggered API request jitter       - 🎖️ Instant HypeSquad Badge Claim
  - Reports every failed region      - 5 games in 15 mins (not 75 mins!)  - Zero SSH/headless lockouts
                │                                   │                                   │
                └───────────────────────────────────┴───────────────────────────────────┘
                                                    │
                                 Dual-Engine Execution Layer
                                                    │
                ┌───────────────────────────────────┴───────────────────────────────────┐
                ▼                                                                       ▼
   [ENGINE 1: AUTONOMOUS API RUNNER]                                       [ENGINE 2: OS PROCESS SPOOFER]
  - 100% Headless (Windows, Linux, Armbian, VPS)                        - For users running Discord Desktop
  - NO browser or Discord app required                                  - Concurrently spawns dummy game stubs
  - Auto video playback & synthetic stream frames                       - Discord Desktop broadcasts Gateway status
```

---

## 🚀 Key Highlights & New Capabilities in v1.1.3

1. **🎖️ Discord Badge Maximizer & HypeSquad Integration (`-max-badge`, `-hypesquad`)**:
   - **Auto-Detect & Display**: Inspects user profile flags and displays active Discord badges (`[🛡️ HypeSquad Bravery]`, `[🛠️ Active Developer]`, `[🚀 Nitro]`, etc.).
   - **Instant HypeSquad Claiming**: Directly interacts with official Discord API (`POST /api/v9/hypesquad/online`) to enroll or switch your HypeSquad House (`bravery` / `1`, `brilliance` / `2`, `balance` / `3`, or `leave` / `0`).
   - **Auto-Maximization by Default**: When running the tool, if your account has no HypeSquad badge, it automatically claims the House of Bravery badge so you immediately maximize your account badges along with quest rewards!
   - **End-of-Session Summary**: Prints a clean Profile Badges & Quests summary at the end of each pass.
2. **🩹 Windows launcher actually launches (`run.cmd` / `setup_token.cmd`)**:
   - Fixed the "console window flashes and disappears" bug caused by Unix (LF) line endings in Windows batch files.
   - `.gitattributes` now pins these files to literal CRLF bytes, and CI **fails the build** if a bare LF ever sneaks back in, so the bug cannot return.
   - The launched program no longer needs Go or Python: the prebuilt `.exe` ships in the ZIP and `run.cmd` finds it automatically.
   - Errors now stay on screen with a clear reason plus the exit code.
2. **🧹 No more console flood from expired quests**:
   - Quests that ended while you weren't using the tool are detected (`HTTP 404 / 410 / 403 / 422`) and dropped after **one** line instead of being retried forever.
   - Progress loops now stop after 5 consecutive errors or 3 minutes without accepted progress, using bounded backoff.
   - The quest table collapses expired campaigns into a single summary line (use `-debug` to see them all).
   - Daemon mode no longer reprints an identical inventory every poll cycle.
3. **🌍 Regional Egress Sweep + never-miss discovery (`-region`, `-proxy`, `-proxy-file`)** — *corrected mechanism*:
   - Quest eligibility is decided by the **request IP** (the country Discord sees), not by `client_locale` / `system_locale`. We measured this on a live account: 22 different locale headers returned the exact same 61-quest set, while VPN users visibly gain other countries' quests. Locale headers only localise text.
   - Phantom sweeps the **local egress first**, then every regional proxy you configure, unions the responses by quest ID, and routes each quest's enroll/heartbeat/video calls back through the proxy that discovered it (its sealed traffic metadata belongs to that fetch).
   - Every target is reported as `OK`/`ERR` with quest/excluded counts and latency. Failed regions get a recovery retry, and anything still failing is printed instead of being silently skipped.
   - Requests are serialized with `Retry-After`-aware pacing: Discord allows only a 2–3 request burst on `/quests/@me` before answering `429`. A last-known-good snapshot bridges transient outages, and `quest_access_suspended_until` / `quest_enrollment_blocked_until` stop the tool before it earns a penalty.
   - Legacy `-locale-sweep trio|all` remains available for experiments, off by default, and explicitly does **not** change eligibility.
4. **⚡ Parallel / Concurrent Quest Execution (`-concurrency 5`)**:
   - Complete 2–5 quests concurrently in parallel lanes with staggered request jitter.
   - Finish 5 games in **15 minutes** instead of 75 minutes!
5. **📱 Local Captcha Web Portal (`-portal` & `-portal-port`)**:
   - Running headless on a 24/7 Linux server/Armbian (e.g. `192.168.1.200`) via SSH?
   - When Discord challenges an enrollment, Phantom spins up a local web portal on port `8080` (with auto port-conflict resolution).
   - Just tap the local LAN link on your phone (same WiFi) to solve the captcha in 3 seconds, and the headless daemon instantly resumes!
6. **🔄 Auto-Update Notification**:
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

### For Windows (no installation, no Go, no Python):
1. Download `discord-quest-phantom-windows-amd64.zip` from [Releases](https://github.com/dvapu/discord-quest-phantom/releases).
2. Extract the archive into any folder (not inside the ZIP preview — extract it first).
3. Double-click `run.cmd`:
   - If `.token` is missing, the launcher starts the token setup assistant, saves your token, and then starts the tool automatically.
   - The prebuilt `discord-quest-phantom.exe` is included, so nothing has to be compiled.
   - When the program exits, the window stays open and shows the exit code plus any error hint.

> **If the window still closes instantly**, run it from a terminal to see the reason:
> ```cmd
> cd /d "C:\path\to\discord-quest-phantom"
> run.cmd
> ```

You can also launch `discord-quest-phantom.exe` directly, but `run.cmd` is recommended because it keeps the window open.

### For Linux / VPS / Armbian / Raspberry Pi (24/7 Background Daemon):
Run this single command to download the standalone binary and start 24/7 background execution:
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.1.3/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "YOUR_DISCORD_TOKEN_HERE" > .token && \
nohup ./discord-quest-phantom -daemon -poll 15m -portal=false > quest.log 2>&1 &
```

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
> Running regional proxies? Append `-proxy-file /opt/discord-quest-phantom/proxies.txt` to `ExecStart` and place the proxy list next to the binary. Without proxies the daemon simply sweeps the server's own egress.

---

## ⚙️ CLI Options & Configuration Flags

| Flag | Default | Description |
|---|---|---|
| `-max-badge` | `true` | Maximize Discord badges (auto-enroll in HypeSquad House if unassigned) |
| `-hypesquad` | `""` | HypeSquad House badge: `bravery` (1), `brilliance` (2), `balance` (3), or `leave` (0) |
| `-region` | `all` | Sweep scope: `all` (local egress + every configured proxy), `direct`, `full` (also every locale), a comma list of proxy/country codes (`us,br,jp`) or Discord locales (`ja`) |
| `-proxy` | `""` | Regional egress entries: `CODE=URL` pairs separated by commas (`us=socks5://…,br=http://…`) |
| `-proxy-file` | `""` | File with one `CODE=proxyURL` per line (`#` comments allowed) |
| `-locale-sweep` | `off` | Locale header sweep for experiments: `off`, `trio`, `all` (does not change quest eligibility) |
| `-sweep-pacing` | `3s` | Minimum delay between quest discovery requests; auto-raised when Discord answers HTTP 429 |
| `-sweep-attempts` | `3` | Attempts per sweep target before it is reported as failed |
| `-concurrency` | `5` | Maximum concurrent quests running in parallel (1–5) |
| `-daemon` | `false` | Run continuously as background daemon/service, re-checking quests periodically |
| `-poll` | `60s` | Polling interval between quest checks when running in daemon mode (e.g. `15m`, `30m`) |
| `-portal` | `true` | Enable local LAN captcha web portal (`false` for completely silent headless background run) |
| `-portal-port` | `8080` | Port for captcha web portal (auto-increments if port is busy) |
| `-spoofer` | `false` | Enable Win32 OS process emulation (requires Discord Desktop running) |
| `-lang` | `auto` | Language override: `auto`, `en`, `vi` |
| `-dry-run` | `false` | Discover and display active quests without taking any actions |
| `-token` | `""` | Provide Discord token directly via command-line argument |
| `-debug` | `false` | Verbose output, including expired quests in the inventory table |
| `-keep-open` | `true` | Wait for Enter before exiting (prevents the window from closing) |

---

## 🌐 Regional Quest Discovery: How It Really Works

Discord decides **which quests an account is eligible for from the request IP** (the country it sees), while `X-Discord-Locale` / `Accept-Language` / `?locale=` only change the language of the response. Measured on a live account: probing 22 different locale headers returned the exact same quest set (61 quests, 98 excluded, identical IDs), while community reports consistently show that a VPN exit node in another country exposes that country's region-locked quests.

Phantom therefore no longer claims that locale spoofing reveals region-locked quests. The reliable way to see every region is to **query Discord from every region you care about and union the answers**:

```bash
# proxies.txt — one regional exit per line. http, https, socks5 and socks5h are accepted.
us=socks5://user:pass@us-exit.example:1080
ca=http://ca-exit.example:8080
br=socks5://br-exit.example:1080
jp=http://jp-exit.example:8080

# Sweep the local IP + every proxy, then run quests normally:
discord-quest-phantom -region all -proxy-file proxies.txt

# Or sweep one region while keeping the local view:
discord-quest-phantom -proxy "us=socks5://127.0.0.1:1080" -region us
```

What the sweep guarantees:

* **Local egress first.** The signed-in user's own view is always queried; a broken proxy can never hide quests.
* **Union + per-quest region binding.** Quests are deduplicated by ID, then every enroll/heartbeat/video call for a quest goes through the proxy that discovered it, with matching `X-Discord-Locale` / `X-Discord-Timezone` headers.
* **No silent misses.** Every target reports `OK`/`ERR`, quest count, excluded count and latency. Failed targets are retried once more after the healthy ones (rate-limit storms clear), and anything still failing is listed.
* **Rate-limit aware pacing.** Discord allows only a small burst on `GET /quests/@me` (observed: 2–3 requests, then `429` with `retry_after` 5–10s). Requests are serialized with a configurable floor (3s default), raised automatically on `429`, and retried using `Retry-After`.
* **Snapshot fallback.** A complete inventory is cached under the tool's cache directory. If a later pass is incomplete (a region returned 429/5xx), unexpired quests seen within the last 15 minutes are merged back instead of silently disappearing; restored quests are shown but not auto-enrolled because their sealed metadata is stale.
* **Account gate awareness.** `quest_access_suspended_until` stops all quest traffic; `quest_enrollment_blocked_until` skips auto-enrollment. The tool prints the timestamp instead of retrying into a penalty.

### Why are some quests still "excluded"?

`GET /quests/@me` also returns `excluded_quests` — quests Discord withheld from this account (region, platform, eligibility, or a replaced campaign). They are IDs only, not full configs. Phantom prints the count and keeps scanning; a proxy in the right region is what moves a quest from `excluded_quests` into `quests` and makes it completable.

### Legacy locale sweep

`-locale-sweep trio` / `-locale-sweep all` still appends language-only probes and now uses canonical Discord locale identifiers (`ja` instead of `ja-JP`, `vi` instead of `vi-VN`). It is off by default: it costs extra rate-limited requests and does not change eligibility.

---

## 🧯 Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Window flashes and closes instantly | Old build with LF line endings in the `.cmd` files | Use v1.1.2+ — `.gitattributes` + CI now enforce CRLF. Run `run.cmd` from a terminal to read the error. |
| "'nul' is not recognized", "'ho' is not recognized", "'xist' is not recognized" | cmd.exe mis-parsing LF batch files (truncated `echo` / `if exist`) | Same as above: upgrade to v1.1.2+ and re-extract the ZIP. |
| "you need to install Go and build it yourself" | Running the source tree instead of the release ZIP | Download `discord-quest-phantom-windows-amd64.zip` — it contains the prebuilt `.exe`. |
| Endless expired-quest or error text scrolling | Old build retrying dead campaigns forever | Fixed in v1.1.2: terminal errors are detected and loops are bounded. |
| `chcp`/garbled emoji in console | Legacy console codepage | The launcher switches to UTF-8 only while the app runs; Windows Terminal is recommended. |
| Fewer quests than a friend in another country; region-locked quests never appear | Quest eligibility follows the request IP, not the locale. Locale sweeps cannot change that. | Add a proxy in the target region via `-proxy` / `-proxy-file`; Phantom unions the local and proxy views. |
| `HTTP 429: rate limited` lines during the sweep | Discord allows only a tiny burst on `/quests/@me`. The engine paces and retries automatically. | Keep the default pacing or raise `-sweep-pacing`; sweep fewer proxies or poll less often. |
| `QUEST ACCESS SUSPENDED` / enrollment blocked message | Discord temporarily locked quests on the account (automation enforcement). | Wait until the printed timestamp. The tool deliberately stops instead of retrying into a longer penalty. |

---

## 🛡️ Anti-Ban Security Analysis

* **Never Auto-Claims Rewards**: Phantom **NEVER** calls the claim reward endpoint. It strictly stops at 100% progress so you can open Discord and claim your rewards manually, guaranteeing maximum safety.
* **Realistic X-Super-Properties**: Base64 encoded client fingerprint matching genuine Discord Desktop x64 builds (`os`, `browser`, `os_version`, `client_version`).
* **Dynamic CDN Build Number**: Dynamically scrapes the active `client_build_number` from Discord CDN at startup. Never sends outdated build tags.
* **Randomized Jitter**: Randomized interval delays between heartbeats and video progress events prevent pattern matching.
* **Bounded, paced discovery**: `/quests/@me` requests are serialized, `Retry-After` aware and never fan out per proxy, so the account stays far away from rate-limit escalation.
* **Quest gate awareness**: `quest_access_suspended_until` / `quest_enrollment_blocked_until` are read from the payload; the tool halts enrolling (or all quest traffic) until the gate expires instead of hammering a flagged account.
* **Per-region consistency**: every quest action is sent through the proxy that discovered the quest, with aligned locale/timezone headers and the current `client_build_number`.

---

## 📂 Project Structure

```
discord-quest-phantom/
├── cmd/
│   ├── completer/main.go       # Core orchestrator (regional sweep + Parallel Runner + Spoofer)
│   └── test_regions/main.go    # Regional/proxy discovery benchmark (passes, exclusives, JSON export)
├── pkg/
│   ├── api/                    # REST client, regional sweep engine, CDN scraper, video & heartbeat
│   ├── captcha/                # Local LAN Captcha Web Portal (auto port resolution)
│   ├── config/                 # Token resolution & CLI parsing (interactive prompt)
│   ├── i18n/                   # Zero-dependency bilingual translation catalogs (EN, VI)
│   ├── inventory/              # Last-known-good quest snapshot (never-miss fallback)
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
