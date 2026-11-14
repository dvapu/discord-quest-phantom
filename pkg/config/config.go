package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"discord-quest-completer/pkg/i18n"
	"discord-quest-completer/pkg/identity"
)

// Config represents the application runtime configuration.
type Config struct {
	Token         string
	PollInterval  time.Duration
	AutoAccept    bool
	Debug         bool
	APIBase       string
	DryRun        bool
	CacheDir      string
	RunStub       bool
	StubTitle     string
	QuestID       string
	Duration      time.Duration
	UseSpoofer    bool
	KeepOpen      bool
	Lang          string
	Region        string
	Proxy         string
	ProxyFile     string
	LocaleSweep   string
	SweepPacing   time.Duration
	SweepAttempts int
	EnablePortal  bool
	PortalPort    int
	Concurrency   int
	Daemon        bool
	HypeSquad     string
	MaxBadge      bool
	Fingerprint   string
	Jitter        float64
	ProfileStore  bool
	PresenceGuard bool
}

// ParseHypeSquadAction parses a user string for HypeSquad House enrollment/leaving.
func ParseHypeSquadAction(val string) (int, bool, error) {
	v := strings.ToLower(strings.TrimSpace(val))
	switch v {
	case "":
		return 0, false, nil
	case "bravery", "1", "house1", "house_1":
		return 1, false, nil
	case "brilliance", "2", "house2", "house_2":
		return 2, false, nil
	case "balance", "3", "house3", "house_3":
		return 3, false, nil
	case "leave", "none", "remove", "0", "off", "exit":
		return 0, true, nil
	default:
		return 0, false, fmt.Errorf("invalid hypesquad house %q: valid options are bravery (1), brilliance (2), balance (3), or leave (0)", val)
	}
}

// MaskToken returns a safe, redacted representation of the Discord token for logs.
func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 12 {
		return "******"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

// ResolveToken discovers the Discord token from CLI, environment variable,
// encrypted vault, or .token files.
func ResolveToken(cliToken string) (string, error) {
	if cliToken != "" {
		return strings.Trim(strings.TrimSpace(cliToken), "\"' \r\n\t"), nil
	}

	if envToken := os.Getenv("DISCORD_TOKEN"); envToken != "" {
		return strings.Trim(strings.TrimSpace(envToken), "\"' \r\n\t"), nil
	}

	cacheDir := DefaultCacheDir()
	if token, ok := readTokenVault(cacheDir); ok {
		return token, nil
	}

	// Determine executable directory for portable lookup
	exePath, err := os.Executable()
	exeDir := ""
	if err == nil {
		exeDir = filepath.Dir(exePath)
	}

	// Working directory
	cwd, _ := os.Getwd()

	candidates := []string{
		".token",
		filepath.Join(cwd, ".token"),
		filepath.Join(cwd, "..", ".token"),
	}

	if exeDir != "" {
		candidates = append(candidates,
			filepath.Join(exeDir, ".token"),
			filepath.Join(exeDir, "..", ".token"),
		)
	}

	candidates = append(candidates,
		filepath.Join(cacheDir, ".token"),
	)

	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".config", "discord-quest-completer", ".token"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			token := strings.Trim(strings.TrimSpace(string(data)), "\"' \r\n\t")
			if token != "" {
				migrateTokenToVault(cacheDir, token)
				return token, nil
			}
		}
	}

	// Interactive Console Prompt: Ask user to paste token if not found
	fmt.Println("==================================================================")
	fmt.Println(" [!] DISCORD TOKEN NOT FOUND / CHƯA TÌM THẤY DISCORD TOKEN!")
	fmt.Println(" [!] Please paste your Discord Token below / Dán Token vào đây:")
	fmt.Println("==================================================================")
	fmt.Print(" > Token: ")
	reader := bufio.NewReader(os.Stdin)
	input, errInput := reader.ReadString('\n')
	if errInput == nil {
		token := strings.Trim(strings.TrimSpace(input), "\"' \r\n\t")
		if len(token) > 20 {
			if writeTokenVault(cacheDir, token) == nil {
				fmt.Printf(i18n.M().TokenSaved, tokenVaultPath(cacheDir))
				return token, nil
			}
			targetFile := ".token"
			if exeDir != "" {
				targetFile = filepath.Join(exeDir, ".token")
			}
			_ = os.WriteFile(targetFile, []byte(token), 0600)
			fmt.Printf(i18n.M().TokenSaved, targetFile)
			return token, nil
		}
	}

	return "", fmt.Errorf("%s", i18n.M().TokenError)
}

