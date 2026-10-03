package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/config"
	"discord-quest-completer/pkg/i18n"
	"discord-quest-completer/pkg/scanner"
	"discord-quest-completer/pkg/spoofer"
)

// Version is injected during compilation via -ldflags="-X main.Version=..."
var Version = "1.1.0"

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
	fmt.Printf("[+] Language: %s (Auto-detected / Configured)\n", i18n.GetLanguage())
	fmt.Printf("[+] Loaded token: %s\n", config.MaskToken(cfg.Token))
	fmt.Printf("[+] Mode: %s\n", getModeString(cfg))
	fmt.Printf("[+] Polling Interval: %v | Auto-Enroll: %t\n", cfg.PollInterval, cfg.AutoAccept)
	fmt.Printf("[+] Region Sweep: %s | Captcha Portal: %t (Port: %d)\n", cfg.Region, cfg.EnablePortal, cfg.PortalPort)
	if cfg.QuestID != "" {
		fmt.Printf("[+] Targeted Quest ID: %s\n", cfg.QuestID)
	}
	if cfg.Duration > 0 {
		fmt.Printf("[+] Max Session Duration: %v\n", cfg.Duration)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Fetch Latest Client Build Number dynamically from Discord CDN
	fmt.Println(i18n.M().ScrapingCDN)
	buildNumber := api.FetchLatestBuildNumber()
	fmt.Printf(i18n.M().ActiveBuild, buildNumber)

	// 2. Initialize API Client
	apiClient := api.NewClient(cfg.Token, buildNumber)

	// 3. Authenticate & Retrieve Account Identity
	fmt.Println(i18n.M().Authenticating)
	user, err := apiClient.ValidateToken(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.M().AuthFailed, err)
		return
	}

	globalName := "None"
	if user.GlobalName != nil {
		globalName = *user.GlobalName
	}
	fmt.Print(i18n.M().AuthSuccess)
	fmt.Printf("    - User ID:     %s\n", user.ID)
	fmt.Printf("    - Username:    %s\n", user.Username)
	fmt.Printf("    - Global Name: %s\n", globalName)
	fmt.Printf("    - 2FA Enabled: %t\n", user.MFAEnabled)

	// 4. Discover Quests (with Multi-Region / Locale Sweep)
	var rawQuests []api.Quest
	if cfg.Region == "all" || strings.Contains(cfg.Region, ",") || cfg.Region != "" {
		fmt.Printf(i18n.M().RegionScanStart, cfg.Region)
		var regions []string
		if cfg.Region == "all" {
			regions = []string{"us", "jp", "vn"}
		} else {
			for _, r := range strings.Split(cfg.Region, ",") {
				regions = append(regions, strings.TrimSpace(r))
			}
		}
		rawQuests, err = apiClient.FetchQuestsMultiRegion(ctx, regions)
		if err == nil {
			fmt.Printf(i18n.M().RegionScanDone, len(rawQuests))
		}
	} else {
		fmt.Println(i18n.M().DiscoveringQuests)
		rawQuests, err = apiClient.FetchQuests(ctx)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.M().FailedRetrieveQuests, err)
		return
	}

	analyzedQuests := scanner.AnalyzeAll(rawQuests)
	fmt.Print(scanner.FormatQuestTable(analyzedQuests))

	// 5. Auto-Enrollment (if enabled and not dry-run)
	if cfg.AutoAccept && !cfg.DryRun {
		fmt.Println(i18n.M().AutoEnrollHeader)
		enrolled, err := scanner.AutoEnrollPending(ctx, apiClient, analyzedQuests, cfg.EnablePortal, cfg.PortalPort)
		if err != nil {
			fmt.Printf(i18n.M().AutoEnrollError, err)
		} else if enrolled > 0 {
			fmt.Printf(i18n.M().AutoEnrollSuccess, enrolled)
			fmt.Println(i18n.M().AutoEnrollRefresh)
			rawQuests, err = apiClient.FetchQuests(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, i18n.M().FailedRefreshQuests, err)
				return
			}
			analyzedQuests = scanner.AnalyzeAll(rawQuests)
			fmt.Print(scanner.FormatQuestTable(analyzedQuests))
		} else {
			fmt.Println(i18n.M().AutoEnrollAllEnrolled)
		}
	}

	// 6. Dry-run Mode
	if cfg.DryRun {
		fmt.Println(i18n.M().DryRunNotice)
		return
	}

	// 7. Execution Engine Branching
	if cfg.UseSpoofer {
		fmt.Printf("\n[*] %s\n", i18n.M().ModeSpoofer)
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
		fmt.Printf("\n[*] %s\n", i18n.M().ModeAPIRunner)
		if err := runAPIRunnerLoop(ctx, cfg, apiClient, analyzedQuests); err != nil {
			if err != context.Canceled && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "[-] API Runner error: %v\n", err)
			}
		}
	}

	fmt.Println(i18n.M().AllTasksFinished)
}

func printBanner() {
	fmt.Println("==================================================================")
	fmt.Printf("   Discord Quest Phantom (Unified Dual-Engine) v%s\n", Version)
	fmt.Printf("   %s\n", i18n.M().BannerSubtitle)
	fmt.Println("==================================================================")
}

