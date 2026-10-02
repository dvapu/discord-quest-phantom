<div align="center">

# 👻 Discord Quest Phantom
### *Công cụ tự động hoàn thành Discord Quest đa nền tảng cho Windows & Linux (x64 / ARM64)*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](../../LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages / Ngôn ngữ:**  
[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.vi.md) | [🇨🇳 简体中文](README.zh.md) | [🇰🇷 한국어](README.ko.md) | [🇯🇵 日本語](README.ja.md) | [🇮🇳 हिन्दी](README.hi.md)

---

</div>

> ### ⚠️ TUYÊN BỐ MIỄN TRỪ TRÁCH NHIỆM / DISCLAIMER
> **TÔI SẼ KHÔNG CHỊU TRÁCH NHIỆM / USE AT YOUR OWN RISK.**
> Phần mềm này được phát triển hoàn toàn phục vụ cho mục đích học tập, giáo dục và nghiên cứu giao thức mạng. Việc sử dụng công cụ tự động hóa có thể vi phạm Điều khoản Dịch vụ (ToS) của Discord. Tác giả không chịu bất kỳ trách nhiệm nào đối với bất kỳ rủi ro, cảnh báo, giới hạn tính năng hay việc tài khoản bị đình chỉ/ban. Bạn hoàn toàn tự chịu trách nhiệm khi quyết định sử dụng.

---

```
                        ┌──────────────────────────────────────────────┐
                        │          DISCORD QUEST PHANTOM 👻            │
                        │    Auto-Scan · Auto-Enroll · Zero-Ban        │
                        └──────────────────────┬───────────────────────┘
                                               │
                        Lựa Chọn Chế Độ Hoạt Động Độc Lập
                                               │
                ┌──────────────────────────────┴──────────────────────────────┐
                ▼                                                             ▼
     [CHẾ ĐỘ 1: AUTONOMOUS RUNNER - MẶC ĐỊNH]                      [CHẾ ĐỘ 2: WIN32 OS SPOOFER]
  - Chạy BẤT KỲ ĐÂU (Windows, Linux, Armbian, VPS)              - Dành cho người muốn mở Discord Desktop
  - KHÔNG CẦN mở ứng dụng Discord Desktop                       - Tạo cửa sổ tiến trình giả lập game thật
  - TỰ ĐỘNG BỎ QUA POPUP "Chọn nền tảng để bắt đầu"             - Discord Desktop tự nhận diện & gửi Gateway
  - Cào build number mới nhất, gửi heartbeat có jitter          - Kích hoạt bằng cờ: -spoofer
```

---

## 🎯 Phân Loại 5 Nhóm Nhiệm Vụ Discord (Quest Classification)

| Nhóm Nhiệm Vụ | Ký hiệu (Task Type) | Bản chất kỹ thuật & Cơ chế tự động xử lý | Khả năng tự động |
|---|---|---|:---:|
| **1. 🖥️ PC Thuần** | `PLAY_ON_DESKTOP` (Chỉ PC) | Game chỉ có trên PC. Tool tự động gửi heartbeat định kỳ 20s qua API Runner hoặc giả lập process Win32 (`-spoofer`). | ✅ 100% Tự động |
| **2. 🎮 Đa Nền Tảng** | `PLAY_ON_DESKTOP` + `XBOX` / `PLAYSTATION` | Discord Desktop hiện popup bắt chọn nền tảng. **Tool tự động vượt qua popup**, gửi heartbeat stream `call:0:<pid>` ghi thẳng giờ chơi PC! | ✅ 100% Tự động |
| **3. 📺 Video Máy Tính** | `WATCH_VIDEO` | Video dành cho Desktop. Tool gửi mốc thời gian tự nhiên qua `/quests/{id}/video-progress`, hoàn thành trong 10–30s. | ✅ 100% Tự động |
| **4. 📱 Video Điện Thoại** | `WATCH_VIDEO_ON_MOBILE` | Video nhắm vào Mobile app. Máy chủ Discord dùng chung API `/video-progress`. **Tool mô phỏng chuẩn xác tiến trình video**, hoàn thành 100% không cần chạm vào điện thoại! | ✅ 100% Tự động |
| **5. 🏆 Thành Tựu Trong Game** | `ACHIEVEMENT_IN_GAME` / `ACHIEVEMENT_IN_ACTIVITY` | Nhiệm vụ yêu cầu thành tích thật trong game (ví dụ: `VALORANT Aces`, `Battlefield 6`). Cần webhook từ máy chủ Riot/EA về Discord nên **bắt buộc người dùng phải tự chơi game**. | ⚠️ Bỏ qua (Cần tự chơi) |

---

## 🔑 1. Hướng Dẫn Lấy Discord Token Bằng F12 DevTools

### Cách 1: Qua Tab Console (1 Dòng Lệnh Duy Nhất — Nhanh Nhất)
1. Mở trình duyệt truy cập [https://discord.com/app](https://discord.com/app) và đăng nhập tài khoản của bạn.
2. Nhấn `F12` (hoặc `Ctrl + Shift + I`) → Chuyển sang tab **Console**.
3. Dán đoạn mã sau vào và nhấn `Enter`:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. Chuỗi token Discord của bạn sẽ xuất hiện ngay lập tức (dạng `NTk5...` hoặc `mfa....`). Hãy copy chuỗi này (không lấy dấu ngoặc kép).

### Cách 2: Qua Tab Network
1. Trong DevTools (`F12`), chuyển sang tab **Network** và gõ filter: `/api`.
2. Bấm vào bất kỳ request nào (`messages`, `users`, `science`).
3. Nhìn cột bên phải kéo xuống mục **Request Headers** → Copy toàn bộ chuỗi ở dòng `authorization:`.

---

## ⚡ 2. Hướng Dẫn Cài Đặt & Sử Dụng

### Dành cho Windows:
1. Tải file zip `discord-quest-phantom-windows-amd64.zip` từ trang [Releases](../../releases).
2. Giải nén vào một thư mục bất kỳ.
3. **Nhấp đúp chuột chạy file `discord-quest-phantom.exe`**:
   - Nếu chưa có file `.token`, chương trình sẽ hỏi bạn dán token trực tiếp trên màn hình console, tự động lưu file `.token` và chạy ngay!
   - Không cần cài đặt Python, không cần cài thư viện, không cần mở ứng dụng Discord Desktop!

### Dành cho Linux / VPS / Armbian (24/7 Daemon):
Chạy 1 dòng lệnh duy nhất để tải bản binary độc lập và chạy nền 24/7:
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.0.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "PASTE_TOKEN_CỦA_BẠN_VÀO_ĐÂY" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## 🛡️ Phân Tích An Toàn / Anti-Ban Risk Analysis

* **Tuyệt Đối Không Tự Nhận Thưởng (Claim Reward)**: Tool chỉ hoàn thành thanh tiến độ thời gian (100%). Người dùng tự mở Discord bấm nhận thưởng để đảm bảo an toàn tối đa cho tài khoản.
* **X-Super-Properties Fingerprinting**: Giả lập trọn vẹn Client Electron x64 thật (`os`, `browser`, `os_version`, `client_version`).
* **Cào Build Number Động**: Tự động kết nối CDN Discord lấy `client_build_number` mới nhất ngay khi chạy.
* **Jitter Ngẫu Nhiên**: Thêm biến thiên thời gian ngẫu nhiên giữa các lần gửi heartbeat và video progress, chống quét mẫu lặp tự động.

---

## 📜 Giấy Phép (License)
Phát hành theo giấy phép [MIT License](../../LICENSE). Dành cho mục đích giáo dục và nghiên cứu tự động hóa giao thức.