// DefaultCacheDir returns the per-user cache directory used for profiles,
// stubs and the encrypted token vault.
func DefaultCacheDir() string {
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return filepath.Join(localAppData, "DiscordQuestCompleter")
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".cache", "discord-quest-completer")
	}
	return filepath.Join(os.TempDir(), "discord-quest-completer")
}

func tokenVaultPath(cacheDir string) string {
	return filepath.Join(cacheDir, "token.bin")
}

func readTokenVault(cacheDir string) (string, bool) {
	sealed, err := os.ReadFile(tokenVaultPath(cacheDir))
	if err != nil {
		return "", false
	}
	plain, err := identity.Unprotect(sealed)
	if err != nil {
		return "", false
	}
	token := strings.Trim(strings.TrimSpace(string(plain)), "\"' \r\n\t")
	if token == "" {
		return "", false
	}
	return token, true
}

func writeTokenVault(cacheDir, token string) error {
	if !identity.ProtectionStrong() {
		return fmt.Errorf("no per-user secret store on this platform")
	}
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return err
	}
	sealed, err := identity.Protect([]byte(token))
	if err != nil {
		return err
	}
	temp := tokenVaultPath(cacheDir) + ".tmp"
	if err := os.WriteFile(temp, sealed, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, tokenVaultPath(cacheDir))
}

// migrateTokenToVault seals a plaintext token once, without deleting the
// original file: removing user files is the user's decision, not the tool's.
func migrateTokenToVault(cacheDir, token string) {
	if !identity.ProtectionStrong() {
		return
	}
	if _, err := os.Stat(tokenVaultPath(cacheDir)); err == nil {
		return
	}
	if err := writeTokenVault(cacheDir, token); err != nil {
		return
	}
	fmt.Printf("[i] Token sealed with %s at %s\n", identity.ProtectionLabel(), tokenVaultPath(cacheDir))
	fmt.Println("    The old plaintext .token file was left in place; delete it yourself when convenient.")
}

