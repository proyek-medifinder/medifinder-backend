package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var ErrPaymentNotFound = errors.New("pembayaran belum tercatat di Midtrans")

type PaymentEvent struct {
	RefundAmount string `json:"refund_amount"`
	OrderID      string `json:"order_id"`
	Status       string `json:"transaction_status"`
	Fraud        string `json:"fraud_status"`
	Amount       string `json:"gross_amount"`
	Currency     string `json:"currency"`
}
type PaymentSession struct {
	Token string `json:"token"`
	URL   string `json:"redirect_url"`
}
type PaymentGateway interface {
	Create(string, int64, time.Time) (PaymentSession, error)
	Status(string) (PaymentEvent, error)
	Expire(string) error
}
type gatewayError struct{ Code int }

func (e gatewayError) Error() string {
	if e.Code == 403 {
		return "Akses Midtrans ditolak (HTTP 403). Periksa izin Server Key untuk API yang digunakan; pengecekan status membutuhkan Core API Read."
	}
	return fmt.Sprintf("Midtrans belum dapat memproses pembayaran (HTTP %d)", e.Code)
}

func IsPaymentAccessDenied(err error) bool {
	var gatewayErr gatewayError
	return errors.As(err, &gatewayErr) && (gatewayErr.Code == 401 || gatewayErr.Code == 403)
}
func definitiveRejection(err error) bool {
	if errors.Is(err, ErrPaymentNotFound) {
		return true
	}
	var e gatewayError
	return errors.As(err, &e) && (e.Code == 400 || e.Code == 401 || e.Code == 403 || e.Code == 404 || e.Code == 422)
}

type MidtransGateway struct {
	Key, SnapURL, CoreURL string
	Client                *http.Client
}

func NewMidtransGateway() *MidtransGateway {
	snapURL, coreURL := "https://app.sandbox.midtrans.com", "https://api.sandbox.midtrans.com"
	if strings.EqualFold(os.Getenv("MIDTRANS_IS_PRODUCTION"), "true") {
		snapURL, coreURL = "https://app.midtrans.com", "https://api.midtrans.com"
	}
	return &MidtransGateway{Key: strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY")), SnapURL: snapURL, CoreURL: coreURL, Client: &http.Client{Timeout: 15 * time.Second}}
}
func (g *MidtransGateway) request(method, endpoint string, body, result interface{}) error {
	if g.Key == "" {
		return errors.New("MIDTRANS_SERVER_KEY belum diatur")
	}
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.SetBasicAuth(g.Key, "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := g.Client.Do(req)
	if err != nil {
		return errors.New("Koneksi Midtrans terputus. Periksa status pesanan sebelum mencoba lagi.")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		return ErrPaymentNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return gatewayError{resp.StatusCode}
	}
	var envelope struct {
		Code string `json:"status_code"`
	}
	_ = json.Unmarshal(raw, &envelope)
	if envelope.Code == "404" {
		return ErrPaymentNotFound
	}
	if envelope.Code != "" && envelope.Code != "200" && envelope.Code != "201" && envelope.Code != "202" && envelope.Code != "407" {
		return fmt.Errorf("Midtrans tidak mengonfirmasi permintaan (kode %s)", envelope.Code)
	}
	if result != nil {
		return json.Unmarshal(raw, result)
	}
	return nil
}
func (g *MidtransGateway) Create(id string, amount int64, start time.Time) (PaymentSession, error) {
	var session PaymentSession
	payload := map[string]interface{}{"transaction_details": map[string]interface{}{"order_id": id, "gross_amount": amount},
		"expiry": map[string]interface{}{"start_time": start.In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02 15:04:05 -0700"), "unit": "minutes", "duration": 15}}
	err := g.request("POST", g.SnapURL+"/snap/v1/transactions", payload, &session)
	if err == nil && (session.Token == "" || !validPaymentURL(session.URL, g.SnapURL)) {
		err = errors.New("Respons pembayaran belum lengkap. Periksa kembali status pesanan.")
	}
	return session, err
}
func validPaymentURL(value, base string) bool {
	u, e := url.Parse(value)
	b, _ := url.Parse(base)
	return e == nil && u.Scheme == "https" && u.Host == b.Host && u.User == nil
}
func (g *MidtransGateway) Status(id string) (PaymentEvent, error) {
	var event PaymentEvent
	err := g.request("GET", g.CoreURL+"/v2/"+url.PathEscape(id)+"/status", nil, &event)
	return event, err
}
func (g *MidtransGateway) Expire(id string) error {
	return g.request("POST", g.CoreURL+"/v2/"+url.PathEscape(id)+"/expire", nil, nil)
}
