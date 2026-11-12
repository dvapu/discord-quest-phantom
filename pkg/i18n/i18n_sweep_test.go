package i18n

import "testing"

func TestSweepMessagesPresentInEveryLanguage(t *testing.T) {
	for language, catalog := range catalogs {
		if catalog.SweepStart == "" || catalog.SweepSummary == "" || catalog.SweepTargetsFailed == "" ||
			catalog.ExcludedHint == "" || catalog.CacheFallbackNote == "" || catalog.LocaleSweepNote == "" ||
			catalog.AccessSuspended == "" || catalog.EnrollBlocked == "" {
			t.Fatalf("language %s is missing one of the sweep/gate messages", language)
		}
	}
}
