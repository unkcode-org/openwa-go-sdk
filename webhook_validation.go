package openwa

import (
	"strings"
	"unicode/utf8"
)

func validWebhookOptions(events []string, secret string, retryCount *int) bool {
	if events != nil && (len(events) == 0 || len(events) > 27) {
		return false
	}
	for _, event := range events {
		if strings.TrimSpace(event) == "" {
			return false
		}
	}
	if secret != "" && (!utf8.ValidString(secret) || utf8.RuneCountInString(secret) < 16 || utf8.RuneCountInString(secret) > 255) {
		return false
	}
	return retryCount == nil || (*retryCount >= 0 && *retryCount <= 5)
}