func getModeString(cfg *config.Config) string {
	if cfg.UseSpoofer {
		return i18n.M().ModeSpoofer
	}
	return i18n.M().ModeAPIRunner
}

func promptExit(isError bool) {
	if isError {
		fmt.Println(i18n.M().PromptExitError)
	}
	fmt.Print(i18n.M().PromptExitKey)
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
		fmt.Println(i18n.M().NoEligibleQuests)
		return nil
	}

	fmt.Printf(i18n.M().FoundEligibleQuests, len(eligible))
	for i, q := range eligible {
		rem := float64(q.TargetSec) - q.CurrentSec
		if rem < 0 {
			rem = 0
		}
		remStr := fmt.Sprintf(i18n.M().RemainingQuestSec, rem, q.TargetSec)
		fmt.Printf("    %d. %-26s | %-24s | %s\n",
			i+1, q.Title, q.Category.Localized(), remStr)
	}

	for i, q := range eligible {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("\n==================================================================\n")
		fmt.Printf("%s\n", i18n.T(func(m i18n.Messages) string { return m.ProcessingQuestHeader }, i+1, len(eligible), q.Title, q.Category.Localized()))
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
			fmt.Printf(i18n.M().SkipUnsupportedTask, q.Title, q.TaskType)
			continue
		}

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			fmt.Printf(i18n.M().QuestErrorSkip, q.Title, err)
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

	fmt.Printf(i18n.M().VideoStart, q.Title, secondsDone, secondsNeeded)

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
				fmt.Printf(i18n.M().VideoProgressWarning, err)
			} else {
				secondsDone = targetTimestamp
				fmt.Printf(i18n.M().VideoProgress, secondsDone, secondsNeeded, (secondsDone/secondsNeeded)*100.0)
				if resp != nil && resp.CompletedAt != nil && *resp.CompletedAt != "" {
					fmt.Printf(i18n.M().VideoCompleted, q.Title)
					return nil
				}
			}
		}

		if timestamp >= secondsNeeded {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Final progress update
	_, _ = client.SendVideoProgress(ctx, q.ID, secondsNeeded)
	fmt.Printf(i18n.M().VideoCompleted, q.Title)
	return nil
}

func completeHeartbeatAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	pid := rand.Intn(28000) + 2000
	streamKey := fmt.Sprintf("call:0:%d", pid)

	rem := secondsNeeded - secondsDone
	fmt.Printf(i18n.M().HeartbeatStart, q.Title, rem/60.0)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			fmt.Printf(i18n.M().HeartbeatWarning, err)
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
			fmt.Printf(i18n.M().HeartbeatProgress,
				time.Now().Format("15:04:05"), secondsDone, secondsNeeded, pct)

			if (resp.CompletedAt != nil && *resp.CompletedAt != "") || secondsDone >= secondsNeeded {
				break
			}
		}

		time.Sleep(20 * time.Second)
	}

	// Final terminal heartbeat
	_, _ = client.SendHeartbeat(ctx, q.ID, streamKey, true)
	fmt.Printf(i18n.M().HeartbeatCompleted, q.Title)
	return nil
}

func completeActivityAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	streamKey := "call:0:1"

	fmt.Printf(i18n.M().ActivityStart, q.Title)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			fmt.Printf(i18n.M().ActivityWarning, err)
		} else if resp != nil {
			if resp.Progress != nil && resp.Progress["PLAY_ACTIVITY"].Value > 0 {
				secondsDone = resp.Progress["PLAY_ACTIVITY"].Value
			} else {
				secondsDone += 20.0
			}
			fmt.Printf(i18n.M().ActivityProgress, secondsDone, secondsNeeded)
			if (resp.CompletedAt != nil && *resp.CompletedAt != "") || secondsDone >= secondsNeeded {
				break
			}
		}

		time.Sleep(20 * time.Second)
	}

	_, _ = client.SendHeartbeat(ctx, q.ID, streamKey, true)
	fmt.Printf(i18n.M().ActivityCompleted, q.Title)
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
		fmt.Println(i18n.M().SpooferNoEligible)
		return nil
	}

	for i, eq := range eligible {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf(i18n.M().SpooferHeader, i+1, len(eligible), eq.Title)
		exeName, gameTitle, err := spooferEngine.ResolveExecutable(eq.AppID)
		if err != nil {
			fmt.Printf(i18n.M().SpooferNoExe, eq.AppID, err)
			continue
		}
		if gameTitle == "" {
			gameTitle = eq.Title
		}

		proc, err := spooferEngine.LaunchGame(ctx, eq.AppID, gameTitle, exeName)
		if err != nil {
			fmt.Printf(i18n.M().SpooferLaunchError, err)
			continue
		}
		fmt.Printf(i18n.M().SpooferProcCreated, proc.ExecutableName(), proc.PID())

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
			fmt.Printf(i18n.M().SpooferProgress, analyzed.CurrentSec, analyzed.TargetSec, analyzed.State)
			if analyzed.State == scanner.StateCompleted || analyzed.State == scanner.StateClaimed || analyzed.CurrentSec >= float64(analyzed.TargetSec) {
				fmt.Printf(i18n.M().SpooferCompleted, analyzed.Title)
				_ = proc.Stop()
				break
			}
		}
	}

	return nil
}
