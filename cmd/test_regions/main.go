package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"discord-quest-completer/pkg/api"
)

type regionTarget struct {
	Code     string
	Country  string
	Locale   string
	Timezone string
}

func getQuestTitle(q api.Quest) string {
	if q.Config.Messages.QuestName != nil && *q.Config.Messages.QuestName != "" {
		return *q.Config.Messages.QuestName
	}
	if q.Config.Messages.GameTitle != nil && *q.Config.Messages.GameTitle != "" {
		return *q.Config.Messages.GameTitle
	}
	return q.ID
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		tokenBytes, err := os.ReadFile(".token")
		if err == nil {
			token = strings.TrimSpace(string(tokenBytes))
		}
	}
	if token == "" {
		home, _ := os.UserHomeDir()
		tokenBytes, err := os.ReadFile(filepath.Join(home, ".discord-quest-phantom", ".token"))
		if err == nil {
			token = strings.TrimSpace(string(tokenBytes))
		}
	}
	if token == "" {
		fmt.Println("No token found")
		return
	}

	targets := []regionTarget{
		{Code: "US", Country: "United States", Locale: "en-US", Timezone: "America/New_York"},
		{Code: "JP", Country: "Japan", Locale: "ja-JP", Timezone: "Asia/Tokyo"},
		{Code: "VN", Country: "Vietnam", Locale: "vi-VN", Timezone: "Asia/Ho_Chi_Minh"},
		{Code: "KR", Country: "South Korea", Locale: "ko-KR", Timezone: "Asia/Seoul"},
		{Code: "CN", Country: "China", Locale: "zh-CN", Timezone: "Asia/Shanghai"},
		{Code: "TW", Country: "Taiwan", Locale: "zh-TW", Timezone: "Asia/Taipei"},
		{Code: "IN", Country: "India", Locale: "en-IN", Timezone: "Asia/Kolkata"},
		{Code: "DE", Country: "Germany / EU", Locale: "de-DE", Timezone: "Europe/Berlin"},
		{Code: "FR", Country: "France / EU", Locale: "fr-FR", Timezone: "Europe/Paris"},
		{Code: "BR", Country: "Brazil / LATAM", Locale: "pt-BR", Timezone: "America/Sao_Paulo"},
		{Code: "RU", Country: "Russia", Locale: "ru-RU", Timezone: "Europe/Moscow"},
		{Code: "TR", Country: "Turkey", Locale: "tr-TR", Timezone: "Europe/Istanbul"},
	}

	client := api.NewClient(token, 0)
	ctx := context.Background()

	allQuestsMap := make(map[string]api.Quest)
	regionQuestsMap := make(map[string][]string)

	fmt.Printf("%-5s | %-16s | %-8s | %-15s | %-11s | %s\n", "CODE", "COUNTRY", "LOCALE", "HTTP STATUS", "QUEST COUNT", "SAMPLE TITLES")
	fmt.Println(strings.Repeat("-", 100))

	for _, t := range targets {
		start := time.Now()
		quests, err := client.FetchQuestsWithLocale(ctx, t.Locale, t.Timezone)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("%-5s | %-16s | %-8s | ERROR (%v) | 0 quests\n", t.Code, t.Country, t.Locale, err)
		} else {
			var titles []string
			var qids []string
			for _, q := range quests {
				qids = append(qids, q.ID)
				if _, exists := allQuestsMap[q.ID]; !exists {
					allQuestsMap[q.ID] = q
				}
				title := getQuestTitle(q)
				if len(titles) < 3 {
					titles = append(titles, title)
				}
			}
			regionQuestsMap[t.Code] = qids
			sample := strings.Join(titles, ", ")
			if len(sample) > 40 {
				sample = sample[:37] + "..."
			}
			fmt.Printf("%-5s | %-16s | %-8s | OK (%4dms)    | %2d quests   | %s\n",
				t.Code, t.Country, t.Locale, elapsed.Milliseconds(), len(quests), sample)
		}

		time.Sleep(300 * time.Millisecond) // Polite pacing
	}

	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("Total unique quests across ALL regions: %d\n", len(allQuestsMap))

	// Find region-exclusive quests
	for code, qids := range regionQuestsMap {
		var exclusives []string
		for _, qid := range qids {
			isExclusive := true
			for otherCode, otherQids := range regionQuestsMap {
				if code == otherCode {
					continue
				}
				for _, otherQid := range otherQids {
					if qid == otherQid {
						isExclusive = false
						break
					}
				}
				if !isExclusive {
					break
				}
			}
			if isExclusive {
				q := allQuestsMap[qid]
				name := getQuestTitle(q)
				exclusives = append(exclusives, fmt.Sprintf("%s (%s)", name, qid))
			}
		}
		if len(exclusives) > 0 {
			fmt.Printf("Region [%s] EXCLUSIVE quests (%d): %s\n", code, len(exclusives), strings.Join(exclusives, "; "))
		}
	}
}