// LoadConfig parses command-line arguments and constructs the active Config.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		APIBase: "https://discord.com/api/v9",
	}

	var tokenFlag string
	var autoAcceptFlag, autoEnrollFlag bool
	var maxBadgeFlag, badgesFlag bool
	flag.StringVar(&tokenFlag, "token", "", "Discord authorization token")
	flag.DurationVar(&cfg.PollInterval, "poll", 60*time.Second, "Poll interval for quest monitoring")
	flag.BoolVar(&autoAcceptFlag, "auto-accept", true, "Auto-enroll in eligible quests")
	flag.BoolVar(&autoEnrollFlag, "auto-enroll", true, "Auto-enroll in eligible quests (alias for -auto-accept)")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable verbose debug output")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Discover and display quests without launching spoofer or enrolling")
	flag.StringVar(&cfg.CacheDir, "cache-dir", "", "Custom cache directory for detectable apps and stubs")
	flag.StringVar(&cfg.QuestID, "quest-id", "", "Filter execution to a specific Quest ID")
	flag.DurationVar(&cfg.Duration, "duration", 0, "Override maximum spoofing duration (e.g. 15m, 30s; 0 for full quest requirement)")
	flag.BoolVar(&cfg.UseSpoofer, "spoofer", false, "Enable OS Process Spoofer mode (requires Discord Desktop running)")
	flag.BoolVar(&cfg.KeepOpen, "keep-open", true, "Wait for keypress before exiting (prevents instant window closing on double-click)")
	flag.StringVar(&cfg.Lang, "lang", "auto", "Language / Ngôn ngữ (auto, en, vi)")
	flag.StringVar(&cfg.Region, "region", "all", "Quest sweep scope: all (local egress + every configured proxy), direct, full (also all locales), a country code / proxy code list, or a Discord locale")
	flag.StringVar(&cfg.Proxy, "proxy", "", "Regional proxy egress entries: optional CODE=URL pairs separated by commas (http, https, socks5)")
	flag.StringVar(&cfg.ProxyFile, "proxy-file", "", "File with one optional CODE=proxyURL per line for the regional sweep")
	flag.StringVar(&cfg.LocaleSweep, "locale-sweep", "off", "Locale header sweep: off, trio or all. Locale headers do not change quest eligibility; IP egress does")
	flag.DurationVar(&cfg.SweepPacing, "sweep-pacing", 3*time.Second, "Minimum delay between quest discovery requests; auto-raised when Discord returns HTTP 429")
	flag.IntVar(&cfg.SweepAttempts, "sweep-attempts", 3, "Attempts per sweep target in the first pass; a recovery pass adds one more attempt for failed targets")
	flag.BoolVar(&cfg.EnablePortal, "portal", true, "Enable local captcha web portal for headless environments")
	flag.IntVar(&cfg.PortalPort, "portal-port", 8080, "Port for local captcha web portal")
	flag.IntVar(&cfg.Concurrency, "concurrency", 5, "Maximum concurrent quests to process in parallel (default: 5)")
	flag.BoolVar(&cfg.Daemon, "daemon", false, "Run continuously as a background daemon/service, re-checking quests periodically")
	flag.StringVar(&cfg.HypeSquad, "hypesquad", "", "HypeSquad House: bravery (1), brilliance (2), balance (3), leave (0)")
	flag.BoolVar(&maxBadgeFlag, "max-badge", true, "Auto-claim HypeSquad House badge if missing (default: true)")
	flag.BoolVar(&badgesFlag, "badges", true, "Alias for -max-badge")
	flag.StringVar(&cfg.Fingerprint, "fingerprint", "auto", "Device fingerprint source: auto (derive from the installed Discord client) or legacy (fixed built-in values)")
	flag.Float64Var(&cfg.Jitter, "jitter", 0.15, "Timing jitter as a fraction of each delay, e.g. 0.15 for +-15% (0 disables jitter)")
	flag.BoolVar(&cfg.ProfileStore, "profile-store", true, "Store one stable device profile per account, sealed with DPAPI on Windows")
	flag.BoolVar(&cfg.PresenceGuard, "presence-guard", true, "Pause desktop-game spoofing while you are actually playing another game or using a full-screen app")

	// Internal stub mode flags (used when spawned as a spoofed game process)
	flag.BoolVar(&cfg.RunStub, "stub", false, "Internal stub mode: run as dummy game process")
	flag.StringVar(&cfg.StubTitle, "title", "Discord Game Stub", "Window title / game name for stub process")

	flag.Parse()

	// Initialize i18n system according to CLI flag, env, or OS locale
	i18n.InitLanguage(cfg.Lang)

	// If in internal stub mode, token resolution is not required
	if cfg.RunStub {
		return cfg, nil
	}

	// Auto-enroll is enabled if neither flag disabled it
	cfg.AutoAccept = autoAcceptFlag && autoEnrollFlag
	// MaxBadge is enabled if neither flag disabled it
	cfg.MaxBadge = maxBadgeFlag && badgesFlag

	token, err := ResolveToken(tokenFlag)
	if err != nil {
		return nil, err
	}
	cfg.Token = token

	if cfg.CacheDir == "" {
		cfg.CacheDir = DefaultCacheDir()
	}

	return cfg, nil
}
