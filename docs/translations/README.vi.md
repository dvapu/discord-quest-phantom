<div align="center">

# 👻 Discord Quest Phantom v1.1.0
### *Công cụ tự động hoàn thành Discord Quest đa nền tảng cho Windows & Linux (x64 / ARM64)*

[![Bản Ổn Định: v1.1.0](https://img.shields.io/badge/Bản%20Ổn%20Định-v1.1.0%20(Stable)-7289da?style=flat-square&logo=github)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.0)
[![Bản Nightly: v1.1.1](https://img.shields.io/badge/Bản%20Nightly-v1.1.1%20(Pre--release)-f39c12?style=flat-square&logo=github)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.1)
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
          ┌─────────────────────────────────────────────────────────────┐
          │                    DISCORD QUEST PHANTOM                    │
          │         v1.1.0 (Bản Ổn Định) · v1.1.1 (Bản Nightly)         │
          │      Quét Đa Vùng · Chạy Song Song · Cổng Giải Captcha      │
          └──────────────────────────────┬──────────────────────────────┘
                                         │
                          Bộ Điều Phối Hệ Thống Trung Tâm
                                         │
             ┌───────────────────────────┼───────────────────────────┐
             ▼                           ▼                           ▼
    [1. QUÉT ĐA VÙNG]          [2. CHẠY SONG SONG]        [3. CỔNG GIẢI CAPTCHA]
  - Quét US, JP và VN        - Chạy song song 2-5 quest - Tự dò IP LAN & đổi port
  - Giả lập Super-Properties - Tạo độ trễ jitter ngẫu   - 1 chạm trên iPhone/web
  - Mở khóa avatar bị ẩn     - 5 game chỉ mất 15 phút!  - Cứu cánh server SSH VPS
             │                           │                           │
             └───────────────────────────┴───────────────────────────┘
                                         │
                            2 Động Cơ Thực Thi Độc Lập
                                         │
                    ┌────────────────────┬────────────────────┐
                    ▼                                         ▼
    [ĐỘNG CƠ 1: AUTONOMOUS API RUNNER]         [ĐỘNG CƠ 2: OS PROCESS SPOOFER]
  - 100% Headless (Linux, VPS, Armbian)    - Dành cho người mở Discord Desktop app
  - KHÔNG CẦN mở app hay trình duyệt       - Tạo đồng thời nhiều tiến trình ảo
  - Tự gửi heartbeat & video tiến trình    - Discord Desktop nhận diện gửi Gateway
```

---

## 🚀 Các Tính Năng Đột Phá (v1.1.0 Bản Ổn Định · v1.1.1 Bản Nightly)

1. **🌍 Quét Đa Vùng Bộ Ba Vàng Tự Động (`--region all`)**:
   - Khắc phục triệt để tình trạng nhiệm vụ và khung viền Avatar bị ẩn tại Việt Nam.
   - Phantom tự động quét song song bộ ba vàng (`en-US`, `ja-JP`, `vi-VN`) giúp gom toàn bộ nhiệm vụ toàn cầu về danh sách, hoàn thành mọi nhiệm vụ độc quyền không cần cắm VPN!
2. **⚡ Chạy Song Song 2–5 Game/Nhiệm Vụ (`-concurrency 5`)**:
   - Thay vì chạy tuần tự tốn 15p x 5 = 75 phút, Phantom hỗ trợ **xử lý song song tối đa 5 quest cùng một lúc** với luồng heartbeat độc lập. Hoàn thành 5 game chỉ vỏn vẹn trong **15 phút**!
3. **📱 Cổng Giải Captcha Cục Bộ Qua WiFi/LAN (`-portal`)**:
   - Khi chạy trên server Linux/Armbian 24/7 (`192.168.1.200`) qua SSH không có màn hình:
   - Nếu Discord bắt xác minh người thật, Phantom tự mở cổng web nội bộ `http://<IP_MAY>:8080` (tự động đổi port nếu bị trùng).
   - Bạn chỉ cần dùng điện thoại iPhone đang kết nối WiFi nhà, bấm vào link xác minh 1 giây là server tự động tiếp tục cày nhiệm vụ!
4. **🔄 Tự Động Kiểm Tra Bản Cập Nhật**:
   - Kết nối về GitHub `dvapu/discord-quest-phantom` khi khởi động để báo khi có bản phát hành mới.

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
1. Mở trình duyệt web bất kỳ, truy cập vào [https://discord.com/app](https://discord.com/app) và đăng nhập vào tài khoản của bạn.
2. Nhấn phím `F12` (hoặc `Ctrl + Shift + I`) để mở Developer Tools → Chuyển sang tab **Console**.
   > 💡 **Cảnh báo chống dán mã lạ (Self-XSS)**: Nếu Discord hiện thông báo đỏ chặn thao tác Paste, hãy gõ dòng chữ `allow pasting` vào Console rồi nhấn `Enter` để mở khóa.
3. Dán dòng mã sau vào Console và nhấn `Enter`:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. Chuỗi Discord Token của bạn sẽ lập tức hiển thị ra màn hình (dạng `NTk5...` hoặc `mfa....`). Hãy copy chuỗi này (không copy dấu ngoặc kép).

### Cách 2: Qua Tab Network
1. Trong cửa sổ `F12`, chuyển sang tab **Network** và gõ `/api` vào ô Filter.
2. Nhấn vào bất kỳ request nào xuất hiện trong danh sách.
3. Ở khung bên phải, kéo xuống phần **Request Headers** và sao chép toàn bộ giá trị tại dòng **`authorization:`**.

---

## ⚡ 2. Hướng Dẫn Cài Đặt & Khởi Chạy

### Cho Người Dùng Windows:
1. Tải bản nén `discord-quest-phantom-windows-amd64.zip` từ mục [Releases (v1.1.0 Bản Ổn Định)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.0) hoặc [Bản Nightly (v1.1.1)](https://github.com/dvapu/discord-quest-phantom/releases/tag/v1.1.1).
2. Giải nén vào thư mục bất kỳ.
3. Nhấp đúp vào `run.cmd` hoặc `discord-quest-phantom.exe`:
   - Nếu chưa có file `.token`, chương trình sẽ tự động mở hộp thoại yêu cầu bạn dán token và tự động lưu.
   - Không cần cài Python, không phụ thuộc Discord Desktop!

### Cho Người Dùng Linux / VPS / Armbian / Raspberry Pi (Chạy Ngầm 24/7):
Chạy một lệnh duy nhất sau trên terminal Linux để tải bản build tĩnh (v1.1.0 Bản Ổn Định) phù hợp với kiến trúc CPU và chạy ngầm liên tục:
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.1.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "DAN_TOKEN_CUA_BAN_VAO_DAY" > .token && \
nohup ./discord-quest-phantom -daemon -poll 15m -portal=false > quest.log 2>&1 &
```
> 💡 *Lưu ý: Để trải nghiệm tính năng mới nhất từ bản Nightly, thay `v1.1.0` thành `v1.1.1` trong dòng lệnh tải về.*

#### Thiết Lập Dịch Vụ Systemd (Tự Khởi Động Khi Bật Máy):
Tạo file `/etc/systemd/system/discord-quest-phantom.service`:
```ini
[Unit]
Description=Discord Quest Phantom Autonomous Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/discord-quest-phantom
ExecStart=/opt/discord-quest-phantom/discord-quest-phantom -daemon -poll 15m -portal=false -concurrency 5 -region all -lang vi
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```
Sau đó bật và chạy dịch vụ:
```bash
systemctl daemon-reload && systemctl enable --now discord-quest-phantom
```

---

## ⚙️ Các Tham Số Dòng Lệnh (CLI Flags)

| Cờ Lệnh | Mặc Định | Ý Nghĩa / Chức Năng |
|---|---|---|
| `-region` | `all` | Khu vực quét: `all` (Bộ Ba Vàng: US+JP+VN), `us`, `jp`, `vn` |
| `-concurrency` | `5` | Số lượng quest chạy song song tối đa (1–5) |
| `-daemon` | `false` | Chạy nền 24/7 như dịch vụ hệ thống, quét định kỳ tự động |
| `-poll` | `60s` | Khoảng thời gian giãn cách giữa các lần quét khi bật daemon (vd: `15m`, `30m`) |
| `-portal` | `true` | Bật cổng giải Captcha qua web nội bộ (`false` để chạy nền âm thầm không mở port) |
| `-portal-port` | `8080` | Cổng web giải Captcha (tự động tăng nếu bị trùng) |
| `-spoofer` | `false` | Bật chế độ giả lập tiến trình game (cần mở Discord Desktop) |
| `-lang` | `auto` | Ngôn ngữ giao diện: `auto`, `vi`, `en` |
| `-dry-run` | `false` | Chỉ quét và hiển thị danh sách quest, không gửi lệnh cày |
| `-token` | `""` | Truyền trực tiếp Discord token qua tham số |

---

## 🛡️ Phân Tích An Toàn & Cơ Chế Chống Ban (Anti-Ban)

* **Không Bao Giờ Tự Động Nhận Thưởng (Zero Auto-Claim)**: Phantom chỉ làm nhiệm vụ đạt 100% tiến độ và dừng lại, để bạn vào Discord bấm **Claim Reward** bằng tay trong phần Kho Quà Tặng. Đây là nguyên tắc cốt lõi giúp 100% tài khoản an toàn tuyệt đối trước các đợt quét bot.
* **Định Danh Trình Duyệt Thực Tế (X-Super-Properties)**: Gói tin chứa fingerprint chuẩn của Discord Desktop x64 (`os`, `browser`, `os_version`, `client_version`).
* **Cào Build Number Động Từ Discord CDN**: Luôn cào build number mới nhất của Discord mỗi khi khởi động, không bao giờ dùng build lỗi thời.
* **Thời Gian Ngẫu Nhiên (Jitter Delay)**: Mọi khoảng thời gian gửi heartbeat và video timestamp đều được cộng thêm độ trễ ngẫu nhiên để tránh bị phát hiện theo chu kỳ máy móc bay pattern.

---

## 📂 Cấu Trúc Mã Nguồn

```
discord-quest-phantom/
├── cmd/
│   └── completer/main.go       # Bộ điều phối trung tâm (Parallel Runner + Spoofer)
├── pkg/
│   ├── api/                    # Discord REST Client, cào CDN Build, Video & Heartbeat
│   ├── captcha/                # Cổng web giải Captcha mạng LAN nội bộ
│   ├── config/                 # Quản lý cấu hình & nhận diện Token tự động
│   ├── i18n/                   # Hệ thống dịch thuật song ngữ (Việt - Anh) 0 dependency
│   ├── scanner/                # Phân loại trạng thái quest & điều phối chu kỳ
│   ├── spoofer/                # Giả lập tiến trình game Win32 & Linux /proc
│   └── updater/                # Kiểm tra cập nhật tự động từ GitHub Releases
├── docs/
│   ├── translations/           # Tài liệu đa ngôn ngữ (VI, ZH, KO, JA, HI)
│   ├── ARCHITECTURE.md         # Phân tích kiến trúc chuyên sâu & ma trận an toàn
│   └── QUICKSTART.txt          # Hướng dẫn nhanh rút gọn
├── scripts/
│   ├── diagnostics/            # Bộ kịch bản chẩn đoán dữ liệu Discord Quest
│   ├── windows_helpers/        # Phím tắt batch tiện ích cho Windows
│   ├── build.ps1               # Script build tự động trên Windows PowerShell
│   └── build.sh                # Script build tự động trên Linux Bash
├── main.py                     # Bản chạy Python độc lập (Đa nền tảng)
├── run.cmd                     # Kịch bản khởi chạy 1 chạm cho Windows
├── setup_token.cmd             # Hỗ trợ cài đặt token cho Windows
├── .gitignore                  # Bộ lọc an toàn (chống lộ token & mã độc)
├── LICENSE                     # Giấy phép MIT License
└── README.md                   # Tài liệu chính của dự án
```

---

## 📜 Giấy Phép (License)

Phát hành dưới giấy phép mã nguồn mở [MIT License](../../LICENSE). Phục vụ cho mục đích học tập và nghiên cứu giao thức mạng.
