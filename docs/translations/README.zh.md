<div align="center">

# 👻 Discord Quest Phantom
### *适用于 Windows 和 Linux (x64 / ARM64) 的跨平台全自动 Discord 任务完成器*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](../../LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages / 语言导航:**  
[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.vi.md) | [🇨🇳 简体中文](README.zh.md) | [🇰🇷 한국어](README.ko.md) | [🇯🇵 日本語](README.ja.md) | [🇮🇳 हिन्दी](README.hi.md)

---

</div>

> ### ⚠️ 免责声明 / DISCLAIMER
> **使用者自担风险 / USE AT YOUR OWN RISK.**
> 本软件仅供学习、教育和网络协议安全研究使用。自动化工具可能违反 Discord 服务条款 (ToS)。作者对使用本工具所导致的任何账户限制、警告或封禁不承担任何法律责任。

---

## 🎯 任务分类与自动化能力 (Quest Classification)

| 任务类型 | 任务代码 (Task Type) | 自动化运行机制 | 支持状态 |
|---|---|---|:---:|
| **1. 🖥️ 纯 PC 游戏** | `PLAY_ON_DESKTOP` (仅限 PC) | 自动向后台发送 20 秒间隔的心跳包，或使用 `-spoofer` 模拟真实 Win32 游戏进程。 | ✅ 100% 自动 |
| **2. 🎮 跨平台游戏** | `PLAY_ON_DESKTOP` + `XBOX` / `PS` | 客户端会弹出“选择平台”弹窗。**工具直接绕过客户端弹窗**，以 stream 心跳将时间累加至 PC 平台！ | ✅ 100% 自动 |
| **3. 📺 桌面视频** | `WATCH_VIDEO` | 桌面端视频任务。模拟自然播放时间戳提交进度，10~30 秒内完成。 | ✅ 100% 自动 |
| **4. 📱 移动端视频** | `WATCH_VIDEO_ON_MOBILE` | 针对手机端设计的视频任务。后端接口统一，**无需安装手机端应用即可完美完成**！ | ✅ 100% 自动 |
| **5. 🏆 游戏内成就** | `ACHIEVEMENT_IN_GAME` / `ACHIEVEMENT_IN_ACTIVITY` | 需要在真实游戏中达成目标（例如《无畏契约》击杀等），依赖游戏商服务器向 Discord 发送 Webhook。 | ⚠️ 自动跳过 (需玩家手动游玩) |

---

## 🔑 1. 获取 Discord Token 教程 (F12 开发者工具)

1. 在浏览器中打开 [https://discord.com/app](https://discord.com/app) 并登录。
2. 按 `F12` 打开开发者工具，切换到 **控制台 (Console)** 标签页。
3. 粘贴以下代码并按回车：
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. 终端会立即显示您的 Token 字符串，请复制该字符串（不要包含引号）。

---

## ⚡ 2. 安装与运行指南

### Windows 用户：
1. 从 [Releases](../../releases) 页面下载 `discord-quest-phantom-windows-amd64.zip`。
2. 解压到任意文件夹。
3. **双击运行 `discord-quest-phantom.exe`**：
   - 如果没有 `.token` 文件，程序会在黑框中提示直接粘贴 Token，保存后即可全自动运行！
   - 无需开启 Discord 客户端，无需安装 Python 或任何运行库。

### Linux / VPS / Armbian 用户 (24/7 后台运行)：
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.0.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "在此粘贴您的Token" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## 📜 许可证 (License)
基于 [MIT License](../../LICENSE) 协议发布。仅供个人技术研究使用。
