//go:build !windows

package i18n

func detectOSLocale() string {
	return ""
}
