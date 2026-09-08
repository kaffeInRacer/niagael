package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	client := New("server-key", "client-key", "sandbox")
	payload := "order-1" + "200" + "10000.00" + "server-key"
	sum := sha512.Sum512([]byte(payload))
	signature := hex.EncodeToString(sum[:])

	if !client.VerifySignature("order-1", "200", "10000.00", signature) {
		t.Fatal("expected a valid Midtrans signature")
	}
	if client.VerifySignature("order-1", "200", "10000.00", "invalid") {
		t.Fatal("expected an invalid Midtrans signature to be rejected")
	}
}
