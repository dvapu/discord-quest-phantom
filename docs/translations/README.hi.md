<div align="center">

# 👻 Discord Quest Phantom
### *Windows और Linux (x64 / ARM64) के लिए स्वचालित Discord क्वेस्ट कम्प्लीटर*

[![GitHub Release](https://img.shields.io/github/v/release/dvapu/discord-quest-phantom?color=7289da&style=flat-square)](https://github.com/dvapu/discord-quest-phantom/releases)
[![Build & Release](https://img.shields.io/github/actions/workflow/status/dvapu/discord-quest-phantom/release.yml?style=flat-square)](https://github.com/dvapu/discord-quest-phantom/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](../../LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20(x64%2C%20ARM64)-blue.svg?style=flat-square)]()
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20Python-cyan.svg?style=flat-square)]()

---

**🌐 Languages / भाषा चुनें:**  
[🇺🇸 English](../../README.md) | [🇻🇳 Tiếng Việt](README.vi.md) | [🇨🇳 简体中文](README.zh.md) | [🇰🇷 한국어](README.ko.md) | [🇯🇵 日本語](README.ja.md) | [🇮🇳 हिन्दी](README.hi.md)

---

</div>

> ### ⚠️ अस्वीकरण / DISCLAIMER
> **अपने जोखिम पर उपयोग करें / USE AT YOUR OWN RISK.**
> यह सॉफ्टवेयर पूरी तरह से शैक्षणिक और नेटवर्क प्रोटोकॉल अनुसंधान उद्देश्यों के लिए बनाया गया है। स्वचालन उपकरण Discord की सेवा शर्तों का उल्लंघन कर सकते हैं। इसके उपयोग से उत्पन्न किसी भी खाते पर प्रतिबंध के लिए लेखक उत्तरदायी नहीं होगा।

---

## 🎯 क्वेस्ट वर्गीकरण (Quest Classification)

| क्वेस्ट प्रकार | टास्क कोड (Task Type) | कार्यप्रणाली | स्थिति |
|---|---|---|:---:|
| **1. 🖥️ शुद्ध PC गेम** | `PLAY_ON_DESKTOP` (केवल PC) | हर 20 सेकंड में स्वचालित रूप से हार्टबीट भेजता है या Win32 स्पूफर (`-spoofer`) चलाता है। | ✅ 100% स्वचालित |
| **2. 🎮 क्रॉस-प्लेटफ़ॉर्म गेम** | `PLAY_ON_DESKTOP` + कंसोल | डेस्कटॉप ऐप के पॉपअप को स्वचालित रूप से अनदेखा कर PC प्ले-टाइम जोड़ता है। | ✅ 100% स्वचालित |
| **3. 📺 डेस्कटॉप वीडियो** | `WATCH_VIDEO` | वीडियो प्रगति भेजकर 10-30 सेकंड में पूरा करता है। | ✅ 100% स्वचालित |
| **4. 📱 मोबाइल वीडियो** | `WATCH_VIDEO_ON_MOBILE` | मोबाइल डिवाइस की आवश्यकता के बिना बैकएंड API के माध्यम से पूरा करता है। | ✅ 100% स्वचालित |
| **5. 🏆 इन-गेम उपलब्धियां** | `ACHIEVEMENT_IN_GAME` | इसके लिए वास्तविक गेमप्ले की आवश्यकता होती है क्योंकि डेवलपर सर्वर सीधे Discord को रिपोर्ट करता है। | ⚠️ छोड़ दिया (हाथ से खेलें) |

---

## 🔑 1. Discord टोकन कैसे प्राप्त करें (F12 DevTools)

1. अपने ब्राउज़र में [https://discord.com/app](https://discord.com/app) खोलें और लॉगिन करें।
2. `F12` दबाएं और **Console** टैब पर जाएं।
3. यह कोड पेस्ट करें और `Enter` दबाएं:
   ```javascript
   (() => { let t = null; webpackChunkdiscord_app.push([[Math.random().toString()], {}, e => { if (e?.c) Object.values(e.c).forEach(m => { ['default', ...Object.keys(m?.exports || {})].forEach(k => { try { const v = m?.exports?.[k]?.getToken?.(); if (typeof v === 'string' && v.length > 20) t = v; } catch(err){} }); }); }]); return t; })()
   ```
4. प्राप्त टोकन को कॉपी करें (उद्धरण चिह्नों के बिना)।

---

## ⚡ 2. डाउनलोड और चलाना

### Windows उपयोगकर्ता:
1. [Releases](../../releases) से `discord-quest-phantom-windows-amd64.zip` डाउनलोड करें।
2. फ़ाइल को अनज़िप करें।
3. **`discord-quest-phantom.exe` पर डबल-क्लिक करें**:
   - यदि `.token` फ़ाइल नहीं है, तो प्रोग्राम सीधे आपसे टोकन पेस्ट करने के लिए कहेगा और स्वचालित रूप से सहेज लेगा!

### Linux / Armbian (24/7 बैकग्राउंड):
```bash
mkdir -p ~/discord-quest && cd ~/discord-quest && \
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
curl -sSL "https://github.com/dvapu/discord-quest-phantom/releases/download/v1.0.0/discord-quest-phantom-linux-${ARCH}.tar.gz" | tar -xz && \
chmod +x discord-quest-phantom && \
echo "अपना_टोकन_यहाँ_डालें" > .token && \
nohup ./discord-quest-phantom > quest.log 2>&1 &
```

---

## 📜 लाइसेंस (License)
[MIT License](../../LICENSE) के तहत जारी।
