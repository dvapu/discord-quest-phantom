package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/captcha"
	"discord-quest-completer/pkg/config"
	"discord-quest-completer/pkg/i18n"
	"discord-quest-completer/pkg/identity"
	"discord-quest-completer/pkg/inventory"
	"discord-quest-completer/pkg/pacing"
	"discord-quest-completer/pkg/presence"
	"discord-quest-completer/pkg/scanner"
	"discord-quest-completer/pkg/spoofer"
	"discord-quest-completer/pkg/updater"
)

// Version is injected during compilation via -ldflags="-X main.Version=..."
var Version = "1.2.0"

// lastInventorySignature remembers the previous scan fingerprint so daemon
// mode can stay silent while nothing changes.
var lastInventorySignature string

// sweepTargets is the regional egress plan built once from the CLI flags.
var sweepTargets []api.RegionTarget

// snapshotSaveWarned keeps a broken cache file from spamming the console.
var snapshotSaveWarned bool

// timingJitter is the configured fraction applied to every inter-request delay.
var timingJitter = 0.15

// presenceGuardEnabled mirrors -presence-guard for code paths that do not
// receive the config (the REST heartbeat loop).
var presenceGuardEnabled = true

// desktopPresenceConflict reports local activity that contradicts a REST
// heartbeat for a desktop quest. Headless machines have nothing local to
// contradict, so they are never blocked; the guard exists for the case where
// the same person is visibly playing something else on this PC.
func desktopPresenceConflict() string {
	if !presenceGuardEnabled {
		return ""
	}
	state := presence.Snapshot()
	if !state.Supported || !state.Discord {
		return ""
	}
	if state.Fullscreen {
		return "a full-screen application is active on this machine"
	}
	return ""
}

// exitCode lets deferred cleanup report failure after the window is kept open.
var exitCode int

