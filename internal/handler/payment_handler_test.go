package handler

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type unavailableGateway struct{}

func (unavailableGateway) Create(string, int64, time.Time) (service.PaymentSession, error) {
	return service.PaymentSession{}, errors.New("offline")
}
func (unavailableGateway) Status(string) (service.PaymentEvent, error) {
	return service.PaymentEvent{}, errors.New("offline")
}
func (unavailableGateway) Expire(string) error { return errors.New("offline") }
func TestNotificationRejectsInvalidAndRetriesProviderFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("MIDTRANS_SIGNATURE_KEY", "signature-test")
	h := &PaymentHandler{Service: &service.PaymentService{Gateway: unavailableGateway{}}}
	r := gin.New()
	r.POST("/payment/notify", h.Notification)
	id := "3867a555-9a9e-498f-8a4b-bf8ca3f971f2"
	hash := sha512.Sum512([]byte(id + "200" + "20000.00" + "signature-test"))
	signature := hex.EncodeToString(hash[:])
	body := `{"order_id":"` + id + `","status_code":"200","gross_amount":"20000.00","transaction_status":"settlement","signature_key":"` + signature + `"}`
	for _, c := range []struct {
		Body string
		Code int
	}{{"invalid", 400}, {strings.Replace(body, signature, "bad", 1), 401}, {body, 503}} {
		req := httptest.NewRequest("POST", "/payment/notify", strings.NewReader(c.Body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != c.Code {
			t.Fatalf("got %d want %d: %s", w.Code, c.Code, w.Body.String())
		}
	}
	req := MidtransNotification{OrderID: id, StatusCode: "200", GrossAmount: "20000.00", SignatureKey: signature}
	if verifyMidtransSignature(req, "") || verifyMidtransSignature(req, "wrong") || !verifyMidtransSignature(req, "signature-test") {
		t.Fatal("signature verification broken")
	}
}
