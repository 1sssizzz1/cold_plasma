package security

import (
	"strings"
	"testing"
)

func TestTextCipherEncryptsAndDecrypts(t *testing.T) {
	c, err := NewTextCipher("test-secret")
	if err != nil {
		t.Fatalf("NewTextCipher returned error: %v", err)
	}

	encrypted, err := c.Encrypt("секретный вопрос")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if !strings.HasPrefix(encrypted, encryptedTextPrefix) {
		t.Fatalf("expected encrypted prefix, got %q", encrypted)
	}
	if strings.Contains(encrypted, "секретный вопрос") {
		t.Fatalf("encrypted value contains plaintext")
	}
	if got := c.DecryptIfEncrypted(encrypted); got != "секретный вопрос" {
		t.Fatalf("unexpected decrypted value: %q", got)
	}
}

func TestTextCipherKeepsLegacyPlainText(t *testing.T) {
	c, err := NewTextCipher("test-secret")
	if err != nil {
		t.Fatalf("NewTextCipher returned error: %v", err)
	}
	if got := c.DecryptIfEncrypted("старый лог"); got != "старый лог" {
		t.Fatalf("unexpected legacy value: %q", got)
	}
}
