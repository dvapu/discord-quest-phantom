package scanner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/captcha"
	"discord-quest-completer/pkg/i18n"
)

// QuestState represents the current user status for a quest.
type QuestState string

const (
	StateClaimed   QuestState = "CLAIMED"
	StateCompleted QuestState = "COMPLETED"
	StateEnrolled  QuestState = "ENROLLED"
	StateAvailable QuestState = "AVAILABLE"
	StateExpired   QuestState = "EXPIRED"
)

// QuestCategory classifies the quest platform and interaction mode.
type QuestCategory string

const (
	CategoryPurePC        QuestCategory = "🖥️ Pure PC (Game)"
	CategoryCrossPlatform QuestCategory = "🎮 Cross-Platform"
	CategoryVideoDesktop  QuestCategory = "📺 Video (Desktop)"
	CategoryVideoMobile   QuestCategory = "📱 Video (Mobile)"
	CategoryActivity      QuestCategory = "🏆 Activity/Mission"
	CategoryUnknown       QuestCategory = "❓ Other"
)

// Localized returns the localized category string.
func (c QuestCategory) Localized() string {
	switch c {
	case CategoryPurePC:
		return i18n.M().CatPurePC
	case CategoryCrossPlatform:
		return i18n.M().CatCrossPlatform
	case CategoryVideoDesktop:
		return i18n.M().CatVideoDesktop
	case CategoryVideoMobile:
		return i18n.M().CatVideoMobile
	case CategoryActivity:
		return i18n.M().CatActivity
	default:
		return i18n.M().CatOther
	}
}

// AnalyzedQuest summarizes quest requirements and user progression.
type AnalyzedQuest struct {
	Quest      api.Quest
	ID         string
	Title      string
	Category   QuestCategory
	TaskType   string
	AppID      string
	TargetSec  int
	CurrentSec float64
	State      QuestState
	IsExpired  bool
	ExpiresAt  string
	Region     string
}

// AnalyzeQuest inspects an api.Quest and classifies its state and primary requirement.
func AnalyzeQuest(q api.Quest) AnalyzedQuest {
	reg := q.DiscoveredRegion
	if reg == "" {
		reg = "US"
	}
	aq := AnalyzedQuest{
		Quest:     q,
		ID:        q.ID,
		Title:     extractTitle(q),
		Region:    reg,
		State:     StateAvailable,
		Category:  CategoryUnknown,
		ExpiresAt: q.Config.ExpiresAt,
	}

	if q.UserStatus != nil {
		if q.UserStatus.ClaimedAt != nil && *q.UserStatus.ClaimedAt != "" {
			aq.State = StateClaimed
		} else if q.UserStatus.CompletedAt != nil && *q.UserStatus.CompletedAt != "" {
			aq.State = StateCompleted
		} else if q.UserStatus.EnrolledAt != nil && *q.UserStatus.EnrolledAt != "" {
			aq.State = StateEnrolled
		}
	}

	// Expiry verification
	if aq.ExpiresAt != "" {
		if expTime, err := time.Parse(time.RFC3339, aq.ExpiresAt); err == nil {
			if time.Now().UTC().After(expTime) {
				aq.IsExpired = true
				if aq.State != StateClaimed && aq.State != StateCompleted {
					aq.State = StateExpired
				}
			}
		}
	}

	// Task configuration lookup (prioritizes TaskConfigV2 over TaskConfig)
	tc := q.Config.TaskConfigV2
	if tc == nil {
		tc = q.Config.TaskConfig
	}

	if tc != nil && len(tc.Tasks) > 0 {
		_, hasMobileVideo := tc.Tasks["WATCH_VIDEO_ON_MOBILE"]
		_, hasDesktopVideo := tc.Tasks["WATCH_VIDEO"]
		_, hasDesktopPlay := tc.Tasks["PLAY_ON_DESKTOP"]
		_, hasXbox := tc.Tasks["PLAY_ON_XBOX"]
		_, hasPS := tc.Tasks["PLAY_ON_PLAYSTATION"]
		_, hasActivity := tc.Tasks["PLAY_ACTIVITY"]
		_, hasAchievement := tc.Tasks["ACHIEVEMENT_IN_ACTIVITY"]

		chosenType := ""

		if hasMobileVideo && !hasDesktopVideo {
			aq.Category = CategoryVideoMobile
			chosenType = "WATCH_VIDEO_ON_MOBILE"
		} else if hasDesktopVideo || (hasDesktopVideo && hasMobileVideo) {
			aq.Category = CategoryVideoDesktop
			chosenType = "WATCH_VIDEO"
			if !hasDesktopVideo {
				chosenType = "WATCH_VIDEO_ON_MOBILE"
			}
		} else if hasDesktopPlay {
			if hasXbox || hasPS {
				aq.Category = CategoryCrossPlatform
			} else {
				aq.Category = CategoryPurePC
			}
			chosenType = "PLAY_ON_DESKTOP"
		} else if hasActivity || hasAchievement {
			aq.Category = CategoryActivity
			if hasActivity {
				chosenType = "PLAY_ACTIVITY"
			} else {
				chosenType = "ACHIEVEMENT_IN_ACTIVITY"
			}
		} else {
			for tType := range tc.Tasks {
				chosenType = tType
				break
			}
			aq.Category = CategoryUnknown
		}

		if req, ok := tc.Tasks[chosenType]; ok {
			aq.TaskType = chosenType
			aq.TargetSec = req.Target
			if len(req.Applications) > 0 {
				aq.AppID = req.Applications[0].ID
			} else if q.Config.Application != nil {
				aq.AppID = q.Config.Application.ID
			}

			if q.UserStatus != nil && q.UserStatus.Progress != nil {
				if prog, hasProg := q.UserStatus.Progress[chosenType]; hasProg {
					aq.CurrentSec = prog.Value
				}
			}
		}
	}

	return aq
}

