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
)

// Config represents the application runtime configuration.
type Config struct {
	Token        string
	PollInterval time.Duration
	AutoAccept   bool
	Debug        bool
	APIBase      string
	DryRun       bool
	CacheDir     string
	RunStub      bool
	StubTitle    string
	QuestID      string
	Duration     time.Duration
	UseSpoofer   bool
	KeepOpen     bool
	Lang         string
	Region       string
	EnablePortal bool
	PortalPort   int
	Concurrency  int
}

// MaskToken returns a safe, redacted representation of the Discord token for logs.
func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 12 {
		return "******"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

// ResolveToken discovers the Discord token from CLI, environment variable, or .token files.
func ResolveToken(cliToken string) (string, error) {
	if cliToken != "" {
		return strings.Trim(strings.TrimSpace(cliToken), "\"' \r\n\t"), nil
	}

	if envToken := os.Getenv("DISCORD_TOKEN"); envToken != "" {
		return strings.Trim(strings.TrimSpace(envToken), "\"' \r\n\t"), nil
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

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "DiscordQuestCompleter", ".token"))
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".config", "discord-quest-completer", ".token"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			token := strings.Trim(strings.TrimSpace(string(data)), "\"' \r\n\t")
			if token != "" {
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

// LoadConfig parses command-line arguments and constructs the active Config.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		APIBase: "https://discord.com/api/v9",
	}

	var tokenFlag string
	var autoAcceptFlag, autoEnrollFlag bool
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
	flag.StringVar(&cfg.Region, "region", "all", "Region for quest discovery: all (US+JP+VN), us, jp, vn")
	flag.BoolVar(&cfg.EnablePortal, "portal", true, "Enable local captcha web portal for headless environments")
	flag.IntVar(&cfg.PortalPort, "portal-port", 8080, "Port for local captcha web portal")
	flag.IntVar(&cfg.Concurrency, "concurrency", 5, "Maximum concurrent quests to process in parallel (default: 5)")

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

	token, err := ResolveToken(tokenFlag)
	if err != nil {
		return nil, err
	}
	cfg.Token = token

	if cfg.CacheDir == "" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			cfg.CacheDir = filepath.Join(localAppData, "DiscordQuestCompleter")
		} else if homeDir, err := os.UserHomeDir(); err == nil {
			cfg.CacheDir = filepath.Join(homeDir, ".cache", "discord-quest-completer")
		} else {
			cfg.CacheDir = filepath.Join(os.TempDir(), "discord-quest-completer")
		}
	}

	return cfg, nil
}
