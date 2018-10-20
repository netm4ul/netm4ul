package security

import "testing"

func TestGenerateNewToken(t *testing.T) {
	if got := GenerateNewToken(); len(got) != 40 {
		t.Errorf("GenerateNewToken() not generating 40 hex long token (20 bytes): got %s", got)
	}
}