// hasEnrollable reports whether any quest still needs (re-)enrollment.
func hasEnrollable(quests []scanner.AnalyzedQuest) bool {
	for _, q := range quests {
		if q.State == scanner.StateAvailable && !q.IsExpired {
			return true
		}
	}
	return false
}

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

	timingJitter = cfg.Jitter
	scanner.EnrollSpread = cfg.Jitter
	presenceGuardEnabled = cfg.PresenceGuard

	defer func() {
		if cfg.KeepOpen {
			promptExit(false)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	// Main Orchestrator Execution
	printBanner()

	// Check for GitHub updates (non-blocking with quick 2s timeout)
	updateCtx, updateCancel := context.WithTimeout(context.Background(), 2*time.Second)
	if rel, hasUpdate := updater.CheckUpdate(updateCtx, Version); hasUpdate && rel != nil {
		fmt.Println(i18n.M().UpdateAvailableBanner)
		fmt.Printf(i18n.M().UpdateAvailableDetails, rel.TagName, rel.HTMLURL)
	}
	updateCancel()

	fmt.Printf("[+] OS: %s | Arch: %s | CPU Cores: %d\n", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	fmt.Printf("[+] Language: %s (Auto-detected / Configured)\n", i18n.GetLanguage())
	fmt.Printf("[+] Loaded token: %s\n", config.MaskToken(cfg.Token))
	fmt.Printf("[+] Mode: %s\n", getModeString(cfg))
	fmt.Printf("[+] Polling Interval: %v | Auto-Enroll: %t | Concurrency: %d\n", cfg.PollInterval, cfg.AutoAccept, cfg.Concurrency)
	fmt.Printf("[+] Region Sweep: %s | Captcha Portal: %t (Port: %d)\n", cfg.Region, cfg.EnablePortal, cfg.PortalPort)
	if cfg.QuestID != "" {
		fmt.Printf("[+] Targeted Quest ID: %s\n", cfg.QuestID)
	}
	if cfg.Duration > 0 {
		fmt.Printf("[+] Max Session Duration: %v\n", cfg.Duration)
	}

	// Initialize Persistent Web UI & Captcha Portal (if enabled)
	if cfg.EnablePortal {
		p, err := captcha.StartBackgroundPortal(cfg.PortalPort)
		if err == nil && p != nil {
			fmt.Printf("\n========================================================================\n")
			fmt.Printf(" [🌐] Local Web UI & Captcha Portal: %s\n", p.URL())
			fmt.Printf("      Dashboard & Test URL: %s/test\n", p.URL())
			fmt.Printf("========================================================================\n")
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Fetch Latest Client Build Number dynamically from Discord CDN
	fmt.Println(i18n.M().ScrapingCDN)
	buildNumber := api.FetchLatestBuildNumber()
	fmt.Printf(i18n.M().ActiveBuild, buildNumber)

	// 2. Initialize API Client
	apiClient := api.NewClient(cfg.Token, buildNumber)

	// 2.5. Build regional sweep targets from CLI flags. The local egress is always
	// included; regional proxies add real cross-region quest coverage because
	// Discord decides eligibility from the request IP, not locale headers.
	targets, sweepWarnings, targetErr := api.BuildTargets(cfg.Region, cfg.Proxy, cfg.ProxyFile, cfg.LocaleSweep)
	if targetErr != nil {
		fmt.Fprintf(os.Stderr, "[-] Invalid quest sweep configuration: %v\n", targetErr)
		exitCode = 1
		return
	}
	sweepTargets = targets
	targetCodes := make([]string, 0, len(targets))
	localeTargets := 0
	for _, target := range targets {
		targetCodes = append(targetCodes, target.Code)
		if target.Kind == api.TargetLocale {
			localeTargets++
		}
	}
	fmt.Printf("[+] Sweep targets (%d): %s\n", len(targets), strings.Join(targetCodes, ", "))
	for _, warning := range sweepWarnings {
		fmt.Printf("[i] %s\n", warning)
	}
	if localeTargets > 0 {
		fmt.Printf(i18n.M().LocaleSweepNote, localeTargets)
	}

	proxyEgress := 0
	for _, target := range targets {
		if target.ProxyURL != "" {
			proxyEgress++
		}
	}
	if proxyEgress > 1 {
		fmt.Printf("[!] %d proxy egress regions configured. Inside one scan this account will\n", proxyEgress)
		fmt.Println("    appear from that many different IPs - the strongest profiling signal the tool can emit.")
		fmt.Println("    Use -region direct (no proxies) for the quietest, most consistent profile.")
	}

	// 3. Authenticate & Retrieve Account Identity
	fmt.Println(i18n.M().Authenticating)
	user, err := apiClient.ValidateToken(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.M().AuthFailed, err)
		// A bad/expired token is a hard failure: exit non-zero so run.cmd can
		// report it clearly instead of printing "Finished successfully".
		fmt.Fprintln(os.Stderr, i18n.M().AuthFailedHint)
		exitCode = 1
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
	badgesStr := user.FormatBadges()
	if badgesStr == "None" {
		fmt.Printf("    - %s:      %s%s\n", i18n.M().BadgeProfileHeader, i18n.M().BadgeNone, i18n.M().BadgeTipPrompt)
	} else {
		fmt.Printf("    - %s:      %s\n", i18n.M().BadgeProfileHeader, badgesStr)
	}

	// 3.5. Badge Maximizer & HypeSquad House Management
	user = handleBadgeMaximizer(ctx, cfg, apiClient, user)

	// 3.6. Device identity: pin one consistent, self-consistent fingerprint to
	// the account instead of a hardcoded one shared by every installation.
	device := identity.Detect()
	if strings.EqualFold(cfg.Fingerprint, "legacy") {
		device.Fingerprint = api.DefaultFingerprint()
		device.Fingerprint.SystemLocale = device.SystemLocale
		device.Notes = append(device.Notes, "legacy mode: fingerprint values are fixed, not derived from this machine")
	}
	clientLocale := strings.TrimSpace(user.Locale)
	if clientLocale == "" {
		clientLocale = device.SystemLocale
	}

	var profile *identity.Profile
	if cfg.ProfileStore {
		store := identity.NewStore(cfg.CacheDir)
		loaded, created, profileErr := store.Enforce(user.ID, device)
		if profileErr != nil {
			fmt.Fprintf(os.Stderr, "[!] Profile store: %v\n", profileErr)
		}
		if loaded != nil {
			loaded.ClientLocale = clientLocale
			loaded.Apply(apiClient)
			profile = loaded
			printProfileReport(profile, device, created)
		}
	} else {
		apiClient.SetFingerprint(device.Fingerprint)
		apiClient.SetLocale(clientLocale, device.Timezone)
		printProfileReport(nil, device, false)
	}

	for {
		executePass(ctx, cfg, apiClient, user)

		if !cfg.Daemon || cfg.DryRun {
			break
		}

		if cfg.EnablePortal {
			fmt.Printf("\n[⏳] Daemon standing by. Web UI online. Next scan in %v...\n", cfg.PollInterval)
		} else {
			fmt.Printf("\n[⏳] Daemon standing by (Headless mode). Next scan in %v...\n", cfg.PollInterval)
		}
		select {
		case <-ctx.Done():
			fmt.Println("\n[*] Shutting down daemon gracefully...")
			return
		case <-time.After(pacing.Duration(cfg.PollInterval, timingJitter)):
			fmt.Printf("\n[⏰] Starting periodic quest scan at %s...\n", time.Now().Format("15:04:05"))
		}
	}
}

// discoverQuests runs the regional sweep and merges the last known good
// inventory when the pass was incomplete. It never silently returns an empty
// list while a recent snapshot still holds active quests.
func discoverQuests(ctx context.Context, cfg *config.Config, apiClient *api.Client, user *api.UserMe) (*api.SweepResult, int, error) {
	opts := api.DefaultSweepOptions()
	if cfg.SweepPacing > 0 {
		opts.MinInterval = cfg.SweepPacing
	}
	if cfg.SweepAttempts > 0 {
		opts.MaxAttempts = cfg.SweepAttempts
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}
	now := time.Now().UTC()

	result, sweepErr := apiClient.SweepQuests(ctx, sweepTargets, opts)
	if result == nil {
		result = &api.SweepResult{}
	}

	// A complete sweep is authoritative: never resurrect snapshot quests that
	// Discord itself no longer returns for a healthy account.
	if sweepErr == nil && result.Failed == 0 {
		saveSnapshot(cfg, userID, result)
		return result, 0, nil
	}

	// Incomplete sweep: salvage active quests the failed targets would have
	// returned, otherwise a single 429 silently hides a whole region for a pass.
	restored := 0
	snapshot, loadErr := inventory.Load(inventory.Path(cfg.CacheDir))
	if loadErr == nil && snapshot != nil {
		merged := inventory.Merge(result.Quests, snapshot, userID, now, 15*time.Minute)
		if len(merged) > len(result.Quests) {
			restored = len(merged) - len(result.Quests)
			result.Quests = merged
			if result.AccessSuspendedUntil == nil {
				result.AccessSuspendedUntil = snapshot.AccessSuspendedUntil
			}
			if result.EnrollmentBlockedUntil == nil {
				result.EnrollmentBlockedUntil = snapshot.EnrollmentBlockedUntil
			}
		}
	}

	if sweepErr != nil && len(result.Quests) == 0 {
		return result, 0, sweepErr
	}
	return result, restored, nil
}

// saveSnapshot persists a good inventory so a later transient outage cannot make
// quests vanish from the next pass.
func saveSnapshot(cfg *config.Config, userID string, result *api.SweepResult) {
	if result == nil || result.Failed > 0 {
		return
	}
	snapshot := inventory.FromSweep(userID, result, time.Now().UTC())
	if err := inventory.Save(inventory.Path(cfg.CacheDir), snapshot); err != nil && !snapshotSaveWarned {
		snapshotSaveWarned = true
		fmt.Fprintf(os.Stderr, "[!] Could not save quest snapshot (discovery fallback disabled): %v\n", err)
	}
}

// refreshQuestsByRegion re-reads the inventory through every egress region used
// by the pass, so a final summary does not mistake proxy-only quests for missing.
func refreshQuestsByRegion(ctx context.Context, client *api.Client, analyzed []scanner.AnalyzedQuest) []api.Quest {
	regionOrder := make([]string, 0, 4)
	regionQuests := make(map[string][]scanner.AnalyzedQuest)
	seenRegion := make(map[string]bool, 4)
	for _, quest := range analyzed {
		region := quest.Region
		if region == "" {
			region = "LOCAL"
		}
		if !seenRegion[region] {
			seenRegion[region] = true
			regionOrder = append(regionOrder, region)
		}
		regionQuests[region] = append(regionQuests[region], quest)
	}
	if len(regionOrder) == 0 {
		regionOrder = append(regionOrder, "LOCAL")
	}

	merged := make([]api.Quest, 0, len(analyzed))
	seenQuest := make(map[string]bool, len(analyzed))
	for _, region := range regionOrder {
		quests, fetchErr := client.FetchQuests(api.WithRegionInfo(ctx, regionInfoFor(region)))
		if fetchErr != nil {
			// Keep the pre-refresh view for this region instead of dropping it.
			for _, quest := range regionQuests[region] {
				if seenQuest[quest.ID] {
					continue
				}
				seenQuest[quest.ID] = true
				copyOfQuest := quest.Quest
				if copyOfQuest.DiscoveredRegion == "" {
					copyOfQuest.DiscoveredRegion = region
				}
				merged = append(merged, copyOfQuest)
			}
			continue
		}
		for _, quest := range quests {
			if quest.ID == "" || seenQuest[quest.ID] {
				continue
			}
			seenQuest[quest.ID] = true
			if quest.DiscoveredRegion == "" {
				quest.DiscoveredRegion = region
			}
			merged = append(merged, quest)
		}
	}
	return merged
}

// regionInfoFor returns the routing metadata for a quest's discovery region.
func regionInfoFor(region string) api.RegionInfo {
	for _, target := range sweepTargets {
		if strings.EqualFold(target.Code, region) {
			return target.Info()
		}
	}
	return api.RegionInfo{Code: region, Locale: "en-US", Timezone: "America/New_York"}
}

// printSweepReport prints per-target sweep results so a failed region can never
// be silently swallowed again.
func printSweepReport(result *api.SweepResult, restored int) {
	if result == nil {
		return
	}
	if len(result.Targets) > 0 {
		fmt.Printf(i18n.M().SweepSummary,
			result.Succeeded, len(result.Targets), len(result.Quests), result.ExcludedCount(),
			result.Requests, result.Duration.Round(100*time.Millisecond))
	}
	for _, target := range result.Targets {
		state := "OK "
		if !target.Ok {
			state = "ERR"
		}
		proxy := target.Proxy
		if proxy == "" {
			proxy = "direct"
		}
		if target.Ok {
			fmt.Printf("    [%s] %-8s locale=%-5s %-28s quests=%-3d excluded=%-3d %dms\n",
				state, target.Code, target.Locale, proxy, target.Quests, target.Excluded, target.Latency.Milliseconds())
			continue
		}
		fmt.Printf("    [%s] %-8s locale=%-5s %-28s %s (attempts=%d)\n",
			state, target.Code, target.Locale, proxy, target.Err, target.Attempts)
	}
	if failed := result.FailedTargets(); len(failed) > 0 {
		fmt.Printf(i18n.M().SweepTargetsFailed, len(failed), strings.Join(failed, ", "))
	}
	if restored > 0 {
		fmt.Printf(i18n.M().CacheFallbackNote, restored)
	}
	if excluded := result.ExcludedCount(); excluded > 0 {
		fmt.Printf(i18n.M().ExcludedHint, excluded)
	}
}

// mergeRefreshed keeps proxy-discovered quests that a direct refresh cannot see.
func mergeRefreshed(existing []scanner.AnalyzedQuest, refreshed []api.Quest) []api.Quest {
	merged := make([]api.Quest, 0, len(refreshed)+len(existing))
	seen := make(map[string]bool, len(refreshed))
	for _, quest := range refreshed {
		seen[quest.ID] = true
		merged = append(merged, quest)
	}
	for _, analyzed := range existing {
		if seen[analyzed.ID] {
			continue
		}
		merged = append(merged, analyzed.Quest)
	}
	return merged
}

// countSkippedEnrollments counts quests that cannot safely be enrolled this
// pass: snapshot-restored quests and quests whose regional proxy disappeared.
func countSkippedEnrollments(quests []scanner.AnalyzedQuest, client *api.Client) int {
	skipped := 0
	for _, quest := range quests {
		if quest.State != scanner.StateAvailable || quest.IsExpired {
			continue
		}
		if quest.Quest.FromCache {
			skipped++
			continue
		}
		if quest.Quest.RegionProxied && quest.Region != "" && !client.HasRegionProxy(quest.Region) {
			skipped++
		}
	}
	return skipped
}

func executePass(ctx context.Context, cfg *config.Config, apiClient *api.Client, user *api.UserMe) {
	// In daemon mode, suppress the verbose scan/table output when the inventory
	// is byte-identical to the previous cycle.
	quiet := cfg.Daemon && !cfg.Debug

	if !quiet {
		fmt.Printf(i18n.M().SweepStart, len(sweepTargets))
	}

	result, restored, err := discoverQuests(ctx, cfg, apiClient, user)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.M().FailedRetrieveQuests, err)
		return
	}
	if result == nil {
		fmt.Fprintf(os.Stderr, i18n.M().FailedRetrieveQuests, errors.New("empty discovery result"))
		return
	}
	if !quiet || restored > 0 || result.Failed > 0 {
		printSweepReport(result, restored)
	}

	// Account-level gates come straight from Discord's payload. A suspended
	// account must not send any quest traffic, or the penalty escalates.
	access := result.AccessState(time.Now().UTC())
	if access.AccessSuspended {
		fmt.Printf(i18n.M().AccessSuspended, access.AccessSuspendedUntil.Format("2006-01-02 15:04 MST"))
		analyzed := scanner.AnalyzeAll(result.Quests)
		fmt.Print(scanner.FormatQuestTable(analyzed, cfg.Debug))
		printSessionSummary(user, analyzed)
		return
	}
	if access.EnrollmentBlocked {
		fmt.Printf(i18n.M().EnrollBlocked, access.EnrollmentBlockedUntil.Format("2006-01-02 15:04 MST"))
	}

	analyzedQuests := scanner.AnalyzeAll(result.Quests)
	sig := scanner.TableSignature(analyzedQuests)
	inventoryChanged := sig != lastInventorySignature
	lastInventorySignature = sig

	if !quiet || inventoryChanged {
		fmt.Print(scanner.FormatQuestTable(analyzedQuests, cfg.Debug))
	}

	// Auto-Enrollment (only when the account is allowed to enroll).
	if cfg.AutoAccept && !cfg.DryRun && !access.EnrollmentBlocked {
		if !quiet || hasEnrollable(analyzedQuests) {
			fmt.Println(i18n.M().AutoEnrollHeader)
		}
		if skipped := countSkippedEnrollments(analyzedQuests, apiClient); skipped > 0 {
			fmt.Printf(i18n.M().CacheEnrollSkip, skipped)
		}
		enrolledIDs, enrollErr := scanner.AutoEnrollPending(ctx, apiClient, analyzedQuests, cfg.EnablePortal, cfg.PortalPort, quiet)
		if enrollErr != nil {
			fmt.Printf(i18n.M().AutoEnrollError, enrollErr)
		}
		if updated := scanner.ApplyEnrolledState(analyzedQuests, enrolledIDs); updated > 0 {
			fmt.Printf(i18n.M().AutoEnrollSuccess, updated)
			// The freshly enrolled quests are known-good locally, so there is no
			// need to spend another rate-limited refresh to start working on them.
			lastInventorySignature = scanner.TableSignature(analyzedQuests)
			fmt.Print(scanner.FormatQuestTable(analyzedQuests, cfg.Debug))
		} else if !quiet {
			fmt.Println(i18n.M().AutoEnrollAllEnrolled)
		}
	}

	// Dry-run Mode
	if cfg.DryRun {
		fmt.Println(i18n.M().DryRunNotice)
		printSessionSummary(user, analyzedQuests)
		return
	}

	// Execution Engine Branching
	if cfg.UseSpoofer {
		fmt.Printf("\n[*] %s\n", i18n.M().ModeSpoofer)
		spooferEngine, spooferErr := spoofer.NewSpoofer(cfg.CacheDir)
		if spooferErr != nil {
			fmt.Fprintf(os.Stderr, "[-] Failed initializing spoofer engine: %v\n", spooferErr)
			return
		}
		if runErr := runSpooferLoop(ctx, cfg, apiClient, spooferEngine, analyzedQuests); runErr != nil {
			if runErr != context.Canceled && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "[-] Spoofer error: %v\n", runErr)
			}
		}
	} else {
		fmt.Printf("\n[*] %s\n", i18n.M().ModeAPIRunner)
		if cfg.PresenceGuard {
			if state := presence.Snapshot(); state.Supported && state.Discord {
				fmt.Println("[i] Discord desktop is running here. For desktop quests the local spoofer (-spoofer) is the")
				fmt.Println("    realistic path; REST heartbeats are only a fallback, and they pause while a full-screen game is active.")
			}
		}
		if runErr := runAPIRunnerLoop(ctx, cfg, apiClient, analyzedQuests); runErr != nil {
			if runErr != context.Canceled && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "[-] API Runner error: %v\n", runErr)
			}
		}
	}

	fmt.Println(i18n.M().AllTasksFinished)
	// A final refresh is cosmetic; skip it in daemon mode so a long-running
	// service does not spend extra rate-limited quest requests every cycle.
	if !cfg.Daemon {
		if refreshedQuests := refreshQuestsByRegion(ctx, apiClient, analyzedQuests); len(refreshedQuests) > 0 {
			analyzedQuests = scanner.AnalyzeAll(refreshedQuests)
		}
	}
	if refreshedUser, userErr := apiClient.ValidateToken(ctx); userErr == nil {
		user = refreshedUser
	}
	printSessionSummary(user, analyzedQuests)
}

