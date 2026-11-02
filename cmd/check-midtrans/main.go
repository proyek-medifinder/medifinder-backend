// Diagnostics are read-only by default. --reconcile explicitly synchronizes
// the specified orders using authenticated Midtrans status and stock updates.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sasaefulanwar/medifinder/internal/config"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	values, fileErr := godotenv.Read(".env")
	inherited := os.Getenv("MIDTRANS_SERVER_KEY")
	_ = godotenv.Load()
	key := os.Getenv("MIDTRANS_SERVER_KEY")
	fmt.Printf("dotenv_readable=%t inherited_key=%t effective_key_present=%t effective_matches_file=%t production=%t\n", fileErr == nil, inherited != "", key != "", key == values["MIDTRANS_SERVER_KEY"], strings.EqualFold(os.Getenv("MIDTRANS_IS_PRODUCTION"), "true"))
	if info, err := os.Stat(".env"); err == nil {
		fmt.Println("dotenv_modified:", info.ModTime().Format(time.RFC3339))
	}
	ids := os.Args[1:]
	reconcile := len(ids) > 0 && ids[0] == "--reconcile"
	if reconcile {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		fmt.Println("Pass an order UUID to inspect")
		return
	}
	g := service.NewMidtransGateway()
	for _, id := range ids {
		req, err := http.NewRequest("GET", g.CoreURL+"/v2/"+id+"/status", nil)
		if err != nil {
			fmt.Println("invalid URL")
			continue
		}
		req.SetBasicAuth(key, "")
		req.Header.Set("Accept", "application/json")
		resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
		if err != nil {
			fmt.Println("Midtrans GET Status network failure")
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 32768))
		resp.Body.Close()
		var result struct {
			Code     string   `json:"status_code"`
			Message  string   `json:"status_message"`
			Status   string   `json:"transaction_status"`
			Order    string   `json:"order_id"`
			Amount   string   `json:"gross_amount"`
			Currency string   `json:"currency"`
			Errors   []string `json:"error_messages"`
		}
		if err = json.Unmarshal(body, &result); err != nil {
			fmt.Printf("Midtrans order=%s HTTP=%d non-JSON body\n", id, resp.StatusCode)
			continue
		}
		safe, _ := json.Marshal(result)
		message := string(safe)
		if key != "" {
			message = strings.ReplaceAll(message, key, "[REDACTED]")
		}
		fmt.Printf("Midtrans HTTP=%d %s\n", resp.StatusCode, message)
	}
	db := config.ConnectDB()
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		fmt.Println("DB read failed")
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SET TRANSACTION READ ONLY`); err != nil {
		fmt.Println("DB read-only setup failed")
		return
	}
	for _, id := range ids {
		var order struct {
			ID          string
			Status      string
			Total       string     `db:"total_harga"`
			Created     time.Time  `db:"created_at"`
			Expires     *time.Time `db:"expires_at"`
			Reservation *string    `db:"reservation_state"`
			HasToken    bool       `db:"has_token"`
			HasURL      bool       `db:"has_url"`
			Details     int        `db:"detail_count"`
		}
		err = tx.GetContext(ctx, &order, `SELECT id,status,total_harga::text,created_at,expires_at,reservation_state,COALESCE(snap_token,'')<>'' AS has_token,COALESCE(payment_url,'')<>'' AS has_url,(SELECT count(*)::int FROM detail_transaksi d WHERE d.transaksi_id=t.id) AS detail_count FROM transaksi t WHERE id=$1`, id)
		if err != nil {
			fmt.Printf("DB order lookup failed: %v\n", err)
			continue
		}
		safe, _ := json.Marshal(order)
		fmt.Println("DB:", string(safe))
	}
	var columns []struct {
		Name string `db:"column_name"`
		Type string `db:"udt_name"`
	}
	if err = tx.SelectContext(ctx, &columns, `SELECT column_name,udt_name FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='transaksi' AND column_name IN ('status','total_harga')`); err == nil {
		data, _ := json.Marshal(columns)
		fmt.Println("DB column types:", string(data))
	}
	var constraints []struct {
		Name       string `db:"conname"`
		Definition string `db:"definition"`
	}
	if err = tx.SelectContext(ctx, &constraints, `SELECT conname,pg_get_constraintdef(oid) AS definition FROM pg_constraint WHERE conrelid='transaksi'::regclass AND contype='c'`); err == nil {
		data, _ := json.Marshal(constraints)
		fmt.Println("DB checks:", string(data))
	}
	if reconcile {
		tx.Rollback()
		for _, id := range ids {
			if err := (&service.PaymentService{DB: db, Gateway: g}).Reconcile(id); err != nil {
				fmt.Printf("Reconcile %s failed: %v\n", id, err)
			} else {
				fmt.Printf("Reconcile %s completed\n", id)
				var state struct {
					Status      string
					Reservation string `db:"reservation_state"`
					Fulfillment string `db:"fulfillment_status"`
				}
				if err := db.Get(&state, `SELECT status,reservation_state,fulfillment_status FROM transaksi WHERE id=$1`, id); err == nil {
					result, _ := json.Marshal(state)
					fmt.Println("DB after reconcile:", string(result))
				}
			}
		}
	} else {
		fmt.Println("Read-only diagnostics finished; no changes made.")
	}
}
