package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"discord-quest-completer/pkg/api"
)

func questTitle(q api.Quest) string {
	if q.Config.Messages.QuestName != nil && *q.Config.Messages.QuestName != "" {
		return *q.Config.Messages.QuestName
	}
	if q.Config.Messages.GameTitle != nil && *q.Config.Messages.GameTitle != "" {
		return *q.Config.Messages.GameTitle
	}
	if q.Config.Application != nil && q.Config.Application.Name != "" {
		return q.Config.Application.Name
	}
	return q.ID
}

func resolveToken(cliToken string) string {
	if cliToken != "" {
		return strings.TrimSpace(cliToken)
	}
	if env := os.Getenv("DISCORD_TOKEN"); env != "" {
		return strings.TrimSpace(env)
	}
	candidates := []string{".token"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".discord-quest-phantom", ".token"))
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		token := strings.Trim(strings.TrimSpace(string(data)), "\"' \r\n\t")
		if token != "" {
			return token
		}
	}
	return ""
}

func passiveTarget(kind api.TargetKind) string {
	if kind == api.TargetProxy {
		return "proxy"
	}
	if kind == api.TargetLocale {
		return "locale"
	}
	return "direct"
}

func main() {
	var (
		cliToken   = flag.String("token", "", "Discord token (defaults to DISCORD_TOKEN or .token)")
		region     = flag.String("region", "all", "Sweep scope: all, direct, full, a country code or a Discord locale")
		proxy      = flag.String("proxy", "", "Regional proxy entries: CODE=url pairs separated by commas")
		proxyFile  = flag.String("proxy-file", "", "File with one CODE=proxyURL per line")
		localeMode = flag.String("locale-sweep", "off", "Locale header sweep: off, trio or all")
		passes     = flag.Int("passes", 1, "Full sweep passes to run (2 exposes transient discovery gaps)")
		pacing     = flag.Duration("sweep-pacing", 3*time.Second, "Minimum delay between quest requests")
		attempts   = flag.Int("sweep-attempts", 3, "Attempts per target")
		jsonOut    = flag.String("json", "", "Write the full benchmark result to this JSON path")
	)
	flag.Parse()

	token := resolveToken(*cliToken)
	if token == "" {
		fmt.Println("[-] No Discord token found. Pass -token, set DISCORD_TOKEN or create a .token file.")
		os.Exit(2)
	}

	targets, warnings, err := api.BuildTargets(*region, *proxy, *proxyFile, *localeMode)
	if err != nil {
		fmt.Printf("[-] Invalid sweep configuration: %v\n", err)
		os.Exit(2)
	}
	for _, warning := range warnings {
		fmt.Printf("[i] %s\n", warning)
	}

	fmt.Printf("[*] Token resolved (%d chars)\n", len(token))
	client := api.NewClient(token, 0)
	ctx := context.Background()

	user, err := client.ValidateToken(ctx)
	if err != nil {
		fmt.Printf("[-] Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Authenticated as %s (%s)\n", user.Username, user.ID)
	fmt.Printf("[*] Sweep target(s): %d\n", len(targets))
	for _, target := range targets {
		proxyLabel := "direct"
		if target.ProxyURL != "" {
			proxyLabel = api.RedactProxyURL(target.ProxyURL)
		}
		fmt.Printf("    - %-8s locale=%-5s tz=%-22s %-10s %s\n",
			target.Code, target.Locale, target.Timezone, passiveTarget(target.Kind), proxyLabel)
	}

	opts := api.DefaultSweepOptions()
	if *pacing > 0 {
		opts.MinInterval = *pacing
	}
	if *attempts > 0 {
		opts.MaxAttempts = *attempts
	}
	if *passes < 1 {
		*passes = 1
	}

	type passSnapshot struct {
		At     time.Time
		Result *api.SweepResult
		Quests map[string]api.Quest
		IDs    []string
	}
	var snapshots []passSnapshot

	for pass := 1; pass <= *passes; pass++ {
		fmt.Printf("\n=== Pass %d/%d (%d target(s)) ===\n", pass, *passes, len(targets))
		result, sweepErr := client.SweepQuests(ctx, targets, opts)
		if result == nil {
			fmt.Printf("[-] Sweep failed before returning data: %v\n", sweepErr)
			os.Exit(1)
		}
		for _, target := range result.Targets {
			proxyLabel := target.Proxy
			if proxyLabel == "" {
				proxyLabel = "direct"
			}
			if target.Ok {
				fmt.Printf("    [OK ] %-8s locale=%-5s %-24s quests=%-3d excluded=%-3d %dms\n",
					target.Code, target.Locale, proxyLabel, target.Quests, target.Excluded, target.Latency.Milliseconds())
				continue
			}
			fmt.Printf("    [ERR] %-8s locale=%-5s %-24s %s (attempts=%d)\n",
				target.Code, target.Locale, proxyLabel, target.Err, target.Attempts)
		}
		if failed := result.FailedTargets(); len(failed) > 0 {
			fmt.Printf("    [!] Failed target(s) this pass: %s\n", strings.Join(failed, ", "))
		}
		fmt.Printf("[+] Pass %d: %d quest(s), %d excluded, %d request(s) in %s\n",
			pass, len(result.Quests), result.ExcludedCount(), result.Requests, result.Duration.Round(100*time.Millisecond))
		if sweepErr != nil {
			fmt.Printf("[!] Sweep warning: %v\n", sweepErr)
		}

		quests := make(map[string]api.Quest, len(result.Quests))
		ids := make([]string, 0, len(result.Quests))
		for _, quest := range result.Quests {
			quests[quest.ID] = quest
			ids = append(ids, quest.ID)
		}
		snapshots = append(snapshots, passSnapshot{At: time.Now(), Result: result, Quests: quests, IDs: ids})
	}

	union := make(map[string]api.Quest)
	for _, snapshot := range snapshots {
		for id, quest := range snapshot.Quests {
			if _, exists := union[id]; !exists {
				union[id] = quest
			}
		}
	}
	unionIDs := make([]string, 0, len(union))
	for id := range union {
		unionIDs = append(unionIDs, id)
	}
	sort.Strings(unionIDs)

	now := time.Now().UTC()
	active := 0
	for _, quest := range union {
		if quest.Config.ExpiresAt == "" {
			active++
			continue
		}
		if expires, err := time.Parse(time.RFC3339, quest.Config.ExpiresAt); err == nil && expires.After(now) {
			active++
		}
	}

	fmt.Printf("\n=== Union across %d pass(es): %d unique quest(s), %d active (not expired) ===\n", len(snapshots), len(union), active)
	for _, id := range unionIDs {
		quest := union[id]
		expires := quest.Config.ExpiresAt
		if expires == "" {
			expires = "no-expiry"
		}
		tasks := "?"
		taskConfig := quest.Config.TaskConfigV2
		if taskConfig == nil {
			taskConfig = quest.Config.TaskConfig
		}
		if taskConfig != nil {
			keys := make([]string, 0, len(taskConfig.Tasks))
			for key := range taskConfig.Tasks {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			tasks = strings.Join(keys, "+")
		}
		fmt.Printf("    %-19s %-8s exp=%-25s tasks=%-28s %s\n", id, "["+quest.DiscoveredRegion+"]", expires, tasks, questTitle(quest))
	}

	// Region exclusives based on the first pass (only meaningful with 2+ regions).
	if len(snapshots) > 0 {
		regionSets := make(map[string]map[string]bool)
		for _, quest := range snapshots[0].Result.Quests {
			for _, code := range quest.SeenIn {
				if regionSets[code] == nil {
					regionSets[code] = make(map[string]bool)
				}
				regionSets[code][quest.ID] = true
			}
		}
		codes := make([]string, 0, len(regionSets))
		for code := range regionSets {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		if len(codes) < 2 {
			codes = nil
		}
		for _, code := range codes {
			var exclusives []string
			for id := range regionSets[code] {
				foundElsewhere := false
				for otherCode, otherSet := range regionSets {
					if otherCode == code {
						continue
					}
					if otherSet[id] {
						foundElsewhere = true
						break
					}
				}
				if !foundElsewhere {
					exclusives = append(exclusives, id+" ("+questTitle(union[id])+")")
				}
			}
			sort.Strings(exclusives)
			if len(exclusives) > 0 {
				fmt.Printf("[🌐] Region [%s] exclusive quest(s): %s\n", code, strings.Join(exclusives, "; "))
			}
		}
	}

	// Cross-pass consistency: quests that a second pass found but the first missed
	// are exactly the transient "occasionally misses a quest" failures.
	for index := 1; index < len(snapshots); index++ {
		previous := snapshots[index-1].Quests
		current := snapshots[index].Quests
		var onlyPrevious, onlyCurrent []string
		for id := range previous {
			if _, ok := current[id]; !ok {
				onlyPrevious = append(onlyPrevious, id)
			}
		}
		for id := range current {
			if _, ok := previous[id]; !ok {
				onlyCurrent = append(onlyCurrent, id+" ("+questTitle(current[id])+")")
			}
		}
		sort.Strings(onlyPrevious)
		sort.Strings(onlyCurrent)
		fmt.Printf("\n[*] Pass %d vs %d: previously-only=%d later-only=%d\n", index, index+1, len(onlyPrevious), len(onlyCurrent))
		if len(onlyCurrent) > 0 {
			fmt.Printf("    Later-only quest(s): %s\n", strings.Join(onlyCurrent, ", "))
		}
		if len(onlyPrevious) > 0 {
			fmt.Printf("    Previously-only quest(s): %s\n", strings.Join(onlyPrevious, ", "))
		}
	}

	if *jsonOut != "" {
		type reportTarget struct {
			Code     string `json:"code"`
			Locale   string `json:"locale"`
			Timezone string `json:"timezone"`
			Proxy    string `json:"proxy"`
			Ok       bool   `json:"ok"`
			Status   int    `json:"status"`
			Quests   int    `json:"quests"`
			Excluded int    `json:"excluded"`
			Error    string `json:"error,omitempty"`
			Attempts int    `json:"attempts"`
		}
		type reportPass struct {
			Quests   []string       `json:"quest_ids"`
			Targets  []reportTarget `json:"targets"`
			Requests int            `json:"requests"`
			Failed   int            `json:"failed_targets"`
		}
		type reportTargetConfig struct {
			Code     string `json:"code"`
			Locale   string `json:"locale"`
			Timezone string `json:"timezone"`
			// Proxy is always redacted; benchmark files are meant to be shared.
			Proxy string `json:"proxy"`
			Kind  string `json:"kind"`
		}
		type report struct {
			GeneratedAt string               `json:"generated_at"`
			Targets     []reportTargetConfig `json:"targets"`
			Passes      []reportPass         `json:"passes"`
			Union       []string             `json:"union_quest_ids"`
			Active      int                  `json:"active_quests"`
			Details     map[string]api.Quest `json:"quests"`
		}
		targetConfigs := make([]reportTargetConfig, 0, len(targets))
		for _, target := range targets {
			targetConfigs = append(targetConfigs, reportTargetConfig{
				Code:     target.Code,
				Locale:   target.Locale,
				Timezone: target.Timezone,
				Proxy:    api.RedactProxyURL(target.ProxyURL),
				Kind:     passiveTarget(target.Kind),
			})
		}
		export := report{
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			Targets:     targetConfigs,
			Union:       unionIDs,
			Active:      active,
			Details:     union,
		}
		for _, snapshot := range snapshots {
			passReport := reportPass{Quests: snapshot.IDs, Requests: snapshot.Result.Requests, Failed: snapshot.Result.Failed}
			for _, target := range snapshot.Result.Targets {
				passReport.Targets = append(passReport.Targets, reportTarget{
					Code: target.Code, Locale: target.Locale, Timezone: target.Timezone, Proxy: target.Proxy,
					Ok: target.Ok, Status: target.Status, Quests: target.Quests, Excluded: target.Excluded,
					Error: target.Err, Attempts: target.Attempts,
				})
			}
			export.Passes = append(export.Passes, passReport)
		}
		data, err := json.MarshalIndent(export, "", " ")
		if err != nil {
			fmt.Printf("[-] Could not encode JSON: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*jsonOut, data, 0o600); err != nil {
			fmt.Printf("[-] Could not write JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[+] Wrote benchmark JSON to %s\n", *jsonOut)
	}
}
