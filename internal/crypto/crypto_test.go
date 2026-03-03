package crypto

import (
	"crypto/sha256"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key, err := DeriveKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key))
	}
	// Deterministic
	key2, _ := DeriveKey()
	for i := range key {
		if key[i] != key2[i] {
			t.Fatal("DeriveKey should be deterministic")
		}
	}
}

func testKey() []byte {
	h := sha256.Sum256([]byte("test-key-material"))
	return h[:]
}

func TestEncryptDecrypt(t *testing.T) {
	key := testKey()
	cases := []string{"my_secret_password", "", "short", "a longer string with spaces and unicode: 你好世界"}
	for _, tc := range cases {
		ct, err := Encrypt(tc, key)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", tc, err)
		}
		pt, err := Decrypt(ct, key)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if pt != tc {
			t.Errorf("roundtrip failed: got %q, want %q", pt, tc)
		}
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := testKey()
	key2 := sha256.Sum256([]byte("different-key"))

	ct, _ := Encrypt("secret", key1)
	_, err := Decrypt(ct, key2[:])
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestEncryptDifferentNonce(t *testing.T) {
	key := testKey()
	ct1, _ := Encrypt("same", key)
	ct2, _ := Encrypt("same", key)
	if ct1 == ct2 {
		t.Error("two encryptions of same plaintext should differ (random nonce)")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := testKey()
	if _, err := Decrypt("not_valid_base64!!!", key); err == nil {
		t.Error("expected error for invalid base64")
	}
	if _, err := Decrypt("dG9vc2hvcnQ=", key); err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}
