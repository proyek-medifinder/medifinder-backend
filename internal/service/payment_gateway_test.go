package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGatewayUsesFixedExpiryAndServerAuthentication(t *testing.T) {
	var server *httptest.Server
	start := time.Date(2026, 11, 2, 10, 0, 0, 0, time.FixedZone("WIB", 7*3600))
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "test-secret" || password != "" {
			t.Error("wrong server authentication")
		}
		var payload struct {
			Details struct {
				ID     string `json:"order_id"`
				Amount int64  `json:"gross_amount"`
			} `json:"transaction_details"`
			Expiry struct {
				Start    string `json:"start_time"`
				Duration int    `json:"duration"`
				Unit     string `json:"unit"`
			} `json:"expiry"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.Details.ID != "order-test" || payload.Details.Amount != 20000 || payload.Expiry.Start != "2026-11-02 10:00:00 +0700" || payload.Expiry.Duration != 15 || payload.Expiry.Unit != "minutes" {
			t.Errorf("invalid payment payload %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(PaymentSession{Token: "token-test", URL: server.URL + "/snap/v2/token-test"})
	}))
	defer server.Close()
	g := &MidtransGateway{Key: "test-secret", SnapURL: server.URL, CoreURL: server.URL, Client: server.Client()}
	session, err := g.Create("order-test", 20000, start)
	if err != nil || session.Token != "token-test" {
		t.Fatal(session, err)
	}
}
func TestGatewayFailureClassification(t *testing.T) {
	for _, code := range []int{401, 404, 500} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) }))
			defer server.Close()
			g := &MidtransGateway{Key: "test", CoreURL: server.URL, Client: server.Client()}
			_, err := g.Status("test")
			if err == nil {
				t.Fatal("failure swallowed")
			}
			if code == 404 && !errors.Is(err, ErrPaymentNotFound) {
				t.Fatal(err)
			}
			if definitiveRejection(err) != (code < 500) {
				t.Fatal("unsafe failure classification", err)
			}
		})
	}
}
func TestPaymentStateNeverRegresses(t *testing.T) {
	for _, c := range []struct{ From, Status, Fraud, Want string }{
		{"pending", "capture", "challenge", "pending"}, {"pending", "capture", "accept", "paid"},
		{"paid", "expire", "", "paid"}, {"cancelled", "settlement", "", "paid"},
		{"refunded", "settlement", "", "refunded"}, {"refunded", "partial_refund", "", "refunded"},
		{"partially_refunded", "settlement", "", "partially_refunded"},
	} {
		if got := paymentState(c.From, c.Status, c.Fraud); got != c.Want {
			t.Errorf("%+v got %s", c, got)
		}
	}
	for _, c := range []struct {
		A, B string
		Want bool
	}{{"20000.00", "20000", true}, {"20000.01", "20000", false}, {"0", "0", false}, {"bad", "20000", false}} {
		if sameAmount(c.A, c.B) != c.Want {
			t.Errorf("amount %+v", c)
		}
	}
}