func handleBadgeMaximizer(ctx context.Context, cfg *config.Config, apiClient *api.Client, user *api.UserMe) *api.UserMe {
	// 1. Explicit CLI HypeSquad action
	if cfg.HypeSquad != "" {
		houseID, isLeave, err := config.ParseHypeSquadAction(cfg.HypeSquad)
		if err != nil {
			fmt.Printf(i18n.M().BadgeError, err)
			return user
		}

		if isLeave {
			if cfg.DryRun {
				fmt.Print(i18n.M().BadgeDryRunLeave)
				return user
			}
			fmt.Print(i18n.M().BadgeLeavingHypeSquad)
			if err := apiClient.LeaveHypeSquad(ctx); err != nil {
				fmt.Printf(i18n.M().BadgeError, err)
			} else {
				fmt.Print(i18n.M().BadgeLeaveSuccess)
				if refreshed, err := apiClient.ValidateToken(ctx); err == nil {
					user = refreshed
					fmt.Printf("    - %s: %s\n", i18n.M().BadgeProfileHeader, user.FormatBadges())
				}
			}
			return user
		}

		houseName := "House of Bravery"
		switch houseID {
		case api.HypeSquadHouseBrilliance:
			houseName = "House of Brilliance"
		case api.HypeSquadHouseBalance:
			houseName = "House of Balance"
		}

		if cfg.DryRun {
			fmt.Printf(i18n.M().BadgeDryRunJoin, houseName)
			return user
		}

		fmt.Printf(i18n.M().BadgeJoiningHypeSquad, houseName)
		if err := apiClient.JoinHypeSquad(ctx, houseID); err != nil {
			fmt.Printf(i18n.M().BadgeError, err)
		} else {
			fmt.Printf(i18n.M().BadgeClaimSuccess, houseName)
			if refreshed, err := apiClient.ValidateToken(ctx); err == nil {
				user = refreshed
				fmt.Printf("    - %s: %s\n", i18n.M().BadgeProfileHeader, user.FormatBadges())
			}
		}
		return user
	}

	// 2. MaxBadge auto-maximization
	if cfg.MaxBadge {
		if user.HasHypeSquad() {
			_, houseName := user.GetHypeSquadHouse()
			fmt.Printf(i18n.M().BadgeAlreadyOwned, houseName)
		} else {
			if cfg.DryRun {
				fmt.Print(i18n.M().BadgeDryRunAutoEnroll)
				return user
			}
			fmt.Print(i18n.M().BadgeAutoEnrolling)
			if err := apiClient.JoinHypeSquad(ctx, api.HypeSquadHouseBravery); err != nil {
				fmt.Printf(i18n.M().BadgeError, err)
			} else {
				fmt.Printf(i18n.M().BadgeClaimSuccess, "House of Bravery")
				if refreshed, err := apiClient.ValidateToken(ctx); err == nil {
					user = refreshed
					fmt.Printf("    - %s: %s\n", i18n.M().BadgeProfileHeader, user.FormatBadges())
				}
			}
		}
	}

	return user
}

