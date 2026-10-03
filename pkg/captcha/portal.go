package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/i18n"
)

// PortalResult holds the solved token or error.
type PortalResult struct {
	Token string
	Error error
}

// GetOutboundIP discovers the host's preferred LAN IP address (e.g. 192.168.1.200).
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	// Fallback to iterating interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Discord Quest Captcha Portal</title>
  <style>
    body {
      background-color: #1e1f22;
      color: #dbdee1;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      margin: 0;
      padding: 16px;
      box-sizing: border-box;
    }
    .card {
      background-color: #2b2d31;
      border-radius: 12px;
      box-shadow: 0 8px 24px rgba(0,0,0,0.4);
      max-width: 480px;
      width: 100%;
      padding: 24px;
      text-align: center;
    }
    .logo {
      font-size: 40px;
      margin-bottom: 8px;
    }
    h1 {
      font-size: 20px;
      color: #f2f3f5;
      margin: 0 0 8px 0;
    }
    .quest-badge {
      background: #5865f2;
      color: #fff;
      display: inline-block;
      padding: 6px 14px;
      border-radius: 16px;
      font-size: 14px;
      font-weight: 600;
      margin-bottom: 16px;
    }
    p {
      color: #949ba4;
      font-size: 14px;
      line-height: 1.5;
      margin: 0 0 20px 0;
    }
    .captcha-box {
      min-height: 80px;
      display: flex;
      justify-content: center;
      align-items: center;
      margin: 16px 0;
    }
    .success-box {
      display: none;
      background: #23a55a;
      color: #fff;
      padding: 16px;
      border-radius: 8px;
      font-weight: 600;
      margin-top: 16px;
    }
    .footer-links {
      margin-top: 20px;
      font-size: 12px;
      color: #80848e;
    }
    .footer-links a {
      color: #00a8fc;
      text-decoration: none;
    }
  </style>
  {{if eq .Service "turnstile"}}
  <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
  {{else}}
  <script src="https://js.hcaptcha.com/1/api.js" async defer></script>
  {{end}}
</head>
<body>
  <div class="card">
    <div class="logo">👻</div>
    <h1>{{.HeaderTitle}}</h1>
    <div class="quest-badge">{{.QuestTitle}}</div>
    <p>{{.Desc}}</p>

    <div class="captcha-box">
      {{if eq .Service "turnstile"}}
      <div class="cf-turnstile" data-sitekey="{{.Sitekey}}" data-callback="handleSolve"></div>
      {{else}}
      <div class="h-captcha" data-sitekey="{{.Sitekey}}" data-callback="handleSolve" data-rqdata="{{.Rqdata}}"></div>
      {{end}}
    </div>

    <div id="success" class="success-box">
      {{.SuccessMsg}}
    </div>

    <div class="footer-links">
      <a href="https://discord.com/quests/{{.QuestID}}" target="_blank">↗ {{.OpenInDiscord}}</a>
    </div>
  </div>

  <script>
    function handleSolve(token) {
      document.querySelector('.captcha-box').style.display = 'none';
      document.getElementById('success').style.display = 'block';
      fetch('/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: token })
      }).then(r => r.json()).then(data => {
        console.log('Submission accepted:', data);
      }).catch(err => {
        console.error('Submit error:', err);
      });
    }
  </script>
</body>
</html>`

type templateData struct {
	Service       string
	Sitekey       string
	Rqdata        string
	QuestID       string
	QuestTitle    string
	HeaderTitle   string
	Desc          string
	SuccessMsg    string
	OpenInDiscord string
}

// FindAvailablePort scans starting from startPort for an open TCP port to avoid conflicts.
func FindAvailablePort(startPort int) int {
	if startPort <= 0 {
		startPort = 8080
	}
	for p := startPort; p < startPort+50; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", p))
		if err == nil {
			ln.Close()
			return p
		}
	}
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err == nil {
		addr := ln.Addr().(*net.TCPAddr)
		ln.Close()
		return addr.Port
	}
	return startPort
}

// StartPortal launches a temporary local web server to present the captcha widget.
func StartPortal(ctx context.Context, port int, challenge *api.CaptchaRequiredError, questTitle string) (string, error) {
	port = FindAvailablePort(port)

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	service := challenge.CaptchaService
	if service == "" {
		service = "hcaptcha"
	}

	tmpl, err := template.New("captcha").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to compile portal template: %w", err)
	}

	var once sync.Once
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		isVI := i18n.GetLanguage() == i18n.LangVI
		headerTitle := "Discord Quest Verification"
		desc := "Discord requires a quick human verification to enroll in this quest. Solve below to proceed:"
		successMsg := "✅ Verified! Quest enrolled. You can return to your terminal."
		openInDiscord := "Open in Official Discord App"
		if isVI {
			headerTitle = "Xác Minh Discord Quest"
			desc = "Discord yêu cầu xác minh người thật để tham gia nhiệm vụ này. Vui lòng bấm xác minh bên dưới:"
			successMsg = "✅ Đã xác minh thành công! Tiến trình đang tiếp tục trên server."
			openInDiscord = "Mở trực tiếp trên Discord"
		}

		data := templateData{
			Service:       service,
			Sitekey:       challenge.CaptchaSitekey,
			Rqdata:        challenge.CaptchaRqdata,
			QuestID:       challenge.QuestID,
			QuestTitle:    questTitle,
			HeaderTitle:   headerTitle,
			Desc:          desc,
			SuccessMsg:    successMsg,
			OpenInDiscord: openInDiscord,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, data)
	})

	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Token) == "" {
			http.Error(w, "Invalid token payload", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})

		once.Do(func() {
			resultChan <- strings.TrimSpace(payload.Token)
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case token := <-resultChan:
		return token, nil
	}
}
