package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrevoAuthenticationIndependentOfTemplate(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	t.Setenv("BREVO_API_KEY", "xkeysib-test-only")
	t.Setenv("EMAIL_SENDER", "sender@example.invalid")
	body, err := ParseTemplate("templates/emails/reset_password.html", struct{ Name, ResetLink string }{
		"Test", "https://example.invalid/reset-password?token=test",
	})
	if err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	defer func() { http.DefaultTransport = original }()
	for _, content := range []string{"<p>Old email body</p>", body} {
		http.DefaultTransport = mailTransport(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("api-key") != "xkeysib-test-only" || req.Header.Get("content-type") != "application/json" {
				t.Fatal("template changed Brevo authentication headers")
			}
			var payload struct {
				HTML string `json:"htmlContent"`
			}
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil || payload.HTML != content {
				t.Fatal("email body was not preserved")
			}
			return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
		})
		if err := SendEmailNow("recipient@example.invalid", "Test", content); err != nil {
			t.Fatal(err)
		}
	}
}

type mailTransport func(*http.Request) (*http.Response, error)

func (f mailTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestBrevoAcceptanceAndFailure(t *testing.T) {
	t.Setenv("BREVO_API_KEY", "test-key")
	t.Setenv("EMAIL_SENDER", "sender@example.invalid")
	original := http.DefaultTransport
	defer func() { http.DefaultTransport = original }()
	for _, status := range []int{201, 401, 429, 500} {
		http.DefaultTransport = mailTransport(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.brevo.com/v3/smtp/email" || req.Header.Get("api-key") != "test-key" {
				t.Fatal("invalid Brevo request")
			}
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
		})
		err := SendEmailNow("recipient@example.invalid", "test", "<p>test</p>")
		if (status == 201) != (err == nil) {
			t.Fatalf("status %d produced error %v", status, err)
		}
	}
	t.Setenv("BREVO_API_KEY", "")
	if err := SendEmailNow("recipient@example.invalid", "test", "test"); err == nil {
		t.Fatal("missing key must fail")
	}
}

func TestBrevoErrorIsUsefulWithoutLeakingKey(t *testing.T) {
	t.Setenv("BREVO_API_KEY", "private-test-key")
	original := http.DefaultTransport
	defer func() { http.DefaultTransport = original }()
	http.DefaultTransport = mailTransport(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 401, Body: io.NopCloser(strings.NewReader(`{"code":"unauthorized","message":"Key not found private-test-key"}`))}, nil
	})
	err := SendEmailNow("recipient@example.invalid", "Test", "<p>Test</p>")
	if err == nil || !strings.Contains(err.Error(), "Key not found") || strings.Contains(err.Error(), "private-test-key") {
		t.Fatalf("expected useful redacted error, got %v", err)
	}
}

func TestBrevoPaymentIdempotency(t *testing.T) {
	t.Setenv("BREVO_API_KEY", "test-only-key")
	original := http.DefaultTransport
	defer func() { http.DefaultTransport = original }()
	http.DefaultTransport = mailTransport(func(req *http.Request) (*http.Response, error) {
		var payload struct {
			Headers map[string]string `json:"headers"`
			To      []struct {
				Email string `json:"email"`
			} `json:"to"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Headers["idempotencyKey"] != "d622dd77-e5c9-4d3a-85bc-e89052e5a86d" || len(payload.To) != 1 || payload.To[0].Email != "receipt@example.invalid" {
			t.Fatal("missing idempotency or incorrect recipient")
		}
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"code":"duplicate_parameter","message":"Email for this key was already accepted"}`))}, nil
	})
	if err := SendEmailNowIdempotent("receipt@example.invalid", "Payment receipt", "<p>Test</p>", "d622dd77-e5c9-4d3a-85bc-e89052e5a86d"); err != nil {
		t.Fatal("duplicate acceptance should not retry", err)
	}
}