func printSessionSummary(user *api.UserMe, quests []scanner.AnalyzedQuest) {
	fmt.Print(i18n.M().SummaryHeader)
	badges := "None"
	if user != nil {
		badges = user.FormatBadges()
	}
	fmt.Printf(i18n.M().SummaryBadges, badges)

	var completedCount, totalActive int
	for _, q := range quests {
		if !q.IsExpired {
			totalActive++
			if q.State == scanner.StateCompleted || q.State == scanner.StateClaimed {
				completedCount++
			}
		}
	}
	fmt.Printf(i18n.M().SummaryQuestsDone, completedCount, totalActive)
	fmt.Print(i18n.M().SummaryClaimReminder)
}

// printProfileReport shows which device identity the session will claim and
// every inconsistency the audit found. It runs once per start, before traffic.
func printProfileReport(profile *identity.Profile, device identity.Device, created bool) {
	fmt.Printf("[+] Device profile: %s\n", device.Summary())
	if profile != nil {
		state := "reused"
		if created {
			state = "created"
		}
		fmt.Printf("[+] Account profile: %s %s | home egress %s | secrets %s\n",
			state, profile.CreatedAt.Format("2006-01-02"), profile.HomeEgress, identity.ProtectionLabel())
	}

	// Report the honest session egress: a proxy region changes the IP the
	// account appears from, so the audit should know which one is in play.
	egressRegion, egressTimezone := "direct", device.Timezone
	for _, target := range sweepTargets {
		if target.ProxyURL != "" {
			egressRegion = target.Code
			if target.Timezone != "" {
				egressTimezone = target.Timezone
			}
			break
		}
	}

	issues := identity.Audit(profile, device, egressRegion, egressTimezone)
	for _, note := range device.Notes {
		fmt.Printf("    [i] %s\n", note)
	}
	for _, issue := range issues {
		switch issue.Level {
		case "error":
			fmt.Printf("    [!] %s\n", issue.Message)
		case "warn":
			fmt.Printf("    [~] %s\n", issue.Message)
		default:
			fmt.Printf("    [i] %s\n", issue.Message)
		}
	}
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

	// Group eligible quests by Region to process region batches cleanly
	// Quests belonging to the same region are executed in parallel up to concurrency.
	// Only when that region finishes does it proceed to the next region.
	regionBuckets := make(map[string][]scanner.AnalyzedQuest)
	var regionOrder []string
	for _, q := range eligible {
		reg := q.Region
		if reg == "" {
			reg = "GLOBAL"
		}
		if _, exists := regionBuckets[reg]; !exists {
			regionOrder = append(regionOrder, reg)
		}
		regionBuckets[reg] = append(regionBuckets[reg], q)
	}

	fmt.Printf(i18n.M().FoundEligibleQuests, len(eligible))
	totalProcessed := 0
	for _, reg := range regionOrder {
		bucket := regionBuckets[reg]
		fmt.Printf("\n[🌐] --- Region [%s] Batch (%d quests) ---\n", reg, len(bucket))
		for i, q := range bucket {
			rem := float64(q.TargetSec) - q.CurrentSec
			if rem < 0 {
				rem = 0
			}
			remStr := fmt.Sprintf(i18n.M().RemainingQuestSec, rem, q.TargetSec)
			fmt.Printf("    %d. [%s] %-24s | %-22s | %s\n",
				i+1, reg, q.Title, q.Category.Localized(), remStr)
		}

		concurrency := cfg.Concurrency
		if concurrency <= 0 {
			concurrency = 5
		}
		if concurrency > len(bucket) {
			concurrency = len(bucket)
		}

		fmt.Printf(i18n.M().ConcurrentHeader, len(bucket), concurrency)

		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for i, q := range bucket {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			sem <- struct{}{}
			wg.Add(1)

			go func(idx int, quest scanner.AnalyzedQuest) {
				defer func() {
					<-sem
					wg.Done()
				}()

				// Stagger start per concurrent lane, jittered, so the batch never
				// looks like N requests fired by one clock.
				laneDelay := pacing.Duration(time.Duration((idx-1)%concurrency)*2500*time.Millisecond, timingJitter)
				select {
				case <-ctx.Done():
					return
				case <-time.After(laneDelay):
				}

				questCtx := api.WithRegionInfo(ctx, regionInfoFor(quest.Region))
				processOneQuest(questCtx, client, quest, idx, len(bucket))
			}(i+1, q)
		}

		wg.Wait()
		totalProcessed += len(bucket)
		fmt.Printf("[✓] Region [%s] Batch Complete!\n", reg)
	}

	fmt.Printf(i18n.M().AllQuestsDone, totalProcessed)
	return nil
}

