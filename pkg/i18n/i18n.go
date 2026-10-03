package i18n

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Language string

const (
	LangEN Language = "en"
	LangVI Language = "vi"
)

// Messages contains all localizable user-facing strings across the application.
type Messages struct {
	BannerSubtitle        string
	ModeAPIRunner         string
	ModeSpoofer           string
	PromptExitError       string
	PromptExitKey         string
	NoEligibleQuests      string
	FoundEligibleQuests   string
	RemainingQuestSec     string
	ProcessingQuestHeader string
	SkipUnsupportedTask   string
	QuestErrorSkip        string
	VideoStart            string
	VideoProgressWarning  string
	VideoProgress         string
	VideoCompleted        string
	HeartbeatStart        string
	HeartbeatWarning      string
	HeartbeatProgress     string
	HeartbeatCompleted    string
	ActivityStart         string
	ActivityWarning       string
	ActivityProgress      string
	ActivityCompleted     string
	SpooferNoEligible     string
	SpooferHeader         string
	SpooferNoExe          string
	SpooferLaunchError    string
	SpooferProcCreated    string
	SpooferProgress       string
	SpooferCompleted      string
	TokenSaved            string
	TokenError            string
	AutoEnrollHeader      string
	AutoEnrollError       string
	AutoEnrollSuccess     string
	AutoEnrollRefresh     string
	AutoEnrollAllEnrolled string
	TableInventory        string
	TableBreakdown        string
	TableHeaderID         string
	TableHeaderTitle      string
	TableHeaderCategory   string
	TableHeaderTaskType   string
	TableHeaderProgress   string
	TableHeaderState      string
	CatPurePC             string
	CatCrossPlatform      string
	CatVideoDesktop       string
	CatVideoMobile        string
	CatActivity           string
	CatOther              string
	ScrapingCDN           string
	ActiveBuild           string
	Authenticating        string
	AuthFailed            string
	AuthSuccess           string
	DiscoveringQuests     string
	FailedRetrieveQuests  string
	FailedRefreshQuests   string
	DryRunNotice          string
	AllTasksFinished      string
	RegionScanStart       string
	RegionScanDone        string
	CaptchaAlertTerminal  string
	CaptchaPortalLink     string
	CaptchaDeepLink       string
	CaptchaWaiting        string
	CaptchaSolvedSuccess  string
	CaptchaTimeoutWarning string
	CaptchaPortalDisabled string
}

