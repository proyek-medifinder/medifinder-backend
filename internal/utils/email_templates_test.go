package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmailTemplates(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	fixtures := []struct {
		file string
		data interface{}
	}{
		{"reset_password.html", struct{ Name, ResetLink string }{"Noval", "https://medifinder.example/reset-password?token=" + strings.Repeat("a", 64)}},
		{"admin_registration.html", struct{ Name, NamaApotek string }{"Noval", "Apotek Sehat Bersama"}},
		{"admin_approved.html", struct{ Name, NamaApotek, LoginLink string }{"Noval", "Apotek Sehat Bersama", "https://medifinder.example/admin/login"}},
		{"admin_rejected.html", struct{ Name, NamaApotek, Reason string }{"Noval", "Apotek Sehat Bersama", "Alamat apotek belum lengkap.\nMohon lengkapi nomor dan nama jalan."}},
		{"password_changed.html", struct{ Email string }{"noval@example.invalid"}},
		{"admin_deleted.html", struct{ Name string }{"Admin Apotek Uji"}},
		{"admin_status_changed.html", struct {
			Name      string
			Enabled   bool
			LoginLink string
		}{"Noval", true, "https://medifinder.example/admin/login"}},
		{"admin_status_changed.html", struct {
			Name      string
			Enabled   bool
			LoginLink string
		}{"Noval", false, "https://medifinder.example/admin/login"}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.file, func(t *testing.T) {
			html, err := ParseTemplate("templates/emails/"+fixture.file, fixture.data)
			if err != nil {
				t.Fatal(err)
			}
			for _, required := range []string{"<!doctype html>", "name=\"viewport\"", "MediFinder", "Tim MediFinder", "max-width:600px", "<h1", "</html>"} {
				if !strings.Contains(html, required) {
					t.Errorf("missing shared layout element %q", required)
				}
			}
			if strings.Contains(html, "{{") || strings.Contains(html, "ZgotmplZ") {
				t.Fatal("unrendered or invalid template data")
			}
			if fixture.file == "reset_password.html" && !strings.Contains(html, "15 menit") {
				t.Fatal("reset expiry is missing")
			}
			if out := os.Getenv("EMAIL_PREVIEW_DIR"); out != "" {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(out, fixture.file), []byte(html), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	html, err := ParseTemplate("templates/emails/admin_rejected.html", struct{ Name, NamaApotek, Reason string }{
		"<script>alert(1)</script>", "Apotek <b>Uji</b>", "<img src=x onerror=alert(1)>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "<img src=x") || strings.Contains(html, "Apotek <b>") {
		t.Fatal("user-controlled email fields must be escaped")
	}
}

func TestPaymentReceiptTemplate(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	data := map[string]interface{}{
		"CustomerName": "<script>Pelanggan</script>", "OrderID": "pesanan-uji", "PharmacyName": "Apotek <b>Uji</b>", "PharmacyAddress": "Jl. Sehat 10",
		"Total": "Rp 20.000", "ConfirmedAt": "02/11/2026 12:00 WIB", "NeedsReview": false,
		"Items": []map[string]interface{}{{"Name": "Paracetamol 500 mg", "Quantity": 2, "UnitPrice": "Rp 10.000", "Subtotal": "Rp 20.000"}},
	}
	for _, review := range []bool{false, true} {
		data["NeedsReview"] = review
		body, err := ParseTemplate("templates/emails/payment_success.html", data)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"<!doctype html>", "max-width:600px", "Paracetamol 500 mg", "Rp 10.000", "Rp 20.000", "Lunas", "pesanan-uji", "Jl. Sehat 10"} {
			if !strings.Contains(body, want) {
				t.Errorf("receipt missing %s", want)
			}
		}
		if strings.Contains(body, "<script>") || strings.Contains(body, "<b>Uji</b>") || strings.Contains(body, "{{") || strings.Contains(body, "ZgotmplZ") {
			t.Fatal("unsafe or invalid receipt template")
		}
		if strings.Contains(body, "Ketersediaan stok pesanan sedang diperiksa") != review {
			t.Fatal("incorrect stock review notice")
		}
	}
}