func processOneQuest(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest, idx, total int) {
	fmt.Printf("\n==================================================================\n")
	fmt.Printf("%s\n", i18n.T(func(m i18n.Messages) string { return m.ProcessingQuestHeader }, idx, total, q.Title, q.Category.Localized()))
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
		return
	}

	if err != nil {
		if ctx.Err() == nil && !errors.Is(err, api.ErrQuestExpired) {
			fmt.Printf(i18n.M().QuestErrorSkip, q.Title, err)
		}
	}
}

// Progress loops must never spin forever when Discord keeps rejecting a quest.
// These guards stop the "endless error spam" that happened when a quest expired
// while the tool was idle: no progress advance -> abort with one clear line.
const (
	maxConsecutiveErrors = 5
	progressStallTimeout = 3 * time.Minute
	maxErrorBackoff      = 30 * time.Second
)

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

	consecutiveErrors := 0
	stallDeadline := time.Now().Add(progressStallTimeout)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Now().After(stallDeadline) {
			fmt.Printf(i18n.M().ProgressStallAbort, q.Title, secondsDone, secondsNeeded)
			return api.ErrQuestExpired
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
				if errors.Is(err, api.ErrQuestExpired) {
					fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
					return err
				}
				consecutiveErrors++
				// Only surface the first failure and periodic heartbeats of it,
				// so a flaky network cannot flood the terminal.
				if consecutiveErrors == 1 || consecutiveErrors%10 == 0 {
					fmt.Printf(i18n.M().VideoProgressWarning, err)
				}
				if consecutiveErrors >= maxConsecutiveErrors {
					fmt.Printf(i18n.M().ProgressErrorAbort, q.Title, consecutiveErrors, err)
					return err
				}
				if !sleepWithContext(ctx, errorBackoff(consecutiveErrors)) {
					return ctx.Err()
				}
				continue
			}

			consecutiveErrors = 0
			stallDeadline = time.Now().Add(progressStallTimeout)
			secondsDone = targetTimestamp
			fmt.Printf(i18n.M().VideoProgress, secondsDone, secondsNeeded, (secondsDone/secondsNeeded)*100.0)
			if resp != nil && resp.CompletedAt != nil && *resp.CompletedAt != "" {
				fmt.Printf(i18n.M().VideoCompleted, q.Title)
				return nil
			}
		}

		if timestamp >= secondsNeeded {
			break
		}
		if !sleepWithContext(ctx, pacing.Duration(time.Second, timingJitter)) {
			return ctx.Err()
		}
	}

	// Final progress update — report failures instead of pretending success.
	if _, err := client.SendVideoProgress(ctx, q.ID, secondsNeeded); err != nil {
		if errors.Is(err, api.ErrQuestExpired) {
			fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
			return err
		}
		fmt.Printf(i18n.M().VideoProgressWarning, err)
		return err
	}
	fmt.Printf(i18n.M().VideoCompleted, q.Title)
	return nil
}