var catalogs = map[Language]Messages{
	LangEN: {
		BannerSubtitle:        "Autonomous API Runner & OS Process Spoofer for Windows & Linux",
		ModeAPIRunner:         "Autonomous API Runner (Default, Discord-Independent)",
		ModeSpoofer:           "OS Process Spoofer (-spoofer, Emulates Game Process)",
		PromptExitError:       "\n[!] An error occurred.",
		PromptExitKey:         "\n[?] Press Enter to exit...",
		NoEligibleQuests:      "[+] No quests need completion at this time (All active quests are either completed or claimed)!",
		FoundEligibleQuests:   "\n[+] Found %d quest(s) to complete:\n",
		RemainingQuestSec:     "Remaining: ~%.0fs / %ds",
		ProcessingQuestHeader: ">>> Processing [%d/%d]: %s (%s) <<<",
		SkipUnsupportedTask:   "[!] Skipping quest %s: Task type %s is not supported for automation.\n",
		QuestErrorSkip:        "[-] Error processing %s: %v. Moving to next quest...\n",
		VideoStart:            "[🎬 Video] Starting video playback: %s (Progress: %.0f/%.0fs)\n",
		VideoProgressWarning:  "    [-] Video progress warning: %v\n",
		VideoProgress:         "    [🎬] Progress: %.0f/%.0fs (%.0f%%)\n",
		VideoCompleted:        "[✅] Video quest completed successfully: %s\n",
		HeartbeatStart:        "[🎮 Game Heartbeat] Starting play simulation: %s (~%.0f min remaining)\n",
		HeartbeatWarning:      "    [-] Heartbeat warning: %v\n",
		HeartbeatProgress:     "    [🎮] [%s] Progress: %.0f/%.0fs (%.1f%%)\n",
		HeartbeatCompleted:    "[✅] Play Quest completed: %s\n",
		ActivityStart:         "[🕹️ Activity] Starting Activity: %s\n",
		ActivityWarning:       "    [-] Activity heartbeat warning: %v\n",
		ActivityProgress:      "    [🕹️] Activity Progress: %.0f/%.0fs\n",
		ActivityCompleted:     "[✅] Activity completed: %s\n",
		SpooferNoEligible:     "[+] No eligible PLAY_ON_DESKTOP quests found for Spoofer mode.",
		SpooferHeader:         "\n>>> Spoofer Quest [%d/%d]: %s <<<\n",
		SpooferNoExe:          "[-] Executable not found for AppID %s: %v\n",
		SpooferLaunchError:    "[-] Error launching dummy game: %v\n",
		SpooferProcCreated:    "[+] Created dummy game process: %s (PID: %d)\n",
		SpooferProgress:       "[★] Progress: %.0fs / %ds (State: %s)\n",
		SpooferCompleted:      "[✔] Spoofer Quest completed: %s!\n",
		TokenSaved:            "[+] Discord token successfully saved to %s!\n\n",
		TokenError:            "no valid Discord token found. Please provide via .token file or -token flag",
		AutoEnrollHeader:      "\n[*] Checking for un-enrolled active quests...",
		AutoEnrollError:       "[-] Auto-enroll encountered error: %v\n",
		AutoEnrollSuccess:     "[+] Successfully auto-enrolled in %d new quest(s)!\n",
		AutoEnrollRefresh:     "[*] Refreshing quest inventory after enrollment...",
		AutoEnrollAllEnrolled: "[+] All eligible active quests are already enrolled.",
		TableInventory:        "\n=== Discord Quest Inventory (Total: %d) ===\n",
		TableBreakdown:        "Status Breakdown: Claimed: %d | Completed: %d | Enrolled: %d | Available: %d | Expired: %d\n",
		TableHeaderID:         "Quest ID",
		TableHeaderTitle:      "Title / Game",
		TableHeaderCategory:   "Category",
		TableHeaderTaskType:   "Task Type",
		TableHeaderProgress:   "Progress",
		TableHeaderState:      "State",
		CatPurePC:             "🖥️ Pure PC (Game)",
		CatCrossPlatform:      "🎮 Cross-Platform",
		CatVideoDesktop:       "📺 Video (Desktop)",
		CatVideoMobile:        "📱 Video (Mobile)",
		CatActivity:           "🏆 Activity/Mission",
		CatOther:              "❓ Other",
		ScrapingCDN:           "\n[*] Scraping latest Discord client_build_number from CDN...",
		ActiveBuild:           "[+] Active Build Number: %d\n",
		Authenticating:        "\n[*] Authenticating with Discord API (GET /api/v9/users/@me)...",
		AuthFailed:            "[-] Authentication failed: %v\n",
		AuthSuccess:           "[+] Authentication Successful!\n",
		DiscoveringQuests:     "\n[*] Discovering Active Quests (GET /api/v9/quests/@me)...",
		FailedRetrieveQuests:  "[-] Failed to retrieve quests: %v\n",
		FailedRefreshQuests:   "[-] Failed to refresh quests after enrollment: %v\n",
		DryRunNotice:          "\n[!] Dry-run enabled: scan completed successfully. No actions taken.",
		AllTasksFinished:      "\n[+] All tasks finished.",
		RegionScanStart:        "\n[*] Multi-Region Sweep: Scanning regions [%s] to reveal hidden quests & frames...",
		RegionScanDone:         "[+] Region scan complete: Found %d unique quest(s) across all probed locales.\n",
		CaptchaAlertTerminal:   "\n[⚠️ CAPTCHA REQUIRED] Discord requires human verification for quest '%s' (%s)!\n",
		CaptchaPortalLink:      "    👉 Open verification portal on your phone or browser (same LAN WiFi):\n       http://%s:%d/?quest=%s\n",
		CaptchaDeepLink:        "    👉 Or open directly in official Discord app: https://discord.com/quests/%s\n",
		CaptchaWaiting:         "    [⏳] Awaiting captcha solution (timeout 3m)...",
		CaptchaSolvedSuccess:   "\n    [✅] Captcha solved successfully! Enrolling into quest '%s'...\n",
		CaptchaTimeoutWarning:  "\n    [-] Captcha solve timed out or skipped. Skipping this quest for now.\n",
		CaptchaPortalDisabled:  "    [!] Captcha portal is disabled. Please accept this quest manually in Discord.\n",
	},
	LangVI: {
		BannerSubtitle:        "Chạy API Tự Động & Giả Lập Tiến Trình Game Cho Windows & Linux",
		ModeAPIRunner:         "Chạy API Tự Động (Mặc định, Độc lập với Discord App)",
		ModeSpoofer:           "Giả Lập Tiến Trình Game Hệ Điều Hành (-spoofer)",
		PromptExitError:       "\n[!] Đã xảy ra lỗi.",
		PromptExitKey:         "\n[?] Nhấn phím Enter để thoát...",
		NoEligibleQuests:      "[+] Không có quest nào cần hoàn thành lúc này (Tất cả quest đang hoạt động đã hoàn tất hoặc đã nhận thưởng)!",
		FoundEligibleQuests:   "\n[+] Tìm thấy %d quest cần hoàn thành:\n",
		RemainingQuestSec:     "Còn lại: ~%.0fs / %ds",
		ProcessingQuestHeader: ">>> Đang xử lý [%d/%d]: %s (%s) <<<",
		SkipUnsupportedTask:   "[!] Bỏ qua quest %s: Loại task %s chưa được hỗ trợ tự động.\n",
		QuestErrorSkip:        "[-] Lỗi khi xử lý %s: %v. Chuyển sang quest tiếp theo...\n",
		VideoStart:            "[🎬 Video] Bắt đầu phát video: %s (Tiến độ: %.0f/%.0fs)\n",
		VideoProgressWarning:  "    [-] Cảnh báo video progress: %v\n",
		VideoProgress:         "    [🎬] Tiến độ: %.0f/%.0fs (%.0f%%)\n",
		VideoCompleted:        "[✅] Hoàn thành xuất sắc video quest: %s\n",
		HeartbeatStart:        "[🎮 Game Heartbeat] Bắt đầu gửi tín hiệu chơi: %s (~%.0f phút còn lại)\n",
		HeartbeatWarning:      "    [-] Cảnh báo heartbeat: %v\n",
		HeartbeatProgress:     "    [🎮] [%s] Tiến độ: %.0f/%.0fs (%.1f%%)\n",
		HeartbeatCompleted:    "[✅] Hoàn thành Play Quest: %s\n",
		ActivityStart:         "[🕹️ Activity] Bắt đầu Activity: %s\n",
		ActivityWarning:       "    [-] Cảnh báo activity heartbeat: %v\n",
		ActivityProgress:      "    [🕹️] Tiến độ Activity: %.0f/%.0fs\n",
		ActivityCompleted:     "[✅] Hoàn thành Activity: %s\n",
		SpooferNoEligible:     "[+] Không có quest PLAY_ON_DESKTOP nào đủ điều kiện để chạy Spoofer.",
		SpooferHeader:         "\n>>> Spoofer Quest [%d/%d]: %s <<<\n",
		SpooferNoExe:          "[-] Không tìm thấy executable cho AppID %s: %v\n",
		SpooferLaunchError:    "[-] Lỗi khởi động dummy game: %v\n",
		SpooferProcCreated:    "[+] Đã tạo tiến trình game ảo: %s (PID: %d)\n",
		SpooferProgress:       "[★] Tiến độ: %.0fs / %ds (State: %s)\n",
		SpooferCompleted:      "[✔] Hoàn thành Spoofer Quest: %s!\n",
		TokenSaved:            "[+] Đã lưu token vào file %s thành công!\n\n",
		TokenError:            "không tìm thấy Discord token hợp lệ. Vui lòng cung cấp qua file .token hoặc cờ -token",
		AutoEnrollHeader:      "\n[*] Đang kiểm tra các quest chưa tham gia...",
		AutoEnrollError:       "[-] Lỗi khi tự động nhận quest: %v\n",
		AutoEnrollSuccess:     "[+] Đã tự động tham gia thành công %d quest mới!\n",
		AutoEnrollRefresh:     "[*] Đang làm mới danh sách quest sau khi nhận...",
		AutoEnrollAllEnrolled: "[+] Tất cả quest hợp lệ đều đã được tham gia.",
		TableInventory:        "\n=== Danh Mục Nhiệm Vụ Discord (Tổng: %d) ===\n",
		TableBreakdown:        "Phân loại trạng thái: Đã nhận quà: %d | Hoàn tất: %d | Đang chạy: %d | Khả dụng: %d | Hết hạn: %d\n",
		TableHeaderID:         "Quest ID",
		TableHeaderTitle:      "Tên Game / Nhiệm vụ",
		TableHeaderCategory:   "Phân loại",
		TableHeaderTaskType:   "Loại Task",
		TableHeaderProgress:   "Tiến độ",
		TableHeaderState:      "Trạng thái",
		CatPurePC:             "🖥️ PC Thuần (Game)",
		CatCrossPlatform:      "🎮 Đa Nền Tảng (PC/Console)",
		CatVideoDesktop:       "📺 Xem Video (PC)",
		CatVideoMobile:        "📱 Xem Video (Điện thoại)",
		CatActivity:           "🏆 Hoạt Động/Nhiệm Vụ",
		CatOther:              "❓ Khác",
		ScrapingCDN:           "\n[*] Đang cào client_build_number mới nhất từ Discord CDN...",
		ActiveBuild:           "[+] Build Number đang hoạt động: %d\n",
		Authenticating:        "\n[*] Đang xác thực với Discord API (GET /api/v9/users/@me)...",
		AuthFailed:            "[-] Xác thực thất bại: %v\n",
		AuthSuccess:           "[+] Xác thực thành công!\n",
		DiscoveringQuests:     "\n[*] Đang tìm kiếm Quest hoạt động (GET /api/v9/quests/@me)...",
		FailedRetrieveQuests:  "[-] Lấy danh sách quest thất bại: %v\n",
		FailedRefreshQuests:   "[-] Không thể làm mới danh sách quest sau khi nhận: %v\n",
		DryRunNotice:          "\n[!] Chế độ Dry-run: quét hoàn tất, không thực hiện hành động nào.",
		AllTasksFinished:      "\n[+] Tất cả nhiệm vụ đã hoàn tất.",
		RegionScanStart:        "\n[*] Quét Đa Vùng (Multi-Region Sweep): Đang quét các vùng [%s] để mở khóa quest & khung avatar bị ẩn...",
		RegionScanDone:         "[+] Quét vùng hoàn tất: Tìm thấy %d nhiệm vụ duy nhất từ các khu vực.\n",
		CaptchaAlertTerminal:   "\n[⚠️ YÊU CẦU CAPTCHA] Discord yêu cầu xác minh người thật cho quest '%s' (%s)!\n",
		CaptchaPortalLink:      "    👉 Mở cổng xác minh trên điện thoại hoặc trình duyệt (cùng mạng WiFi/LAN):\n       http://%s:%d/?quest=%s\n",
		CaptchaDeepLink:        "    👉 Hoặc mở trực tiếp trên app Discord: https://discord.com/quests/%s\n",
		CaptchaWaiting:         "    [⏳] Đang chờ giải Captcha (thời gian chờ: 3 phút)...",
		CaptchaSolvedSuccess:   "\n    [✅] Đã giải Captcha thành công! Đang tiến hành nhận quest '%s'...\n",
		CaptchaTimeoutWarning:  "\n    [-] Quá thời gian giải Captcha hoặc đã bỏ qua. Bỏ qua quest này lúc này.\n",
		CaptchaPortalDisabled:  "    [!] Cổng Captcha đang tắt. Vui lòng tự nhận quest này trên app Discord.\n",
	},
}

