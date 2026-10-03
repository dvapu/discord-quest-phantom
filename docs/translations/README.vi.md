<div align="center">

# 👻 Discord Quest Phantom v1.1.0
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
                        ┌────────────────────────────────────────────────────────┐
                        │              DISCORD QUEST PHANTOM 👻 v1.1.0           │
                        │   Quét Đa Vùng · Chạy Song Song · Cổng Giải Captcha    │
                        └───────────────────────────┬────────────────────────────┘
                                                    │
                                     Bộ Điều Phối Trung Tâm
                                                    │
                ┌───────────────────────────────────┼───────────────────────────────────┐
                ▼                                   ▼                                   ▼
    [1. QUÉT ĐA VÙNG (MULTI-REGION)]    [2. CHẠY SONG SONG (PARALLEL)]        [3. CỔNG CAPTCHA NỘI BỘ]
  - Quét đồng thời US, JP và VN      - Chạy đồng thời 2-5 quest           - Tự dò IP LAN & tự đổi port
  - Giả lập X-Super-Properties       - Tạo độ trễ ngẫu nhiên (Jitter)     - 1 chạm trên điện thoại iPhone
  - Mở khóa khung avatar bị ẩn       - Hoàn thành 5 game trong 15 phút!   - Cứu cánh server SSH Linux
                │                                   │                                   │
                └───────────────────────────────────┴───────────────────────────────────┘
                                                    │
                                  2 Động Cơ Thực Thi Độc Lập
                                                    │
                ┌───────────────────────────────────┴───────────────────────────────────┐
                ▼                                                                       ▼
   [ĐỘNG CƠ 1: AUTONOMOUS API RUNNER - MẶC ĐỊNH]                           [ĐỘNG CƠ 2: WIN32 OS SPOOFER]
  - Chạy Headless (Windows, Linux, Armbian, VPS)                        - Dành cho người muốn mở Discord Desktop
  - KHÔNG CẦN mở app Discord hay trình duyệt                            - Tạo đồng thời nhiều game ảo song song
  - Tự động bỏ qua popup chọn nền tảng game                             - Discord Desktop nhận diện gửi Gateway
```

---

## 🚀 Các Tính Năng Đột Phá Mới Trong Bản v1.1.0

1. **🌍 Quét Đa Vùng Tự Động (`--region all`)**:
   - Khắc phục triệt để tình trạng nhiệm vụ và khung viền Avatar bị ẩn tại Việt Nam.
   - Phantom tự động giả lập định danh vùng (`en-US`, `ja-JP`, `vi-VN`) giúp gom toàn bộ nhiệm vụ toàn cầu về danh sách, hoàn thành mọi nhiệm vụ độc quyền không cần cắm VPN!
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
1. Tải bản nén `discord-quest-phantom-windows-amd64.zip` từ mục [Releases](https://github.com/dvapu/discord-quest-phantom/releases).
2. Giải nén vào thư mục bất kỳ.
3. Nhấp đúp vào `run.cmd` hoặc `discord-quest-phantom.exe`:
   - Nếu chưa có file `.token`, chương trình sẽ tự động mở hộp thoại yêu cầu bạn dán token và tự động lưu.
   - Không cần cài Python, không phụ thuộc Discord Desktop!

### Cho Người Dùng Linux / VPS / Armbian / Raspberry Pi (Chạy Ngầm 24/7):
Chạy một lệnh duy nhất sau trên terminal Linux để tải bản build tĩnh phù hợp với kiến trúc CPU và chạy ngầm liên tục:
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.1.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "DAN_TOKEN_CUA_BAN_VAO_DAY" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## ⚙️ Các Tham Số Dòng Lệnh (CLI Flags)

| Cờ Lệnh | Mặc Định | Ý Nghĩa / Chức Năng |
|---|---|---|
| `-region` | `all` | Khu vực quét: `all` (US+JP+VN), `us`, `jp`, `vn` |
| `-concurrency` | `5` | Số lượng quest chạy song song tối đa (1–5) |
| `-portal` | `true` | Bật cổng giải Captcha qua web nội bộ khi chạy headless |
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
