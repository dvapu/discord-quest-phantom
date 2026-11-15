<div align="center">

# 👻 Discord Quest Phantom v1.2.0
### *Công cụ tự động hoàn thành Discord Quest & Tối ưu Huy hiệu đa nền tảng cho Windows & Linux (x64 / ARM64)*

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
                        │              DISCORD QUEST PHANTOM 👻 v1.2.0           │
                        │   Quét Đa Vùng · Chạy Song Song · Tối Ưu Huy Hiệu      │
                        └───────────────────────────┬────────────────────────────┘
                                                    │
                                     Bộ Điều Phối Trung Tâm
                                                    │
                ┌───────────────────────────────────┼───────────────────────────────────┐
                ▼                                   ▼                                   ▼
    [1. QUÉT THEO VÙNG IP]              [2. CHẠY SONG SONG (PARALLEL)]        [3. HUY HIỆU & CAPTCHA]
  - IP hiện tại trước, rồi proxy     - Chạy đồng thời 2-5 quest           - Tự dò IP LAN & tự đổi port
  - Hợp nhất quest nhiều vùng        - Tạo độ trễ ngẫu nhiên (Jitter)     - 🎖️ Mở khóa Huy hiệu HypeSquad tức thì
  - Báo rõ vùng nào bị lỗi           - Hoàn thành 5 game trong 15 phút!   - Cứu cánh server SSH Linux
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

## 🚀 Các Tính Năng Đột Phá Mới Trong Bản v1.2.0

