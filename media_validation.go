package openwa

import (
	"encoding/base64"
	"mime"
	"strings"
	"unicode/utf8"
)

// validateMedia checks the shared SendMediaMessageDto fields before dispatch.
func validateMedia(chatID, sourceURL, encoded, mimeType, caption string, limit int64, tooLarge error) error {
	if strings.TrimSpace(chatID) == "" || (encoded == "") == (sourceURL == "") || !utf8.ValidString(caption) || utf8.RuneCountInString(caption) > 1024 {
		return ErrInvalidInput
	}
	if encoded != "" {
		// Check encoded size before allocating a decoded buffer. CR/LF and data
		// URL prefixes are not accepted; callers pass standard base64 bytes.
		if strings.ContainsAny(encoded, "\r\n") {
			return ErrInvalidInput
		}
		n := int64(base64.StdEncoding.DecodedLen(len(encoded)))
		if strings.HasSuffix(encoded, "==") {
			n -= 2
		} else if strings.HasSuffix(encoded, "=") {
			n--
		}
		if n > limit {
			return tooLarge
		}
		data, err := base64.StdEncoding.Strict().DecodeString(encoded)
		if err != nil || len(data) == 0 {
			return ErrInvalidInput
		}
		if int64(len(data)) > limit {
			return tooLarge
		}
		if _, _, err := mime.ParseMediaType(mimeType); err != nil {
			return ErrInvalidInput
		}
	} else if !validHTTPURL(sourceURL) {
		return ErrInvalidInput
	}
	return nil
}
