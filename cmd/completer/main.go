package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/config"
	"discord-quest-completer/pkg/scanner"
	"discord-quest-completer/pkg/spoofer"
)

// Version is injected during compilation via -ldflags="-X main.Version=..."
var Version = "1.0.0"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
		promptExit(true)
		os.Exit(1)
	}

	// Internal Stub Mode Execution:
	// When launched by the spoofer engine to mimic a game process
	if cfg.RunStub {
		if err := spoofer.RunStub(cfg.StubTitle); err != nil {
			fmt.Fprintf(os.Stderr, "Stub execution error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	defer func() {
		if cfg.KeepOpen {
			promptExit(false)
		}
	}()

	// Main Orchestrator Execution
	printBanner()

	fmt.Printf("[+] OS: %s | Arch: %s | CPU Cores: %d\n", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	fmt.Printf("[+] Loaded token: %s\n", config.MaskToken(cfg.Token))
	fmt.Printf("[+] Mode: %s\n", getModeString(cfg))
	fmt.Printf("[+] Polling Interval: %v | Auto-Enroll: %t\n", cfg.PollInterval, cfg.AutoAccept)
	if cfg.QuestID != "" {
		fmt.Printf("[+] Targeted Quest ID: %s\n", cfg.QuestID)
	}
	if cfg.Duration > 0 {
		fmt.Printf("[+] Max Session Duration: %v\n", cfg.Duration)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Fetch Latest Client Build Number dynamically from Discord CDN
	fmt.Println("\n[*] Scraping latest Discord client_build_number from CDN...")
	buildNumber := api.FetchLatestBuildNumber()
	fmt.Printf("[+] Active Build Number: %d\n", buildNumber)

	// 2. Initialize API Client
	apiClient := api.NewClient(cfg.Token, buildNumber)

	// 3. Authenticate & Retrieve Account Identity
	fmt.Println("\n[*] Authenticating with Discord API (GET /api/v9/users/@me)...")
	user, err := apiClient.ValidateToken(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Authentication failed: %v\n", err)
		return
	}

	globalName := "None"
	if user.GlobalName != nil {
		globalName = *user.GlobalName
	}
	fmt.Printf("[+] Authentication Successful!\n")
	fmt.Printf("    - User ID:     %s\n", user.ID)
	fmt.Printf("    - Username:    %s\n", user.Username)
	fmt.Printf("    - Global Name: %s\n", globalName)
	fmt.Printf("    - 2FA Enabled: %t\n", user.MFAEnabled)

	// 4. Discover Quests
	fmt.Println("\n[*] Discovering Active Quests (GET /api/v9/quests/@me)...")
	rawQuests, err := apiClient.FetchQuests(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Failed to retrieve quests: %v\n", err)
		return
	}

	analyzedQuests := scanner.AnalyzeAll(rawQuests)
	fmt.Print(scanner.FormatQuestTable(analyzedQuests))

	// 5. Auto-Enrollment (if enabled and not dry-run)
	if cfg.AutoAccept && !cfg.DryRun {
		fmt.Println("\n[*] Checking for un-enrolled active quests...")
		enrolled, err := scanner.AutoEnrollPending(ctx, apiClient, analyzedQuests)
		if err != nil {
			fmt.Printf("[-] Auto-enroll encountered error: %v\n", err)
		} else if enrolled > 0 {
			fmt.Printf("[+] Successfully auto-enrolled in %d new quest(s)!\n", enrolled)
			fmt.Println("[*] Refreshing quest inventory after enrollment...")
			rawQuests, err = apiClient.FetchQuests(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[-] Failed to refresh quests after enrollment: %v\n", err)
				return
			}
			analyzedQuests = scanner.AnalyzeAll(rawQuests)
			fmt.Print(scanner.FormatQuestTable(analyzedQuests))
		} else {
			fmt.Println("[+] All eligible active quests are already enrolled.")
		}
	}

	// 6. Dry-run Mode
	if cfg.DryRun {
		fmt.Println("\n[!] Dry-run enabled: scan completed successfully. No actions taken.")
		return
	}

	// 7. Execution Engine Branching
	if cfg.UseSpoofer {
		fmt.Println("\n[*] Mode: OS Process Spoofer (Win32 dummy process emulation)")
		spooferEngine, err := spoofer.NewSpoofer(cfg.CacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Failed initializing spoofer engine: %v\n", err)
			return
		}
		if err := runSpooferLoop(ctx, cfg, apiClient, spooferEngine, analyzedQuests); err != nil {
			if err != context.Canceled && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "[-] Spoofer error: %v\n", err)
			}
		}
	} else {
		fmt.Println("\n[*] Mode: Autonomous API Runner (Heartbeat + Video-Progress API)")
		if err := runAPIRunnerLoop(ctx, cfg, apiClient, analyzedQuests); err != nil {
			if err != context.Canceled && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "[-] API Runner error: %v\n", err)
			}
		}
	}

	fmt.Println("\n[+] All tasks finished.")
}

