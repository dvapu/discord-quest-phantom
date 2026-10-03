package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/i18n"
)

// ActiveChallenge represents a pending captcha challenge for a specific quest.
type ActiveChallenge struct {
	QuestID        string `json:"quest_id"`
	QuestTitle     string `json:"quest_title"`
	CaptchaService string `json:"service"`
	Sitekey        string `json:"sitekey"`
	Rqdata         string `json:"rqdata"`
	Rqtoken        string `json:"rqtoken"`
	CreatedAt      time.Time `json:"created_at"`
}

// WebPortal manages the persistent local web dashboard and captcha server.
type WebPortal struct {
	port       int
	lanIP      string
	server     *http.Server
	mu         sync.Mutex
	challenge  *ActiveChallenge
	solveChan  chan string
	isTestMode bool
}

var (
	defaultPortal     *WebPortal
	defaultPortalOnce sync.Once
)

// GetOutboundIP discovers the host's preferred LAN IP address (e.g. 192.168.1.200).
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

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

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Discord Quest Phantom — Web Portal</title>
  <style>
    * { box-sizing: border-box; }
    body {
      background-color: #1e1f22;
      color: #dbdee1;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      margin: 0;
      padding: 20px;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
    }
    .container {
      background-color: #2b2d31;
      border-radius: 16px;
      box-shadow: 0 12px 32px rgba(0,0,0,0.5);
      max-width: 520px;
      width: 100%;
      padding: 32px 28px;
      text-align: center;
    }
    .logo {
      font-size: 48px;
      margin-bottom: 6px;
      animation: float 3s ease-in-out infinite;
    }
    @keyframes float {
      0%, 100% { transform: translateY(0); }
      50% { transform: translateY(-6px); }
    }
    h1 {
      font-size: 22px;
      color: #f2f3f5;
      margin: 0 0 6px 0;
      font-weight: 700;
    }
    .subtitle {
      color: #949ba4;
      font-size: 13px;
      margin-bottom: 20px;
    }
    .badge-bar {
      display: flex;
      justify-content: center;
      gap: 10px;
      margin-bottom: 24px;
      flex-wrap: wrap;
    }
    .badge {
      font-size: 12px;
      padding: 5px 12px;
      border-radius: 20px;
      font-weight: 600;
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }
    .badge-online { background: #23a55a22; color: #23a55a; border: 1px solid #23a55a44; }
    .badge-host { background: #5865f222; color: #5865f2; border: 1px solid #5865f244; }
    .badge-ip { background: #f0b23222; color: #f0b232; border: 1px solid #f0b23244; }

    .card {
      background: #1e1f22;
      border-radius: 12px;
      padding: 20px;
      margin-bottom: 20px;
      border: 1px solid #35373c;
    }
    .quest-title {
      color: #5865f2;
      font-size: 16px;
      font-weight: 700;
      margin: 8px 0;
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
      margin-top: 14px;
    }
    .btn {
      background: #5865f2;
      color: #fff;
      border: none;
      padding: 10px 20px;
      border-radius: 8px;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
      text-decoration: none;
      display: inline-block;
      transition: background 0.2s;
    }
    .btn:hover { background: #4752c4; }
    .btn-secondary {
      background: #35373c;
      color: #dbdee1;
    }
    .btn-secondary:hover { background: #3f4147; }
    .footer {
      font-size: 11px;
      color: #80848e;
      margin-top: 24px;
      line-height: 1.6;
    }
  </style>
  {{if .Sitekey}}
    {{if eq .Service "turnstile"}}
    <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
    {{else}}
    <script src="https://js.hcaptcha.com/1/api.js" async defer></script>
    {{end}}
  {{end}}
</head>
<body>
  <div class="container">
    <div class="logo">👻</div>
    <h1>Discord Quest Phantom</h1>
    <div class="subtitle">Autonomous Headless Multi-Region Daemon</div>

    <div class="badge-bar">
      <div class="badge badge-online">🟢 <span>ONLINE</span></div>
      <div class="badge badge-host">💻 <span>{{.HostOS}}</span></div>
      <div class="badge badge-ip">🌐 <span>{{.LANAddress}}</span></div>
    </div>

    {{if .HasChallenge}}
    <div class="card" id="challenge-card">
      <div style="font-size: 28px; margin-bottom: 4px;">⚠️</div>
      <h2 style="font-size: 16px; margin: 0; color: #f2f3f5;">{{.ChallengeHeader}}</h2>
      <div class="quest-title">{{.QuestTitle}}</div>
      <p style="font-size: 13px; color: #949ba4; margin: 6px 0 16px 0;">{{.ChallengeDesc}}</p>

      <div class="captcha-box" id="captcha-container">
        {{if eq .Service "turnstile"}}
        <div class="cf-turnstile" data-sitekey="{{.Sitekey}}" data-callback="handleSolve"></div>
        {{else}}
        <div class="h-captcha" data-sitekey="{{.Sitekey}}" data-callback="handleSolve" data-rqdata="{{.Rqdata}}"></div>
        {{end}}
      </div>

      <div id="success" class="success-box">
        ✅ {{.SuccessMsg}}
      </div>
    </div>
    {{else}}
    <div class="card">
      <div style="font-size: 32px; margin-bottom: 8px;">✨</div>
      <h3 style="margin: 0 0 6px 0; color: #f2f3f5; font-size: 16px;">{{.NoPendingTitle}}</h3>
      <p style="font-size: 13px; color: #949ba4; margin: 0 0 16px 0;">
        {{.NoPendingDesc}}
      </p>
      <div style="display: flex; gap: 10px; justify-content: center;">
        <a href="/test" class="btn btn-secondary">🧪 {{.TestWidgetBtn}}</a>
        <a href="https://discord.com/channels/@me" target="_blank" class="btn">🚀 {{.OpenDiscordBtn}}</a>
      </div>
    </div>
    {{end}}

    <div class="footer">
      <div>Phantom Local Captcha Portal &bull; Auto-resolves region-locked quests</div>
      <div>Zero browser dependencies &bull; Parallel Multi-Region Runner</div>
    </div>
  </div>

  <script>
    function handleSolve(token) {
      const container = document.getElementById('captcha-container');
      if (container) container.style.display = 'none';
      const success = document.getElementById('success');
      if (success) success.style.display = 'block';

      fetch('/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: token })
      }).then(r => r.json()).then(data => {
        console.log('Submission accepted:', data);
        setTimeout(() => {
          window.location.href = '/';
        }, 3000);
      }).catch(err => {
        console.error('Submit error:', err);
      });
    }

    // Auto-poll for new challenges if currently idle
    {{if not .HasChallenge}}
    setInterval(() => {
      fetch('/api/status')
        .then(r => r.json())
        .then(status => {
          if (status.has_challenge) {
            window.location.reload();
          }
        }).catch(() => {});
    }, 3000);
    {{end}}
  </script>
</body>
</html>`

type templateViewData struct {
	HostOS          string
	LANAddress      string
	HasChallenge    bool
	Service         string
	Sitekey         string
	Rqdata          string
	QuestID         string
	QuestTitle      string
	ChallengeHeader string
	ChallengeDesc   string
	SuccessMsg      string
	NoPendingTitle  string
	NoPendingDesc   string
	TestWidgetBtn   string
	OpenDiscordBtn  string
}

// StartBackgroundPortal initiates a persistent local web portal on an open port.
func StartBackgroundPortal(port int) (*WebPortal, error) {
	var initErr error
	defaultPortalOnce.Do(func() {
		actualPort := FindAvailablePort(port)
		ip := GetOutboundIP()

		portal := &WebPortal{
			port:      actualPort,
			lanIP:     ip,
			solveChan: make(chan string, 1),
		}

		mux := http.NewServeMux()

		// 1. Dashboard View
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			portal.renderDashboard(w, r, false)
		})

		// 2. Test Captcha View
		mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			portal.renderDashboard(w, r, true)
		})

		// 3. Captcha Submission Endpoint
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

			token := strings.TrimSpace(payload.Token)
			portal.mu.Lock()
			select {
			case portal.solveChan <- token:
			default:
			}
			portal.challenge = nil
			portal.mu.Unlock()
		})

		// 4. Status API Endpoint for live dashboard polling
		mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
			portal.mu.Lock()
			defer portal.mu.Unlock()

			hasChal := portal.challenge != nil
			res := map[string]interface{}{
				"status":        "online",
				"host_os":       runtime.GOOS + "/" + runtime.GOARCH,
				"lan_ip":        portal.lanIP,
				"port":          portal.port,
				"has_challenge": hasChal,
			}
			if hasChal {
				res["quest_id"] = portal.challenge.QuestID
				res["quest_title"] = portal.challenge.QuestTitle
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
		})

		server := &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", actualPort),
			Handler: mux,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("[!] Captcha Web Portal error: %v\n", err)
			}
		}()

		portal.server = server
		defaultPortal = portal
	})

	return defaultPortal, initErr
}

func (p *WebPortal) renderDashboard(w http.ResponseWriter, r *http.Request, testMode bool) {
	p.mu.Lock()
	chal := p.challenge
	p.mu.Unlock()

	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	isVI := i18n.GetLanguage() == i18n.LangVI
	data := templateViewData{
		HostOS:     fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
		LANAddress: fmt.Sprintf("%s:%d", p.lanIP, p.port),
	}

	if isVI {
		data.ChallengeHeader = "Xác Minh Người Thật Bắt Buộc"
		data.ChallengeDesc = "Discord yêu cầu giải Captcha để nhận nhiệm vụ này. Vui lòng bấm xác minh bên dưới:"
		data.SuccessMsg = "Đã xác minh thành công! Phantom đang tự động nhận quest trên server."
		data.NoPendingTitle = "Không Có Captcha Nào Đang Chờ"
		data.NoPendingDesc = "Phantom đang chạy ngầm, tự động cày và hoàn thành mọi nhiệm vụ bình thường."
		data.TestWidgetBtn = "Thử Nghiệm Widget"
		data.OpenDiscordBtn = "Mở Discord"
	} else {
		data.ChallengeHeader = "Human Verification Required"
		data.ChallengeDesc = "Discord requires a quick captcha solution to enroll in this quest. Solve below:"
		data.SuccessMsg = "Verified successfully! Phantom is proceeding with quest enrollment."
		data.NoPendingTitle = "No Captchas Pending"
		data.NoPendingDesc = "Phantom daemon is actively monitoring and completing quests in the background."
		data.TestWidgetBtn = "Test Widget"
		data.OpenDiscordBtn = "Open Discord"
	}

	if testMode {
		data.HasChallenge = true
		data.Service = "turnstile"
		data.Sitekey = "1x00000000000000000000AA" // Cloudflare Turnstile Always-Pass test key
		data.QuestTitle = "🧪 Captcha Portal Test Mode"
		data.QuestID = "test-portal-001"
	} else if chal != nil {
		data.HasChallenge = true
		data.Service = chal.CaptchaService
		if data.Service == "" {
			data.Service = "hcaptcha"
		}
		data.Sitekey = chal.Sitekey
		data.Rqdata = chal.Rqdata
		data.QuestID = chal.QuestID
		data.QuestTitle = chal.QuestTitle
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// SetChallenge registers an active captcha challenge to be solved by the user.
func (p *WebPortal) SetChallenge(challenge *api.CaptchaRequiredError, questTitle string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.challenge = &ActiveChallenge{
		QuestID:        challenge.QuestID,
		QuestTitle:     questTitle,
		CaptchaService: challenge.CaptchaService,
		Sitekey:        challenge.CaptchaSitekey,
		Rqdata:         challenge.CaptchaRqdata,
		Rqtoken:        challenge.CaptchaRqtoken,
		CreatedAt:      time.Now(),
	}
}

// ClearChallenge removes the current challenge after completion or timeout.
func (p *WebPortal) ClearChallenge() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.challenge = nil
}

// WaitForSolution blocks until the user solves the captcha or context cancels.
func (p *WebPortal) WaitForSolution(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		p.ClearChallenge()
		return "", ctx.Err()
	case token := <-p.solveChan:
		return token, nil
	}
}

// URL returns the full reachable local web address.
func (p *WebPortal) URL() string {
	return fmt.Sprintf("http://%s:%d", p.lanIP, p.port)
}

// Port returns the listening port.
func (p *WebPortal) Port() int {
	return p.port
}

// StartPortal provides backwards-compatible single-call portal launch.
func StartPortal(ctx context.Context, port int, challenge *api.CaptchaRequiredError, questTitle string) (string, error) {
	portal, err := StartBackgroundPortal(port)
	if err != nil {
		return "", err
	}

	portal.SetChallenge(challenge, questTitle)
	return portal.WaitForSolution(ctx)
}
