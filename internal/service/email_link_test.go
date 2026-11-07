package service

import (
	"net/url"
	"testing"
)

func TestEmailLink(t *testing.T) {
	for _, tc := range []struct{ name, emailURL, appURL, want string }{
		{"public default", "", "", "https://medifinder.online/reset-password?token=a%2Bb"},
		{"local frontend", "", "http://localhost:3000", "http://localhost:3000/reset-password?token=a%2Bb"},
		{"email override", "https://medifinder.online/", "http://localhost:3000", "https://medifinder.online/reset-password?token=a%2Bb"},
		{"reject insecure public URL", "http://medifinder.online", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("EMAIL_APP_URL", tc.emailURL)
			t.Setenv("APP_URL", tc.appURL)
			got, err := emailLink("/reset-password", url.Values{"token": {"a+b"}})
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("emailLink = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