func extractTitle(q api.Quest) string {
	if q.Config.Messages.QuestName != nil && *q.Config.Messages.QuestName != "" {
		return *q.Config.Messages.QuestName
	}
	if q.Config.Messages.GameTitle != nil && *q.Config.Messages.GameTitle != "" {
		return *q.Config.Messages.GameTitle
	}
	if q.Config.Application != nil && q.Config.Application.Name != "" {
		return q.Config.Application.Name
	}
	return "Quest #" + q.ID
}

// AnalyzeAll processes a slice of raw quests into AnalyzedQuest items.
func AnalyzeAll(rawQuests []api.Quest) []AnalyzedQuest {
	analyzed := make([]AnalyzedQuest, len(rawQuests))
	for i, q := range rawQuests {
		analyzed[i] = AnalyzeQuest(q)
	}
	return analyzed
}

// AutoEnrollPending automatically enrolls into all un-enrolled active quests.
func AutoEnrollPending(ctx context.Context, client *api.Client, quests []AnalyzedQuest, enablePortal bool, portalPort int) (int, error) {
	enrolledCount := 0
	for _, q := range quests {
		if q.State != StateAvailable || q.IsExpired {
			continue
		}

		if i18n.GetLanguage() == i18n.LangVI {
			fmt.Printf(" [Tự Động Nhận] Đang tham gia quest: %s (%s) [%s]...\n", q.Title, q.ID, q.Category.Localized())
		} else {
			fmt.Printf(" [Auto-Enroll] Enrolling in quest: %s (%s) [%s]...\n", q.Title, q.ID, q.Category.Localized())
		}
		err := client.EnrollQuest(ctx, q.Quest)
		if err != nil {
			var captchaErr *api.CaptchaRequiredError
			if errors.As(err, &captchaErr) {
				fmt.Printf(i18n.M().CaptchaAlertTerminal, q.Title, q.ID)
				fmt.Printf(i18n.M().CaptchaDeepLink, q.ID)
				if enablePortal {
					activePort := captcha.FindAvailablePort(portalPort)
					outboundIP := captcha.GetOutboundIP()
					fmt.Printf(i18n.M().CaptchaPortalLink, outboundIP, activePort, q.ID)
					fmt.Print(i18n.M().CaptchaWaiting)

					portalCtx, cancelPortal := context.WithTimeout(ctx, 3*time.Minute)
					token, solveErr := captcha.StartPortal(portalCtx, activePort, captchaErr, q.Title)
					cancelPortal()

					if solveErr == nil && token != "" {
						fmt.Printf(i18n.M().CaptchaSolvedSuccess, q.Title)
						enrollErr := client.EnrollQuestWithCaptcha(ctx, q.Quest, token, captchaErr.CaptchaRqtoken)
						if enrollErr == nil {
							enrolledCount++
							continue
						}
						fmt.Printf(" [-] Error enrolling after captcha: %v\n", enrollErr)
					} else {
						fmt.Print(i18n.M().CaptchaTimeoutWarning)
					}
				} else {
					fmt.Print(i18n.M().CaptchaPortalDisabled)
				}
				continue
			}

			if i18n.GetLanguage() == i18n.LangVI {
				fmt.Printf(" [Tự Động Nhận] Cảnh báo: không thể nhận quest %s: %v\n", q.Title, err)
			} else {
				fmt.Printf(" [Auto-Enroll] Warning: failed enrolling in %s: %v\n", q.Title, err)
			}
			continue
		}
		enrolledCount++
		// Polite delay between enrollments to respect Discord burst rate limits
		select {
		case <-ctx.Done():
			return enrolledCount, ctx.Err()
		case <-time.After(2500 * time.Millisecond):
		}
	}
	return enrolledCount, nil
}

