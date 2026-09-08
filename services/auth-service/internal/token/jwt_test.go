package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"kaffein/auth-service/config"
)

func TestPairAndParse(t *testing.T) {
	manager := NewManager(config.JWTConfig{Issuer: "test", Secret: "test-secret", AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour})
	userID := uuid.New()
	access, refresh, sid, _, err := manager.Pair(userID, "admin")
	if err != nil {
		t.Fatal(err)
	}
	accessClaims, err := manager.Parse(access, "access")
	if err != nil {
		t.Fatal(err)
	}
	if accessClaims.Subject != userID.String() || accessClaims.SessionID != sid || accessClaims.Roles[0] != "admin" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}
	refreshClaims, err := manager.Parse(refresh, "refresh")
	if err != nil {
		t.Fatal(err)
	}
	if refreshClaims.SessionID != sid {
		t.Fatalf("refresh sid %s does not match %s", refreshClaims.SessionID, sid)
	}
}

func TestParseRejectsWrongTokenTypeAndIssuer(t *testing.T) {
	manager := NewManager(config.JWTConfig{Issuer: "test", Secret: "test-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour})
	access, _, _, _, err := manager.Pair(uuid.New(), "tenant")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Parse(access, "refresh"); err == nil {
		t.Fatal("expected wrong token type rejection")
	}
	otherIssuer := NewManager(config.JWTConfig{Issuer: "other", Secret: "test-secret"})
	if _, err := otherIssuer.Parse(access, "access"); err == nil {
		t.Fatal("expected wrong issuer rejection")
	}
}
