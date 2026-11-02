package utils

import "testing"

func TestResetTokens(t *testing.T) {
	first, err := GenerateResetToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateResetToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 64 || first == second {
		t.Fatal("reset tokens must be unique 256-bit secrets")
	}
	hash, err := HashResetToken(first)
	if err != nil || hash == first || len(hash) != 64 {
		t.Fatal("must store a hash, not the link secret")
	}
	repeat, _ := HashResetToken(first)
	if repeat != hash {
		t.Fatal("token lookup must be deterministic")
	}
	for _, token := range []string{"", "bad", "eyJhbGciOiJIUzI1NiJ9.payload.signature", string(make([]byte, 64))} {
		if _, err := HashResetToken(token); err == nil {
			t.Fatalf("accepted malformed token %q", token)
		}
	}
}