// FormatQuestTable returns a clean CLI overview of quests with categories.
func FormatQuestTable(quests []AnalyzedQuest) string {
	var sb strings.Builder

	var claimed, completed, enrolled, available, expired int
	for _, q := range quests {
		switch q.State {
		case StateClaimed:
			claimed++
		case StateCompleted:
			completed++
		case StateEnrolled:
			enrolled++
		case StateAvailable:
			available++
		case StateExpired:
			expired++
		}
	}

	sb.WriteString(fmt.Sprintf(i18n.M().TableInventory, len(quests)))
	sb.WriteString(fmt.Sprintf(i18n.M().TableBreakdown,
		claimed, completed, enrolled, available, expired))
	sb.WriteString(strings.Repeat("-", 116) + "\n")
	sb.WriteString(fmt.Sprintf("%-19s | %-6s | %-24s | %-24s | %-14s | %-9s | %s\n",
		i18n.M().TableHeaderID, "REGION", i18n.M().TableHeaderTitle, i18n.M().TableHeaderCategory, i18n.M().TableHeaderTaskType, i18n.M().TableHeaderProgress, i18n.M().TableHeaderState))
	sb.WriteString(strings.Repeat("-", 116) + "\n")

	for _, q := range quests {
		progressStr := fmt.Sprintf("%.0fs / %ds", q.CurrentSec, q.TargetSec)
		if q.TargetSec == 0 {
			progressStr = "N/A"
		}
		title := q.Title
		if len([]rune(title)) > 24 {
			title = string([]rune(title)[:21]) + "..."
		}
		cat := q.Category.Localized()
		if len([]rune(cat)) > 24 {
			cat = string([]rune(cat)[:21]) + "..."
		}
		reg := q.Region
		if reg == "" {
			reg = "US"
		}
		sb.WriteString(fmt.Sprintf("%-19s | %-6s | %-24s | %-24s | %-14s | %-9s | %s\n",
			q.ID, "["+reg+"]", title, cat, q.TaskType, progressStr, q.State))
	}
	sb.WriteString(strings.Repeat("-", 116) + "\n")

	return sb.String()
}