// sleepWithContext waits for d, returning false if the context was cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// errorBackoff grows the retry delay but stays bounded.
func errorBackoff(consecutive int) time.Duration {
	d := time.Duration(consecutive) * 2 * time.Second
	if d > maxErrorBackoff {
		d = maxErrorBackoff
	}
	return pacing.Duration(d, timingJitter)
}

func completeHeartbeatAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	pid := rand.Intn(28000) + 2000
	streamKey := fmt.Sprintf("call:0:%d", pid)

	rem := secondsNeeded - secondsDone
	fmt.Printf(i18n.M().HeartbeatStart, q.Title, rem/60.0)

	consecutiveErrors := 0
	stallDeadline := time.Now().Add(progressStallTimeout)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Now().After(stallDeadline) {
			fmt.Printf(i18n.M().ProgressStallAbort, q.Title, secondsDone, secondsNeeded)
			return api.ErrQuestExpired
		}

		// A REST heartbeat claims a desktop game is running. While the owner is
		// visibly playing something else, that claim is exactly the contradiction
		// that gets accounts flagged, so wait instead of sending it.
		if reason := desktopPresenceConflict(); reason != "" {
			fmt.Printf("    [~] Holding %s: %s.\n", q.Title, reason)
			for reason != "" {
				if !pacing.Sleep(ctx, 20*time.Second, timingJitter) {
					return ctx.Err()
				}
				reason = desktopPresenceConflict()
			}
			stallDeadline = time.Now().Add(progressStallTimeout)
			fmt.Printf("    [i] Resuming %s.\n", q.Title)
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			if errors.Is(err, api.ErrQuestExpired) {
				fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
				return err
			}
			consecutiveErrors++
			if consecutiveErrors == 1 || consecutiveErrors%10 == 0 {
				fmt.Printf(i18n.M().HeartbeatWarning, err)
			}
			if consecutiveErrors >= maxConsecutiveErrors {
				fmt.Printf(i18n.M().ProgressErrorAbort, q.Title, consecutiveErrors, err)
				return err
			}
			if !sleepWithContext(ctx, errorBackoff(consecutiveErrors)) {
				return ctx.Err()
			}
			continue
		}

		if resp != nil {
			consecutiveErrors = 0
			previousDone := secondsDone
			if resp.Progress != nil {
				if prog, ok := resp.Progress[q.TaskType]; ok {
					secondsDone = prog.Value
				} else if prog, ok := resp.Progress["PLAY_ON_DESKTOP"]; ok {
					secondsDone = prog.Value
				}
			} else {
				secondsDone += 20.0
			}
			// Discord accepted the heartbeat but progress did not move: keep
			// waiting, but the stall deadline above prevents an endless loop.
			if secondsDone > previousDone {
				stallDeadline = time.Now().Add(progressStallTimeout)
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

		if !sleepWithContext(ctx, pacing.Duration(20*time.Second, timingJitter)) {
			return ctx.Err()
		}
	}

	// Final terminal heartbeat
	if _, err := client.SendHeartbeat(ctx, q.ID, streamKey, true); err != nil {
		if errors.Is(err, api.ErrQuestExpired) {
			fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
			return err
		}
		fmt.Printf(i18n.M().HeartbeatWarning, err)
		return err
	}
	fmt.Printf(i18n.M().HeartbeatCompleted, q.Title)
	return nil
}

