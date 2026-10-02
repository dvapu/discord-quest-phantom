<div align="center">

# 👻 Discord Quest Phantom
### *Windows 및 Linux(x64 / ARM64)를 위한 크로스 플랫폼 Discord 퀘스트 자동 완료 프로그램*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](../../LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages / 언어 탐색:**  
[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.vi.md) | [🇨🇳 简体中文](README.zh.md) | [🇰🇷 한국어](README.ko.md) | [🇯🇵 日本語](README.ja.md) | [🇮🇳 हिन्दी](README.hi.md)

---

</div>

> ### ⚠️ 면책 조항 / DISCLAIMER
> **사용자의 전적인 책임 하에 사용하십시오 / USE AT YOUR OWN RISK.**
> 본 소프트웨어는 순수 교육, 학습 및 네트워크 프로토콜 연구 목적으로 제작되었습니다. 자동화 도구 사용은 Discord 이용 약관(ToS)을 위반할 수 있으며, 개발자는 이로 인한 계정 제재, 경고 등에 대해 어떠한 법적 책임도 지지 않습니다.

---

## 🎯 퀘스트 분류 및 자동화 범위 (Quest Classification)

| 퀘스트 유형 | 작업 코드 (Task Type) | 작동 원리 및 처리 방식 | 지원 여부 |
|---|---|---|:---:|
| **1. 🖥️ 순수 PC 게임** | `PLAY_ON_DESKTOP` (PC 전용) | 20초 주기 하트비트 패킷 자동 전송 또는 `-spoofer` 플래그로 실제 Win32 가상 프로세스 실행. | ✅ 100% 자동 |
| **2. 🎮 크로스 플랫폼 게임** | `PLAY_ON_DESKTOP` + `콘솔` | 디스코드 데스크톱 앱의 '플랫폼 선택 팝업'을 **자동 우회**하여 PC 플레이타임으로 즉시 적립! | ✅ 100% 자동 |
| **3. 📺 데스크톱 비디오** | `WATCH_VIDEO` | 동영상 시청 타임스탬프를 자연스럽게 시뮬레이션하여 10~30초 내에 100% 완료. | ✅ 100% 자동 |
| **4. 📱 모바일 비디오** | `WATCH_VIDEO_ON_MOBILE` | 모바일 전용 퀘스트도 동일한 API를 통해 **모바일 기기 없이 완벽하게 자동 처리**! | ✅ 100% 자동 |
| **5. 🏆 인게임 업적/활동** | `ACHIEVEMENT_IN_GAME` / `ACHIEVEMENT_IN_ACTIVITY` | 게임 개발사 서버(Riot, EA 등)에서 디스코드 서버로 웹훅을 보내야 하므로 **직접 게임을 플레이해야 함**. | ⚠️ 자동 건너뜀 (직접 플레이 필요) |

---

## 🔑 1. Discord 토큰 추출 방법 (F12 개발자 도구)

1. 브라우저에서 [https://discord.com/app](https://discord.com/app)에 접속하여 로그인합니다.
2. `F12`를 눌러 개발자 도구를 열고 **콘솔 (Console)** 탭으로 이동합니다.
   > 💡 **콘솔 붙여넣기 보안 경고 (Self-XSS)**: 브라우저 콘솔에서 코드 붙여넣기가 차단되는 경우, 콘솔 입력창에 먼저 `allow pasting`을 입력하고 `Enter`를 누르면 붙여넣기가 정상적으로 활성화됩니다.
3. 다음 코드를 붙여넣고 `Enter`를 누릅니다:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. 콘솔에 출력된 토큰 문자열을 복사합니다 (따옴표 제외).

---

## ⚡ 2. 다운로드 및 실행 방법

### Windows 사용자:
1. [Releases](../../releases) 페이지에서 `discord-quest-phantom-windows-amd64.zip`을 다운로드합니다.
2. 압축을 해제합니다.
3. **`discord-quest-phantom.exe`를 더블 클릭하여 실행합니다**:
   - `.token` 파일이 없는 경우, 콘솔 창에서 토큰 입력을 직접 요청하며 자동으로 저장 후 실행됩니다!
   - 디스코드 클라이언트를 켤 필요가 없으며, Python 설치도 전혀 필요 없습니다.

### Linux / VPS / Armbian (24/7 백그라운드 데몬):
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.0.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "여기에_토큰_입력" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## 📜 라이선스 (License)
본 프로젝트는 [MIT License](../../LICENSE)에 따라 제공됩니다.
