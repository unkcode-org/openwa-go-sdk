package openwa

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Verify that the supported wire surface actually exists in the pinned contract.
// This is intentionally not a claim that the entire OpenWA API is implemented.
func TestPinnedContractAndSupportedWireTypes(t *testing.T) {
	data, err := os.ReadFile("testdata/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != ContractSHA256 {
		t.Fatal("contract snapshot changed; review it explicitly")
	}
	var spec struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Info.Version != ContractVersion {
		t.Fatal("wrong contract version")
	}
	for _, op := range []struct{ path, method string }{
		{"/api/sessions/{sessionId}", "get"},
		{"/api/sessions/{sessionId}/start", "post"},
		{"/api/sessions/{sessionId}/qr", "get"},
		{"/api/sessions/{sessionId}/contacts/check/{number}", "get"},
		{"/api/sessions/{sessionId}/messages/send-text", "post"},
		{"/api/sessions/{sessionId}/messages/send-document", "post"},
	} {
		if spec.Paths[op.path][op.method] == nil {
			t.Errorf("operation absent from contract: %s %s", op.method, op.path)
		}
	}
	for _, tc := range []struct {
		schema string
		value  any
	}{
		{"SessionResponseDto", Session{}},
		{"AccountRestrictionDto", AccountRestriction{}},
		{"QRCodeResponseDto", QRCode{}},
		{"NumberCheckResponseDto", NumberCheckResult{}},
		{"SendTextMessageDto", TextMessage{}},
		{"SendMediaMessageDto", DocumentMessage{}},
		{"CustomLinkPreviewDto", CustomLinkPreview{}},
		{"MessageResponseDto", MessageResult{}},
	} {
		typ := reflect.TypeOf(tc.value)
		for i := 0; i < typ.NumField(); i++ {
			name := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
			if spec.Components.Schemas[tc.schema].Properties[name] == nil {
				t.Errorf("%s.%s is not in contract", tc.schema, name)
			}
		}
	}
}