func completeActivityAPI(ctx context.Context, client *api.Client, q scanner.AnalyzedQuest) error {
	secondsNeeded := float64(q.TargetSec)
	secondsDone := q.CurrentSec
	streamKey := "call:0:1"

	fmt.Printf(i18n.M().ActivityStart, q.Title)

	consecutiveErrors := 0
	stallDeadline := time.Now().Add(progressStallTimeout)

	for secondsDone < secondsNeeded {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Now().After(stallDeadline) {
			fmt.Printf(i18n.M().ProgressStallAbort, q.Title, secondsDone, secondsNeeded)
			return api.ErrQuestExpired
		}

		resp, err := client.SendHeartbeat(ctx, q.ID, streamKey, false)
		if err != nil {
			if errors.Is(err, api.ErrQuestExpired) {
				fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
				return err
			}
			consecutiveErrors++
			if consecutiveErrors == 1 || consecutiveErrors%10 == 0 {
				fmt.Printf(i18n.M().ActivityWarning, err)
			}
			if consecutiveErrors >= maxConsecutiveErrors {
				fmt.Printf(i18n.M().ProgressErrorAbort, q.Title, consecutiveErrors, err)
				return err
			}
			if !sleepWithContext(ctx, errorBackoff(consecutiveErrors)) {
				return ctx.Err()
			}
			continue
		}

		if resp != nil {
			consecutiveErrors = 0
			previousDone := secondsDone
			if resp.Progress != nil && resp.Progress["PLAY_ACTIVITY"].Value > 0 {
				secondsDone = resp.Progress["PLAY_ACTIVITY"].Value
			} else {
				secondsDone += 20.0
			}
			if secondsDone > previousDone {
				stallDeadline = time.Now().Add(progressStallTimeout)
			}
			fmt.Printf(i18n.M().ActivityProgress, secondsDone, secondsNeeded)
			if (resp.CompletedAt != nil && *resp.CompletedAt != "") || secondsDone >= secondsNeeded {
				break
			}
		}

		if !sleepWithContext(ctx, pacing.Duration(20*time.Second, timingJitter)) {
			return ctx.Err()
		}
	}

	if _, err := client.SendHeartbeat(ctx, q.ID, streamKey, true); err != nil {
		if errors.Is(err, api.ErrQuestExpired) {
			fmt.Printf(i18n.M().QuestExpiredSkip, q.Title, q.ID)
			return err
		}
		fmt.Printf(i18n.M().ActivityWarning, err)
		return err
	}
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

	// One machine, one game. Discord's client reports at most one detected game
	// at a time, so launching several stubs in parallel tells Discord this
	// account is playing several games simultaneously - something no real
	// machine can do, and the exact pattern that produced "unusual activity"
	// warnings. Desktop quests therefore run strictly one after another.
	//
	// The known-game table covers every desktop quest in the inventory, not
	// just the eligible ones: a title the owner plays only needs to have had a
	// quest once for us to recognise it and stay out of the way.
	knownGames := make(map[string]string, len(quests))
	for _, q := range quests {
		if q.TaskType != "PLAY_ON_DESKTOP" {
			continue
		}
		exe, title, err := spooferEngine.ResolveExecutable(q.AppID)
		if err != nil || exe == "" {
			continue
		}
		if title == "" {
			title = q.Title
		}
		knownGames[presence.Lower(exe)] = title
	}

	if cfg.PresenceGuard {
		state := presence.Snapshot()
		if state.Supported && !state.Discord {
			fmt.Println("[!] Discord desktop is not running. Desktop quests only advance when the local client can see the game,")
			fmt.Println("    so spoofing now would produce activity nobody reports. Start Discord and run again.")
			fmt.Println("    (If this machine really has no Discord client, pass -presence-guard=false to skip the check.)")
			return nil
		}
	}

	for i := range eligible {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := runSpooferQuest(ctx, cfg, apiClient, spooferEngine, eligible[i], knownGames); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			fmt.Fprintf(os.Stderr, "[-] Spoofer: %v\n", err)
		}
	}
	return nil
}

// acquireSpooferLock blocks while another instance already shows a game on
// this machine. Two accounts spoofing two different games at once is a
// machine-level contradiction no fingerprint can explain.
func acquireSpooferLock(ctx context.Context) func() {
	announced := false
	for {
		release, ok := presence.TryLock("DiscordQuestSpoofer")
		if ok {
			return release
		}
		if !announced {
			fmt.Println("[~] Another Phantom instance is already showing a game on this machine.")
			fmt.Println("    Waiting for it to finish; only one game can be visible at a time.")
			announced = true
		}
		if !pacing.Sleep(ctx, 30*time.Second, timingJitter) {
			return nil
		}
	}
}

