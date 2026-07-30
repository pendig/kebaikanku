package payment

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestEncryptedSettingRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	encoded := base64.StdEncoding.EncodeToString(key)
	decoded, err := DecodeSettingsKey(encoded)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := EncryptSetting(decoded, "SB-Mid-server-secret")
	if err != nil {
		t.Fatal(err)
	}
	if ciphertext == "SB-Mid-server-secret" {
		t.Fatal("secret stored as plaintext")
	}
	plain, err := DecryptSetting(decoded, ciphertext)
	if err != nil || plain != "SB-Mid-server-secret" {
		t.Fatalf("round trip = %q, %v", plain, err)
	}
}
