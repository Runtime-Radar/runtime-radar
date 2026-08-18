package service

import (
	"testing"

	"github.com/runtime-radar/runtime-radar/auth-center/api"
)

func TestMaskTokens(t *testing.T) {
	resp := &api.SignInResp{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access.signature",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.signature",
		TokenType:    "Bearer",
	}

	masked := maskTokens(resp)

	if masked.AccessToken != maskedToken {
		t.Fatalf("Expected access token to be masked, got %q", masked.AccessToken)
	}
	if masked.RefreshToken != maskedToken {
		t.Fatalf("Expected refresh token to be masked, got %q", masked.RefreshToken)
	}
	if masked.TokenType != "Bearer" {
		t.Fatalf("Expected token type to be kept, got %q", masked.TokenType)
	}

	// the caller's response must not be affected, it is the one actually returned to the user
	if resp.AccessToken == maskedToken || resp.RefreshToken == maskedToken {
		t.Fatal("Expected the original response to be left intact")
	}
}

func TestMaskTokensNil(t *testing.T) {
	if masked := maskTokens(nil); masked != nil {
		t.Fatalf("Expected nil, got %v", masked)
	}
}