func printBanner() {
	fmt.Println("==================================================================")
	fmt.Printf("   Discord Quest Phantom (Unified Dual-Engine) v%s\n", Version)
	fmt.Println("   Autonomous API Runner & OS Process Spoofer for Windows & Linux")
	fmt.Println("==================================================================")
}

func getModeString(cfg *config.Config) string {
	if cfg.UseSpoofer {
		return "OS Process Spoofer (-spoofer)"
	}
	return "Autonomous API Runner (Default, Discord-Independent)"
}

func promptExit(isError bool) {
	if isError {
		fmt.Println("\n[!] Đã xảy ra lỗi.")
	}
	fmt.Print("\n[?] Nhấn phím Enter để thoát...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

// ── Autonomous API Runner Loop ───────────────────────────────────────────────────

func runAPIRunnerLoop(ctx context.Context, cfg *config.Config, client *api.Client, quests []scanner.AnalyzedQuest) error {
	var eligible []scanner.AnalyzedQuest
	for _, q := range quests {
		if cfg.QuestID != "" && q.ID != cfg.QuestID {
			continue
		}
		if q.IsExpired {
			continue
		}
		if q.State != scanner.StateEnrolled {
			continue
		}
		if q.TargetSec > 0 && q.CurrentSec >= float64(q.TargetSec) {
			continue
		}
		eligible = append(eligible, q)
	}

	if len(eligible) == 0 {
		fmt.Println("[+] Không có quest nào cần hoàn thành lúc này (Tất cả quest đang hoạt động đã hoàn tất hoặc đã nhận thưởng)!")
		return nil
	}

	fmt.Printf("\n[+] Tìm thấy %d quest cần hoàn thành:\n", len(eligible))
	for i, q := range eligible {
		rem := float64(q.TargetSec) - q.CurrentSec
		if rem < 0 {
			rem = 0
		}
		fmt.Printf("    %d. %-26s | %-24s | Còn lại: ~%.0fs / %ds\n",
			i+1, q.Title, q.Category, rem, q.TargetSec)
	}

	for i, q := range eligible {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("\n==================================================================\n")
		fmt.Printf(">>> Đang xử lý [%d/%d]: %s (%s) <<<\n", i+1, len(eligible), q.Title, q.Category)
		fmt.Printf("==================================================================\n")

		var err error
		switch q.TaskType {
		case "WATCH_VIDEO", "WATCH_VIDEO_ON_MOBILE":
			err = completeVideoAPI(ctx, client, q)
		case "PLAY_ON_DESKTOP", "STREAM_ON_DESKTOP":
			err = completeHeartbeatAPI(ctx, client, q)
		case "PLAY_ACTIVITY":
			err = completeActivityAPI(ctx, client, q)
		default:
			fmt.Printf("[!] Bỏ qua quest %s: Loại task %s chưa được hỗ trợ tự động.\n", q.Title, q.TaskType)
			continue
		}

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			fmt.Printf("[-] Lỗi khi xử lý %s: %v. Chuyển sang quest tiếp theo...\n", q.Title, err)
			continue
		}

		if i < len(eligible)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return nil
}

func completeVideoAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec

	fmt.Printf("[🎬 Video] Bắt đầu phát video: %s (Tiến độ: %.0f/%.0fs)\n", q.Title, secondsDone, secondsNeeded)

	var enrolledTime time.Time
	if q.Quest.UserStatus != nil && q.Quest.UserStatus.EnrolledAt != nil && *q.Quest.UserStatus.EnrolledAt != "" {
		if t, err := time.Parse(time.RFC3339, *q.Quest.UserStatus.EnrolledAt); err == nil {
			enrolledTime = t
		}
	}
	if enrolledTime.IsZero() {
		enrolledTime = time.Now().Add(-10 * time.Minute)
	}

	const maxFuture = 10.0
	const speed = 7.0

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		maxAllowed := time.Since(enrolledTime).Seconds() + maxFuture
		diff := maxAllowed - secondsDone
		timestamp := secondsDone + speed

		if diff >= speed || timestamp >= secondsNeeded {
			targetTimestamp := timestamp
			if targetTimestamp > secondsNeeded {
				targetTimestamp = secondsNeeded
			}
			jitter := float64(rand.Intn(100)) / 100.0
			sendVal := targetTimestamp + jitter
			if sendVal > secondsNeeded {
				sendVal = secondsNeeded
			}

			resp, err := client.SendVideoProgress(ctx, q.ID, sendVal)
			if err != nil {
				fmt.Printf("    [-] Cảnh báo video progress: %v\n", err)
			} else {
				secondsDone = targetTimestamp
				fmt.Printf("    [🎬] Tiến độ: %.0f/%.0fs (%.0f%%)\n", secondsDone, secondsNeeded, (secondsDone/secondsNeeded)*100.0)
				if resp != nil && resp.CompletedAt != nil && *resp.CompletedAt != "" {
					fmt.Printf("[✅] Hoàn thành xuất sắc: %s\n", q.Title)
					return nil
				}
			}
		}

		if timestamp >= secondsNeeded {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Chốt tiến độ cuối cùng
	_, _ = client.SendVideoProgress(ctx, q.ID, secondsNeeded)
	fmt.Printf("[✅] Hoàn thành video quest: %s\n", q.Title)
	return nil
}

func completeHeartbeatAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	pid := rand.Intn(28000) + 2000
	streamKey := fmt.Sprintf("call:0:%d", pid)

	rem := secondsNeeded - secondsDone
	fmt.Printf("[🎮 Game Heartbeat] Bắt đầu gửi tín hiệu chơi: %s (~%.0f phút còn lại)\n", q.Title, rem/60.0)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			fmt.Printf("    [-] Cảnh báo heartbeat: %v\n", err)
		} else if resp != nil {
			if resp.Progress != nil {
				if prog, ok := resp.Progress[q.TaskType]; ok {
					secondsDone = prog.Value
				} else if prog, ok := resp.Progress["PLAY_ON_DESKTOP"]; ok {
					secondsDone = prog.Value
				}
			} else {
				secondsDone += 20.0
			}
			pct := (secondsDone / secondsNeeded) * 100.0
			if pct > 100.0 {
				pct = 100.0
			}
			fmt.Printf("    [🎮] [%s] Tiến độ: %.0f/%.0fs (%.1f%%)\n",
				time.Now().Format("15:04:05"), secondsDone, secondsNeeded, pct)

			if (resp.CompletedAt != nil && *resp.CompletedAt != "") || secondsDone >= secondsNeeded {
				break
			}
		}

		time.Sleep(20 * time.Second)
	}

	// Gửi heartbeat terminal để chốt kết thúc session
	_, _ = client.SendHeartbeat(ctx, q.ID, streamKey, true)
	fmt.Printf("[✅] Hoàn thành Play Quest: %s\n", q.Title)
	return nil
}

func completeActivityAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	streamKey := "call:0:1"

	fmt.Printf("[🕹️ Activity] Bắt đầu Activity: %s\n", q.Title)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			fmt.Printf("    [-] Cảnh báo activity heartbeat: %v\n", err)
		} else if resp != nil {
			if resp.Progress != nil && resp.Progress["PLAY_ACTIVITY"].Value > 0 {
				secondsDone = resp.Progress["PLAY_ACTIVITY"].Value
			} else {
				secondsDone += 20.0
			}
			fmt.Printf("    [🕹️] Tiến độ Activity: %.0f/%.0fs\n", secondsDone, secondsNeeded)
			if (resp.CompletedAt != nil && *resp.CompletedAt != "") || secondsDone >= secondsNeeded {
				break
			}
		}

		time.Sleep(20 * time.Second)
	}

	_, _ = client.SendHeartbeat(ctx, q.ID, streamKey, true)
	fmt.Printf("[✅] Hoàn thành Activity: %s\n", q.Title)
	return nil
}

// ── OS Process Spoofer Loop ──────────────────────────────────────────────────────

func runSpooferLoop(ctx context.Context, cfg *config.Config, apiClient *api.Client, spooferEngine spoofer.Spoofer, quests []scanner.AnalyzedQuest) error {
	var eligible []scanner.AnalyzedQuest
	for _, q := range quests {
		if cfg.QuestID != "" && q.ID != cfg.QuestID {
			continue
		}
		if q.TaskType != "PLAY_ON_DESKTOP" {
			continue
		}
		if q.IsExpired || q.State != scanner.StateEnrolled {
			continue
		}
		if q.TargetSec > 0 && q.CurrentSec >= float64(q.TargetSec) {
			continue
		}
		eligible = append(eligible, q)
	}

	if len(eligible) == 0 {
		fmt.Println("[+] Không có quest PLAY_ON_DESKTOP nào đủ điều kiện để chạy Spoofer.")
		return nil
	}

	for i, eq := range eligible {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("\n>>> Spoofer Quest [%d/%d]: %s <<<\n", i+1, len(eligible), eq.Title)
		exeName, gameTitle, err := spooferEngine.ResolveExecutable(eq.AppID)
		if err != nil {
			fmt.Printf("[-] Không tìm thấy executable cho AppID %s: %v\n", eq.AppID, err)
			continue
		}
		if gameTitle == "" {
			gameTitle = eq.Title
		}

		proc, err := spooferEngine.LaunchGame(ctx, eq.AppID, gameTitle, exeName)
		if err != nil {
			fmt.Printf("[-] Lỗi khởi động dummy game: %v\n", err)
			continue
		}
		fmt.Printf("[+] Đã tạo tiến trình game ảo: %s (PID: %d)\n", proc.ExecutableName(), proc.PID())

		// Poll loop
		for {
			select {
			case <-ctx.Done():
				_ = proc.Stop()
				return ctx.Err()
			case <-time.After(cfg.PollInterval):
			}

			rawQuests, err := apiClient.FetchQuests(ctx)
			if err != nil {
				continue
			}
			var curQ *api.Quest
			for _, rq := range rawQuests {
				if rq.ID == eq.ID {
					curQ = &rq
					break
				}
			}
			if curQ == nil {
				continue
			}
			analyzed := scanner.AnalyzeQuest(*curQ)
			fmt.Printf("[★] Tiến độ: %.0fs / %ds (State: %s)\n", analyzed.CurrentSec, analyzed.TargetSec, analyzed.State)
			if analyzed.State == scanner.StateCompleted || analyzed.State == scanner.StateClaimed || analyzed.CurrentSec >= float64(analyzed.TargetSec) {
				fmt.Printf("[✔] Hoàn thành Spoofer Quest: %s!\n", analyzed.Title)
				_ = proc.Stop()
				break
			}
		}
	}

	return nil
}