var (
	currentLang Language = LangEN
	mu          sync.RWMutex
)

// DetectLocale auto-detects the system's preferred language.
func DetectLocale() Language {
	for _, env := range []string{"QUEST_LANG", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := strings.ToLower(os.Getenv(env)); v != "" {
			if strings.HasPrefix(v, "vi") {
				return LangVI
			}
			if strings.HasPrefix(v, "en") {
				return LangEN
			}
		}
	}

	if osLoc := strings.ToLower(detectOSLocale()); osLoc != "" {
		if strings.HasPrefix(osLoc, "vi") {
			return LangVI
		}
	}

	return LangEN
}

// InitLanguage sets the active language. If preferred is "auto" or empty, it auto-detects.
func InitLanguage(preferred string) Language {
	mu.Lock()
	defer mu.Unlock()

	p := strings.ToLower(strings.TrimSpace(preferred))
	if p == "vi" || strings.HasPrefix(p, "vi") {
		currentLang = LangVI
		return currentLang
	}
	if p == "en" || strings.HasPrefix(p, "en") {
		currentLang = LangEN
		return currentLang
	}

	currentLang = DetectLocale()
	return currentLang
}

// GetLanguage returns the active language.
func GetLanguage() Language {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// SetLanguage directly sets the active language.
func SetLanguage(lang Language) {
	mu.Lock()
	defer mu.Unlock()
	currentLang = lang
}

// M returns the current active catalog messages.
func M() Messages {
	mu.RLock()
	defer mu.RUnlock()
	if cat, ok := catalogs[currentLang]; ok {
		return cat
	}
	return catalogs[LangEN]
}

// T formats a message string using the active language catalog.
func T(msgSelector func(m Messages) string, args ...any) string {
	cat := M()
	tmpl := msgSelector(cat)
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}
