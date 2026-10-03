package i18n

import (
	"testing"
)

func TestI18nLanguages(t *testing.T) {
	// Test English
	InitLanguage("en")
	if GetLanguage() != LangEN {
		t.Fatalf("expected LangEN, got %v", GetLanguage())
	}
	if M().CatPurePC != "🖥️ Pure PC (Game)" {
		t.Fatalf("unexpected English CatPurePC: %s", M().CatPurePC)
	}
	msgEn := T(func(m Messages) string { return m.FoundEligibleQuests }, 5)
	if msgEn != "\n[+] Found 5 quest(s) to complete:\n" {
		t.Fatalf("unexpected formatted English message: %s", msgEn)
	}

	// Test Vietnamese
	InitLanguage("vi")
	if GetLanguage() != LangVI {
		t.Fatalf("expected LangVI, got %v", GetLanguage())
	}
	if M().CatPurePC != "🖥️ PC Thuần (Game)" {
		t.Fatalf("unexpected Vietnamese CatPurePC: %s", M().CatPurePC)
	}
	msgVi := T(func(m Messages) string { return m.FoundEligibleQuests }, 3)
	if msgVi != "\n[+] Tìm thấy 3 quest cần hoàn thành:\n" {
		t.Fatalf("unexpected formatted Vietnamese message: %s", msgVi)
	}
}
