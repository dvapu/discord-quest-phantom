# Anti-Profiling Design

This document explains what the tool does to look like **one ordinary Discord
desktop client on one ordinary machine**, what it deliberately does *not* do,
and which detection surfaces remain open.

The guiding rule: **consistency beats randomness**. A rotating fingerprint, a
new device id per run or a random user agent is itself an anomaly — real users
do not change hardware overnight. Every value below is therefore pinned to the
account once and only re-derived when the device genuinely changes.

---

## 1. One source of truth: `api.Fingerprint`

`pkg/api/fingerprint.go` holds every field Discord correlates:

| Field | Source | Note |
|---|---|---|
| `os`, `os_arch`, `app_arch` | `runtime.GOOS` / `GOARCH` | cannot disagree with the binary |
| `os_version` | Windows registry `CurrentBuildNumber` | validated as `10.0.x`; never a bare major |
| `client_version` | version catalog entry | paired with the rest of the release |
| `browser_version` (Electron) | same catalog entry | must appear as `Electron/x` in the UA |
| Chromium version | same catalog entry | must appear as `Chrome/x` in the UA |
| `native_build_number` | same catalog entry | desktop clients always send it |
| `browser_user_agent` | rebuilt from the three values above | never typed by hand |
| `system_locale` | OS UI locale (`i18n.OSLocale`) | machine property, never follows a proxy |
| `client_locale` | the account's own Discord locale (`/users/@me`) | per-request, may follow a sweep target |
| timezone | real OS IANA zone when available | falls back to a locale default |

`Fingerprint.Validate()` re-checks the result before the first request:
the UA must carry `discord/<client_version>` and `Electron/<browser_version>`,
the channel must be known, `native_build_number` must be positive and the
Windows build must be `10.0.x`. `Normalize()` fills holes from the shipped
default so a half-populated fingerprint can never be sent.

Headers are derived from that one struct in `Client.setHeaders`:
`User-Agent`, `X-Super-Properties`, `X-Discord-Locale`, `X-Discord-Timezone`
and `Accept-Language`. There is no second place where a UA or a super-property
blob is assembled.

### Client build numbers

Two different numbers exist and they are not interchangeable:

* `client_build_number` — the web/desktop asset build. The project scrapes the
  currently published value from the CDN at startup, because quest endpoints
  gate features on it. Sending an old asset build can hide quests.
* `native_build_number` — the native module build shipped **inside** one
  desktop release. It is catalog-pinned, because inventing it is a direct
  contradiction with `client_version`.

These two move independently in real installations as well (stable, PTB and
canary all differ), so pinning one and following the other is intentional, not
an oversight.

---

## 2. Version catalog (`pkg/identity`)

`knownClients` only accepts entries where **all four values come from the same
release**: `ClientVersion`, `Electron`, `Chrome`, `NativeBuild`.

On startup `Detect()`:

1. scans `%LOCALAPPDATA%\Discord\app-*` for an installed client version;
2. looks it up in the catalog;
3. if the installed build is not indexed, keeps the newest catalog entry and
   says so in the profile report — it never mixes a new client version with an
   old Electron/Chromium pair to "look fresh".

Adding an entry requires verifying the Electron/Chromium pair locally (the
versions are embedded in `Discord.exe`) and a known `native_build_number`.
An entry without the native build is worse than a slightly older consistent
one.

`-fingerprint legacy` restores the old fixed values for comparison or
troubleshooting; it keeps the OS locale real so the machine still matches the
headers it sends.

---

## 3. Per-account profile store

`-profile-store` (default `true`) keeps one profile per Discord account under
`<cache>/profiles/<sha256(account id)[:8]>.profile`.

* Content: the pinned fingerprint, system/client locale, timezone, creation
  date, home egress, provenance (`installed-catalog` / `catalog` / `fallback`)
  and a local device tag.
* At rest: sealed with **DPAPI** (`CryptProtectData`, current Windows user) and
  prefixed with `DQC1`. A copied file is unreadable for another user or on
  another machine, and the account id inside is checked on load, so profiles
  can never be swapped between accounts.
* Refresh policy: only when `client_version`, the user agent or the system
  locale changed on this machine. Nothing rotates per run.
* Fallbacks: an unreadable or foreign profile is rebuilt and reported instead
  of blocking the run.

`identity.Audit` prints every inconsistency before traffic starts: invalid
fingerprint fields, a profile created on another machine, an installed client
that is newer than the pinned one, weak at-rest protection, and a proxy egress
whose timezone differs from the home identity.

---

## 4. Token storage

`-token` / `DISCORD_TOKEN` still work. Otherwise the token is read from, in
order, an encrypted vault (`<cache>/token.bin`, DPAPI), then the legacy
plaintext candidates (`.token`, …).

A plaintext token that is found is sealed into the vault **once** and the
original file is left untouched — deleting user files is the user's decision.
On platforms without a per-user secret store the vault is skipped entirely
rather than writing a fake "encrypted" file.

---

## 5. Timing

