package openwa_test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"time"

	openwa "github.com/unkcode-org/openwa-go-sdk"
)

// This example compiles but does not run: it needs a real, authorized session.
func ExampleClient_SendDocument() {
	client, err := openwa.NewClient(openwa.Config{
		BaseURL: os.Getenv("OPENWA_BASE_URL"),
		APIKey:  os.Getenv("OPENWA_API_KEY"),
	})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err = client.SendDocument(ctx, os.Getenv("OPENWA_SESSION_ID"), openwa.DocumentMessage{
		ChatID:   "5493510000000@c.us",
		Base64:   base64.StdEncoding.EncodeToString([]byte("PDF bytes supplied by the application")),
		MIMEType: "application/pdf",
		Filename: "factura.pdf",
		Caption:  "Tu factura",
	})
	if errors.Is(err, openwa.ErrOutcomeUnknown) {
		// Persist the ambiguous attempt and reconcile it; do not blindly resend.
		return
	}
}