// runSpooferQuest farms one desktop quest. It either launches a single stub or
// piggybacks on the real, already-running game, and it stops pretending the
// moment the owner starts playing something else.
func runSpooferQuest(ctx context.Context, cfg *config.Config, apiClient *api.Client, spooferEngine spoofer.Spoofer, quest scanner.AnalyzedQuest, knownGames map[string]string) error {
	exeName, gameTitle, err := spooferEngine.ResolveExecutable(quest.AppID)
	if err != nil {
		fmt.Printf(i18n.M().SpooferNoExe, quest.AppID, err)
		return nil
	}
	if gameTitle == "" {
		gameTitle = quest.Title
	}
	others := otherGames(knownGames, exeName)

	var proc spoofer.GameProcess
	var releaseLock func()
	stubPID := 0
	stopStub := func() {
		if proc != nil {
			_ = proc.Stop()
			proc = nil
			stubPID = 0
		}
		if releaseLock != nil {
			releaseLock()
			releaseLock = nil
		}
	}
	launch := func() error {
		// The machine-wide slot is taken only while a stub is actually visible,
		// so two accounts on one PC can take turns instead of one starving.
		if cfg.PresenceGuard {
			releaseLock = acquireSpooferLock(ctx)
			if releaseLock == nil {
				return ctx.Err()
			}
		}
		started, launchErr := spooferEngine.LaunchGame(ctx, quest.AppID, gameTitle, exeName)
		if launchErr != nil {
			if releaseLock != nil {
				releaseLock()
				releaseLock = nil
			}
			return launchErr
		}
		proc = started
		stubPID = started.PID()
		fmt.Printf(i18n.M().SpooferProcCreated, started.ExecutableName(), started.PID())
		return nil
	}
	defer stopStub()

	if cfg.PresenceGuard {
		verdict, waitErr := waitForPresence(ctx, exeName, gameTitle, others, 0)
		if waitErr != nil {
			return waitErr
		}
		switch verdict.Kind {
		case presence.NoDiscord:
			fmt.Println("[!] Discord desktop closed; pausing desktop quests until it is back.")
			return nil
		case presence.WaitOtherGame:
			fmt.Printf("    [~] Deferring %s: %s. Its own quest gets this turn instead.\n", gameTitle, verdict.Detail)
			return nil
		case presence.RealSession:
			fmt.Printf("    [i] %s is already running for real - letting your own session count instead of starting a second copy.\n", gameTitle)
		default:
			if launchErr := launch(); launchErr != nil {
				fmt.Printf(i18n.M().SpooferLaunchError, launchErr)
				return nil
			}
		}
	} else if launchErr := launch(); launchErr != nil {
		fmt.Printf(i18n.M().SpooferLaunchError, launchErr)
		return nil
	}

	region := quest.Region
	if region == "" {
		region = "LOCAL"
	}

	consecutiveFetchErrors := 0
	noProgressPolls := 0
	lastSec := quest.CurrentSec

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pacing.Duration(cfg.PollInterval, timingJitter)):
		}

		// Presence watch. A real game always wins: fighting it for Discord's
		// single "Playing" slot is what makes the account look automated.
		if cfg.PresenceGuard {
			verdict := presence.Evaluate(presence.Snapshot(), presence.Target{
				Exe:     exeName,
				Title:   gameTitle,
				StubPID: stubPID,
			}, others)

			switch verdict.Kind {
			case presence.NoDiscord:
				stopStub()
				fmt.Println("[!] Discord desktop closed; pausing desktop quests until it is back.")
				return nil
			case presence.Ready:
				// The real session ended while we were riding it; take over with
				// a stub so the quest keeps moving.
				if proc == nil {
					if launchErr := launch(); launchErr != nil {
						fmt.Printf(i18n.M().SpooferLaunchError, launchErr)
						return nil
					}
				}
			case presence.WaitOtherGame:
				stopStub()
				fmt.Printf("    [~] Pausing %s: %s. Not competing with your own game.\n", gameTitle, verdict.Detail)
				return nil
			case presence.WaitFullscreen:
				stopStub()
				fmt.Printf("    [~] Pausing %s: %s.\n", gameTitle, verdict.Detail)
				ready, waitErr := waitForPresence(ctx, exeName, gameTitle, others, 0)
				if waitErr != nil {
					return waitErr
				}
				if ready.Kind == presence.NoDiscord || ready.Kind == presence.WaitOtherGame {
					return nil
				}
				noProgressPolls = 0
				if ready.Kind != presence.RealSession {
					if launchErr := launch(); launchErr != nil {
						fmt.Printf(i18n.M().SpooferLaunchError, launchErr)
						return nil
					}
				}
			}
		}

		rawQuests, fetchErr := apiClient.FetchQuests(api.WithRegionInfo(ctx, regionInfoFor(region)))
		if fetchErr != nil {
			if errors.Is(fetchErr, api.ErrQuestExpired) {
				return nil
			}
			// Never spin on a broken connection without a bound.
			consecutiveFetchErrors++
			if consecutiveFetchErrors == 1 || consecutiveFetchErrors%10 == 0 {
				fmt.Printf(i18n.M().SpooferFetchWarning, fetchErr)
			}
			if consecutiveFetchErrors >= maxConsecutiveErrors {
				fmt.Printf(i18n.M().ProgressErrorAbort, "spoofer", consecutiveFetchErrors, fetchErr)
				return nil
			}
			continue
		}
		consecutiveFetchErrors = 0

		var current *api.Quest
		for i := range rawQuests {
			if rawQuests[i].ID == quest.ID {
				current = &rawQuests[i]
				break
			}
		}
		if current == nil {
			// Quest vanished from the inventory (ended / removed).
			fmt.Printf(i18n.M().QuestExpiredSkip, quest.Title, quest.ID)
			return nil
		}

		analyzed := scanner.AnalyzeQuest(*current)
		fmt.Printf("    [ðŸŽ® %s] Progress: %.0fs / %ds (State: %s)\n", analyzed.Title, analyzed.CurrentSec, analyzed.TargetSec, analyzed.State)

		if analyzed.IsExpired || analyzed.State == scanner.StateExpired {
			fmt.Printf(i18n.M().QuestExpiredSkip, analyzed.Title, analyzed.ID)
			return nil
		}

		if analyzed.State == scanner.StateCompleted || analyzed.State == scanner.StateClaimed ||
			(analyzed.TargetSec > 0 && analyzed.CurrentSec >= float64(analyzed.TargetSec)) {
			fmt.Printf(i18n.M().SpooferCompleted, analyzed.Title)
			return nil
		}

		// A quest that reports "not finished" but never advances means Discord
		// stopped counting it. Give up on this pass instead of holding a process
		// for hours.
		if analyzed.CurrentSec > lastSec {
			lastSec = analyzed.CurrentSec
			noProgressPolls = 0
		} else {
			noProgressPolls++
			if cfg.PollInterval > 0 {
				if maxIdle := int(progressStallTimeout / cfg.PollInterval); maxIdle > 0 && noProgressPolls >= maxIdle {
					fmt.Printf(i18n.M().ProgressStallAbort, "spoofer", analyzed.CurrentSec, float64(analyzed.TargetSec))
					return nil
				}
			}
		}
	}
}

// otherGames maps the other eligible quests' executables to their titles.
func otherGames(known map[string]string, targetExe string) map[string]string {
	target := presence.Lower(targetExe)
	others := make(map[string]string, len(known))
	for exe, title := range known {
		if exe != target {
			others[exe] = title
		}
	}
	return others
}

// waitForPresence blocks until the machine can show title without lying. A
// different game that is already running defers the quest (its own turn comes
// next); a full-screen app we cannot name means "wait it out".
func waitForPresence(ctx context.Context, exe, title string, others map[string]string, stubPID int) (presence.Verdict, error) {
	var lastKind presence.Kind
	var lastNotice time.Time

	for {
		verdict := presence.Evaluate(presence.Snapshot(), presence.Target{
			Exe:     exe,
			Title:   title,
			StubPID: stubPID,
		}, others)

		switch verdict.Kind {
		case presence.Ready, presence.RealSession, presence.NoDiscord:
			return verdict, nil
		case presence.WaitOtherGame:
			if stubPID == 0 {
				return verdict, nil
			}
		}

		if verdict.Kind != lastKind || time.Since(lastNotice) >= 5*time.Minute {
			fmt.Printf("    [~] Holding %s: %s. Waiting for the machine to be free...\n", title, verdict.Detail)
			lastKind = verdict.Kind
			lastNotice = time.Now()
		}
		if !pacing.Sleep(ctx, 20*time.Second, timingJitter) {
			return verdict, ctx.Err()
		}
	}
}
