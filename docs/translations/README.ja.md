<div align="center">

# 👻 Discord Quest Phantom
### *Windows & Linux (x64 / ARM64) 向けクロスプラットフォーム Discord クエスト自動完了ツール*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](../../LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages / 言語選択:**  
[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.vi.md) | [🇨🇳 简体中文](README.zh.md) | [🇰🇷 한국어](README.ko.md) | [🇯🇵 日本語](README.ja.md) | [🇮🇳 हिन्दी](README.hi.md)

---

</div>

> ### ⚠️ 免責事項 / DISCLAIMER
> **自己責任でご利用ください / USE AT YOUR OWN RISK.**
> 本ツールは教育およびネットワークプロトコルの研究を目的として作成されています。自動化ツールの使用は Discord の利用規約 (ToS) に抵触する可能性があります。アカウントに対する警告や利用停止に関して、作者は一切の責任を負いません。

---

## 🎯 クエスト分類と自動化対応 (Quest Classification)

| クエスト分類 | タスク識別子 (Task Type) | 動作メカニズム | 自動化対応 |
|---|---|---|:---:|
| **1. 🖥️ 純 PC ゲーム** | `PLAY_ON_DESKTOP` (PC限定) | 20秒周期のハートビート送信または `-spoofer` によるWin32プロセス偽装。 | ✅ 100% 自動 |
| **2. 🎮 クロスプラットフォーム** | `PLAY_ON_DESKTOP` + コンソール | デスクトップアプリの「プラットフォーム選択ポップアップ」を**自動回避**し、PCプレイ時間として加算！ | ✅ 100% 自動 |
| **3. 📺 デスクトップ動画** | `WATCH_VIDEO` | 自然な再生タイムスタンプをシミュレートし、10〜30秒で完了。 | ✅ 100% 自動 |
| **4. 📱 モバイル動画** | `WATCH_VIDEO_ON_MOBILE` | モバイル端末不要で、バックエンドAPI経由で直接完了！ | ✅ 100% 自動 |
| **5. 🏆 ゲーム内実績** | `ACHIEVEMENT_IN_GAME` | ゲームサーバー（Riot, EA等）から直接Discordに通知されるため、**手動プレイが必要**。 | ⚠️ 自動スキップ |

---

## 🔑 1. Discord トークン取得方法 (F12 開発者ツール)

1. ブラウザで [https://discord.com/app](https://discord.com/app) にアクセスしてログインします。
2. `F12` キーを押して開発者ツールを開き、**Console** タブに切り替えます。
   > 💡 **コンソールの貼り付け保護 (Self-XSS)**: ブラウザのコンソールでコードの貼り付けがブロックされる場合は、まずコンソールに `allow pasting` と入力して `Enter` キーを押してロックを解除してください。
3. 以下のコードを貼り付けて `Enter` を押します:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. 表示されたトークン文字列をコピーします（引用符を除く）。

---

## ⚡ 2. 実行方法

### Windows:
1. [Releases](../../releases) から `discord-quest-phantom-windows-amd64.zip` をダウンロードして解凍します。
2. **`discord-quest-phantom.exe` をダブルクリックして起動します**。
   - `.token` ファイルがない場合、コンソール画面でトークンの入力を求められ、自動保存して実行されます！

### Linux / Armbian (24時間常駐):
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.0.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "ここにトークンを貼り付け" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## 📜 ライセンス (License)
[MIT License](../../LICENSE) に基づいて公開されています。