`pkg/pacing` applies a bell-shaped (sum of two uniforms) jitter to every delay:
heartbeats, video progress, retries, enrollment spacing, per-lane start
staggering and the daemon poll interval. Default spread is ±15%, configurable
with `-jitter`; `-jitter 0` restores fixed delays for debugging.

The average delay is unchanged, so rate-limit behaviour stays the same while
the cadence stops being clockwork.

---

## 6. Process emulation (`-spoofer`)

The stub is built to look like an installed game rather than a temp payload:

* **Install-shaped path**: `<cache>\stubs\<Game Name>\<relative exe path>`,
  launched with that folder as the working directory.
* **Hard link instead of copy** where the volume allows it: one binary on disk,
  no duplicated 10 MB payload per game, and nothing new written to the game
  folder on every run. A copy is the fallback.
* **Window class per engine** (`UnityWndClass`, `UnrealWindow`, `SDL_app`, …)
  instead of one hardcoded class for every title.
* **Reserved working set** of 256–512 MB (`DQC_STUB_MEM_MB` overrides) and a
  small idle CPU load, so the process does not look like a zero-cost dummy.
* No network activity, no child processes, no console window.

---

## 7. Presence: the one conflicting signal a fingerprint cannot fix

A desktop quest is not advanced by REST calls. Discord's local client detects a
game window/process and reports it, and the backend credits playtime for that
game. That has two consequences the rest of this document cannot paper over:

* only **one** game can be reported at a time, and
* if the owner is really playing something else, the fake game is competing
  with a real presence.

`pkg/presence` therefore watches the machine (`-presence-guard`, default on):

| Situation | What the tool does |
|---|---|
| Discord desktop is not running | Skips desktop quests for the pass: nothing would report the game anyway |
| The quest's own game is already running | Keeps polling without launching a second copy, so real play counts |
| Another eligible quest's game is running | Defers this quest and gives the running game's quest its turn |
| A full-screen app is active (`SHQueryUserNotificationState`, or a borderless window covering the screen) | Stops the stub, waits, then resumes when the screen is free |
| Another Phantom instance already shows a game | Waits on a per-session named mutex; the OS releases it if that instance dies |

The same guard exists in the **default REST mode**, which has no stub at all:
desktop quests there are advanced by `POST /quests/{id}/heartbeat`. While a
full-screen app is active on a machine where Discord is running,
`completeHeartbeatAPI` holds those heartbeats instead of sending them, because
a heartbeat claiming a game while your own client reports something else is the
contradiction Discord flags. The heartbeat resumes (and its stall deadline is
reset) as soon as the screen is free.

Mode choice follows from that:

* **Discord runs on this machine** → use `-spoofer` for desktop quests. The
  local client really reports the game, so the account never needs a heartbeat
  that disagrees with its own presence. The REST runner still handles video and
  activity quests.
* **Headless / VPS / no local Discord** → the REST runner is the only option.
  It is not blocked by the guard, and it is also the weaker disguise, because
  progress arrives with no gateway session behind it. Do not share that token
  with a machine you actively play on.

Desktop quests also run strictly one after another. Launching five stubs at
once, which older builds did under `-concurrency`, tells Discord the account is
playing five games simultaneously - something no real machine can do. Progress
speed is unchanged in practice, because Discord credits one game at a time.

The guard is deliberately conservative: pausing costs time, while a presence
that contradicts reality is exactly the pattern that produces "unusual activity"
warnings.

The remaining operational rule is outside the code: run Phantom on the machine
where you actually use Discord, and do not share one token between a VPS and
your gaming PC. Two simultaneous sessions from two networks cannot be made to
look like one person.

---

## 8. What this cannot fix

These are the remaining detection surfaces, in decreasing order of strength.
None of them is addressed by fingerprint cosmetics:

1. **Multiple egress IPs in one session.** A regional sweep makes the same
   token appear from several countries within minutes. The tool now warns when
   more than one proxy egress is configured; this is the strongest signal it
   can emit. Prefer `-region direct` if you care more about quiet than reach.
2. **TLS fingerprint.** Go's `crypto/tls` ClientHello (JA3/JA4, extension order,
   ALPN, signature algorithms) does not match Chromium/BoringSSL, and HTTP/2
   SETTINGS differ as well. Matching it would require a uTLS-style stack; this
   project stays dependency-free, so the mismatch remains. It is an
   infrastructure-level signal, evaluated before headers are even parsed.
3. **No real Discord client.** The tool speaks REST only: no gateway session,
   no voice/RTC, no local IPC pipe, no hardware/device ids, no telemetry. An
   account that only ever appears over REST looks different from one that also
   holds a gateway connection.
4. **Progress realism.** Heartbeats and video events are still synthetic; the
   intervals are jittered, but the sequence (enroll → linear progress → stop at
   100%) is structurally simpler than a real client.
5. **Account and IP reputation.** Datacenter proxies, brand-new accounts and
   shared exit nodes are judged on reputation that no client-side change can
   influence.

Everything in sections 1-7 is implemented and covered by tests. The ranking in
section 8 is inference from how such systems are normally built, not a
measurement of Discord's internals.