1. **🪪 Danh tính thiết bị nhất quán cho từng tài khoản (`-fingerprint`, `-profile-store`)**:
   - OS build, bản Discord client / Electron / Chromium / native build, locale hệ thống và timezone đều lấy từ một nguồn duy nhất và được ghim theo tài khoản, thay vì dùng chung một fingerprint cứng cho mọi máy.
   - `system_locale` giữ đúng locale của hệ điều hành; `client_locale` / `X-Discord-Locale` theo tài khoản và vùng quét.
   - Hồ sơ và kho token được mã hoá bằng DPAPI trên Windows; file `.token` plaintext chỉ được chuyển vào vault một lần và vẫn để nguyên để bạn tự xoá.
   - Xem [`docs/ANTI_PROFILING.md`](https://github.com/dvapu/discord-quest-phantom/blob/main/docs/ANTI_PROFILING.md) để biết thiết kế và giới hạn còn lại.
2. **🛡️ Presence guard (`-presence-guard`, mặc định bật)**:
   - Nhiệm vụ game PC chạy **tuần tự**, không bao giờ mở nhiều stub cùng lúc.
   - Stub tạm dừng khi bạn đang chơi một game khác hoặc có app full-screen; nếu đúng game của nhiệm vụ đang chạy thật, tool để session thật tính tiến độ thay vì mở bản sao.
   - Mutex chống 2 instance cùng hiển thị game trên một máy.
   - Heartbeat REST của nhiệm vụ game tạm dừng khi có full-screen trên máy đang chạy Discord.
3. **⏱️ Jitter thời gian (`-jitter`, mặc định 0.15)**: nhiễu hình chuông ±15% cho heartbeat, retry, khoảng cách auto-enroll và poll.
4. **🎖️ Tối Ưu Huy Hiệu Profile & Tích Hợp HypeSquad (`-max-badge`, `-hypesquad`)**:
   - **Tự động nhận diện & hiển thị huy hiệu**: Đọc bitmask public_flags của tài khoản và hiển thị đầy đủ huy hiệu (`[🛡️ HypeSquad Bravery]`, `[🛠️ Active Developer]`, `[🚀 Nitro]`, v.v.).
   - **Mở khóa Huy hiệu HypeSquad tức thì**: Gọi trực tiếp Discord API (`POST /api/v9/hypesquad/online`) để tham gia hoặc đổi nhà HypeSquad (`bravery` / `1`, `brilliance` / `2`, `balance` / `3`, hoặc rời `leave` / `0`).
   - **Tự động Max Badge mặc định**: Khi chạy tool, nếu tài khoản chưa có huy hiệu HypeSquad, tool tự động nhận huy hiệu House of Bravery để bạn có trọn bộ huy hiệu đẹp nhất cùng quà nhiệm vụ!
   - **Bảng tổng kết cuối phiên**: Hiển thị bảng tổng kết huy hiệu hồ sơ và số lượng quest đã hoàn thành.
5. **🩹 File `run.cmd` / `setup_token.cmd` đã chạy được thật sự**:
   - Sửa triệt để lỗi "cửa sổ console chớp rồi tắt" do file batch lưu bằng chuẩn Unix (LF) thay vì Windows (CRLF).
   - `.gitattributes` giờ khóa cứng các file `.cmd` sang CRLF nguyên bản, và CI sẽ **fail build** nếu có dấu LF đơn lẻ lọt vào, nên lỗi này không thể tái diễn.
   - Máy nào tải về cũng chạy ngay: file `discord-quest-phantom.exe` đã đóng gói sẵn trong ZIP, **không cần cài Go hay Python**.
   - Khi có lỗi, chương trình dừng lại và in rõ mã lỗi kèm gợi ý, cửa sổ không tự đóng nữa.
6. **🧹 Hết tràn màn hình khi quest hết hạn**:
   - Các quest đã kết thúc trong lúc bạn không mở tool (lỗi `HTTP 404 / 410 / 403 / 422`) được nhận diện và bỏ qua chỉ với **một dòng** thay vì thử lại vô tận.
   - Vòng lặp tiến độ tự dừng sau 5 lỗi liên tiếp hoặc 3 phút không được Discord chấp nhận tiến độ, có giãn cách chờ hợp lý.
   - Bảng danh sách quest gom các quest hết hạn thành **một dòng tổng kết** (dùng `-debug` để xem đầy đủ).
   - Chế độ daemon không còn in lại y nguyên bảng cũ ở mỗi chu kỳ quét.
7. **🌍 Quét Quest Theo Vùng IP + Chống Bỏ Lỡ Nhiệm Vụ (`-region`, `-proxy`, `-proxy-file`)** — *cơ chế đã được sửa lại*:
   - Discord quyết định **bạn có thấy nhiệm vụ hay không dựa trên IP gửi request** (quốc gia Discord nhìn thấy), không phải theo `client_locale` / `system_locale`. Đo thực tế trên một tài khoản: 22 kiểu header locale khác nhau trả về **cùng một danh sách 61 quest** (98 quest bị loại, ID giống hệt nhau), trong khi người dùng VPN thấy rõ quest riêng của quốc gia khác. Header locale chỉ đổi ngôn ngữ văn bản.
   - Phantom quét **IP hiện tại trước**, sau đó tới từng proxy vùng bạn khai báo, hợp nhất kết quả theo ID quest và **gửi mọi request nhận/cày tiến độ của quest đó qua đúng proxy đã tìm ra nó** (metadata traffic_metadata_sealed gắn với vùng đó).
   - Mỗi vùng được báo `OK`/`ERR` kèm số quest, số quest bị loại và thời gian phản hồi. Vùng lỗi được thử lại một lượt sau khi các vùng khỏe mạnh xong; vùng vẫn lỗi sẽ được in thành cảnh báo thay vì bỏ qua âm thầm.
   - Request được xếp hàng tuần tự và tôn trọng `Retry-After`: Discord chỉ cho phép burst 2–3 request vào `/quests/@me` rồi trả `429`. Bản lưu inventory gần nhất giúp chống rỗng danh sách khi mạng chập chờn, còn `quest_access_suspended_until` / `quest_enrollment_blocked_until` khiến tool dừng lại trước khi bị phạt thêm.
   - `-locale-sweep trio|all` vẫn còn cho mục đích thử nghiệm, mặc định tắt và **không** thay đổi điều kiện nhận quest.
8. **⚡ Chạy Song Song Nhiều Nhiệm Vụ (`-concurrency 5`)**:
   - Tối đa 5 nhiệm vụ **video / API** chạy cùng lúc, mỗi luồng heartbeat độc lập, thời điểm bắt đầu có jitter.
- Nhiệm vụ **game PC là ngoại lệ**: luôn chạy tuần tự từng game một, vì client thật chỉ báo cáo một game tại một thời điểm.
9. **📱 Cổng Giải Captcha Cục Bộ Qua WiFi/LAN (`-portal`)**:
   - Khi chạy trên server Linux/Armbian 24/7 (`192.168.1.200`) qua SSH không có màn hình:
   - Nếu Discord bắt xác minh người thật, Phantom tự mở cổng web nội bộ `http://<IP_MAY>:8080` (tự động đổi port nếu bị trùng).
   - Bạn chỉ cần dùng điện thoại iPhone đang kết nối WiFi nhà, bấm vào link xác minh 1 giây là server tự động tiếp tục cày nhiệm vụ!
10. **🔄 Tự Động Kiểm Tra Bản Cập Nhật**:
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
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.2.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "DAN_TOKEN_CUA_BAN_VAO_DAY" > .token && \
nohup ./discord-quest-phantom -daemon -poll 15m -portal=false > quest.log 2>&1 &
```

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
> Nếu dùng proxy theo vùng: thêm `-proxy-file /opt/discord-quest-phantom/proxies.txt` vào `ExecStart` và đặt file proxy cạnh binary. Không có proxy thì daemon chỉ quét IP của chính server.

---

## ⚙️ Các Tham Số Dòng Lệnh (CLI Flags)

| Cờ Lệnh | Mặc Định | Ý Nghĩa / Chức Năng |
|---|---|---|
| `-max-badge` | `true` | Tối ưu huy hiệu Discord (tự động nhận huy hiệu HypeSquad nếu chưa có) |
| `-hypesquad` | `""` | Chọn nhà HypeSquad: `bravery` (1), `brilliance` (2), `balance` (3), hoặc `leave` (0) |
| `-region` | `all` | Phạm vi quét: `all` (IP hiện tại + mọi proxy đã khai báo), `direct`, `full` (thêm toàn bộ locale), danh sách mã proxy/quốc gia (`us,br,jp`) hoặc locale Discord (`ja`) |
| `-proxy` | `""` | Proxy thoát theo vùng: các cặp `MÃ=URL` cách nhau dấu phẩy (`us=socks5://…,br=http://…`) |
| `-proxy-file` | `""` | File mỗi dòng một proxy `MÃ=URL` (dòng bắt đầu bằng `#` là chú thích) |
| `-locale-sweep` | `off` | Quét header locale để thử nghiệm: `off`, `trio`, `all` (không đổi điều kiện nhận quest) |
| `-sweep-pacing` | `3s` | Giãn cách tối thiểu giữa các request quét quest; tự tăng khi gặp HTTP 429 |
| `-sweep-attempts` | `3` | Số lần thử cho mỗi vùng trước khi báo lỗi |
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

## 🌐 Quét Quest Theo Vùng: Cơ Chế Thực Tế

Discord quyết định **quest nào tài khoản được thấy dựa trên IP gửi request** (quốc gia Discord nhìn thấy), còn `X-Discord-Locale` / `Accept-Language` / `?locale=` chỉ đổi ngôn ngữ phản hồi. Đo trên tài khoản thật: 22 kiểu header locale trả về **cùng một danh sách quest** (61 quest, 98 quest bị loại, ID giống hệt), trong khi cộng đồng ghi nhận rõ ràng việc dùng VPN thoát ở quốc gia khác sẽ mở ra quest riêng của quốc gia đó.

Vì vậy Phantom **không còn** quảng cáo rằng đổi locale sẽ mở khóa quest vùng. Cách chắc chắn để thấy mọi vùng là **gửi request từ mỗi vùng bạn quan tâm rồi hợp nhất kết quả**:

```bash
# proxies.txt — mỗi dòng một đầu ra vùng. Hỗ trợ http, https, socks5 và socks5h.
us=socks5://user:pass@us-exit.example:1080
ca=http://ca-exit.example:8080
br=socks5://br-exit.example:1080
jp=http://jp-exit.example:8080

# Quét IP hiện tại + toàn bộ proxy rồi chạy quest bình thường:
discord-quest-phantom -region all -proxy-file proxies.txt

# Hoặc chỉ quét thêm một vùng (vẫn giữ góc nhìn IP hiện tại):
discord-quest-phantom -proxy "us=socks5://127.0.0.1:1080" -region us
```

Những gì cơ chế quét đảm bảo:

* **Luôn quét IP hiện tại trước.** Góc nhìn của chính tài khoản luôn được hỏi đầu tiên; proxy hỏng không thể làm mất quest.
* **Hợp nhất + gắn vùng cho từng quest.** Quest trùng ID được gộp; mọi request nhận/cày của quest được gửi qua đúng proxy đã tìm ra nó, kèm header `X-Discord-Locale` / `X-Discord-Timezone` tương ứng.
* **Không bỏ lỡ âm thầm.** Mỗi vùng báo `OK`/`ERR`, số quest, số quest bị loại và thời gian phản hồi. Vùng lỗi được thử lại sau khi các vùng khác xong, danh sách vẫn lỗi được in rõ.
* **Giãn cách theo rate limit.** Discord chỉ cho burst 2–3 request vào `GET /quests/@me` (thực đo: sau đó trả `429` với `retry_after` 5–10s). Request được xếp hàng với giãn cách cấu hình được (mặc định 3s), tự nâng khi gặp `429` và thử lại theo `Retry-After`.
* **Bản lưu dự phòng.** Inventory đầy đủ được lưu vào thư mục cache; nếu lượt sau bị thiếu vùng (429/5xx), các quest còn hạn thấy trong 15 phút gần nhất được gộp lại thay vì biến mất âm thầm. Quest khôi phục từ bản lưu vẫn hiển thị nhưng không tự nhận vì metadata đã cũ.
* **Tôn trọng khóa tài khoản.** `quest_access_suspended_until` dừng toàn bộ request quest; `quest_enrollment_blocked_until` bỏ qua bước tự nhận quest. Tool in mốc thời gian thay vì thử lại tới khi bị phạt nặng hơn.

### Vì sao vẫn có quest hiển thị là "bị loại"?

`GET /quests/@me` còn trả về `excluded_quests` — các quest Discord không cho tài khoản này tham gia (theo vùng, nền tảng, điều kiện, hoặc đã được thay thế bằng chiến dịch khác). Mục này chỉ có ID, không có config đầy đủ. Phantom in số lượng và tiếp tục quét; thêm proxy đúng vùng là cách đưa quest từ `excluded_quests` sang `quests` để cày được.

### Quét locale kiểu cũ

`-locale-sweep trio` / `-locale-sweep all` vẫn thêm các lượt chỉ đổi locale, nay dùng đúng mã locale của Discord (`ja` thay vì `ja-JP`, `vi` thay vì `vi-VN`). Mặc định tắt vì tốn thêm request bị rate limit mà không đổi điều kiện nhận quest.

---

## 🧯 Khắc Phục Sự Cố Nhanh

| Hiện tượng | Nguyên nhân | Cách xử lý |
|---|---|---|
| Thấy ít quest hơn bạn bè ở nước khác; quest vùng không bao giờ xuất hiện | Điều kiện quest theo IP gửi request, không theo locale | Thêm proxy đúng vùng bằng `-proxy` / `-proxy-file`; Phantom sẽ hợp nhất góc nhìn địa phương và proxy |
| Dòng `HTTP 429: rate limited` khi quét | Discord chỉ cho burst rất nhỏ ở `/quests/@me` | Giữ giãn cách mặc định hoặc tăng `-sweep-pacing`; giảm số proxy hoặc poll thưa hơn |
| Thông báo `QUEST ACCESS SUSPENDED` / bị chặn nhận quest | Discord tạm khóa quest trên tài khoản (do cơ chế chống automation) | Chờ tới mốc thời gian được in ra; tool cố ý dừng thay vì thử lại |

---

## 🛡️ Phân Tích An Toàn & Cơ Chế Chống Ban (Anti-Ban)

* **Không Bao Giờ Tự Động Nhận Thưởng (Zero Auto-Claim)**: Phantom chỉ làm nhiệm vụ đạt 100% tiến độ và dừng lại, để bạn vào Discord bấm **Claim Reward** bằng tay trong phần Kho Quà Tặng. Đây là nguyên tắc cốt lõi giúp 100% tài khoản an toàn tuyệt đối trước các đợt quét bot.
* **Định Danh Trình Duyệt Thực Tế (X-Super-Properties)**: Gói tin chứa fingerprint chuẩn của Discord Desktop x64 (`os`, `browser`, `os_version`, `client_version`).
* **Cào Build Number Động Từ Discord CDN**: Luôn cào build number mới nhất của Discord mỗi khi khởi động, không bao giờ dùng build lỗi thời.
* **Thời Gian Ngẫu Nhiên (Jitter Delay)**: Mọi khoảng thời gian gửi heartbeat và video timestamp đều được cộng thêm độ trễ ngẫu nhiên để tránh bị phát hiện theo chu kỳ máy móc bay pattern.
* **Quét có kiểm soát**: Request `/quests/@me` được xếp hàng tuần tự, tôn trọng `Retry-After`, không bắn song song qua nhiều proxy — tránh đẩy tài khoản vào vòng xoáy rate-limit.
* **Nhận biết khóa quest**: Đọc `quest_access_suspended_until` / `quest_enrollment_blocked_until` và dừng đúng lúc thay vì gửi tiếp vào tài khoản đã bị gắn cờ.
* **Nhất quán theo vùng**: Mọi thao tác quest đi qua đúng proxy đã tìm ra quest, header locale/timezone và `client_build_number` hiện hành.

---

## 📂 Cấu Trúc Mã Nguồn

```
discord-quest-phantom/
├── cmd/
│   ├── completer/main.go       # Bộ điều phối trung tâm (quét vùng + Parallel Runner + Spoofer)
│   └── test_regions/main.go    # Benchmark quét vùng/proxy (nhiều lượt, quest độc quyền, xuất JSON)
├── pkg/
│   ├── api/                    # REST client, động cơ quét vùng, cào CDN Build, Video & Heartbeat
│   ├── captcha/                # Cổng web giải Captcha mạng LAN nội bộ
│   ├── config/                 # Quản lý cấu hình & nhận diện Token tự động
│   ├── i18n/                   # Hệ thống dịch thuật song ngữ (Việt - Anh) 0 dependency
│   ├── inventory/              # Bản lưu inventory gần nhất (chống mất quest khi mạng lỗi)
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
