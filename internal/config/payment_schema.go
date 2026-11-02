package config

import "github.com/jmoiron/sqlx"

// Additive, repeatable migration. Existing pending Snap sessions may remain valid
// for seven days; never shorten their deadline using the old local 15 minute timer.
func MigratePayments(db *sqlx.DB) error {
	_, err := db.Exec(`
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS expires_at timestamptz;
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS reservation_state text;
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS fulfillment_status text NOT NULL DEFAULT 'awaiting_payment';
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS checkout_cart_id uuid;
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS cart_cleared boolean NOT NULL DEFAULT true;
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS payment_attempted_at timestamptz;
 ALTER TABLE transaksi ADD COLUMN IF NOT EXISTS refunded_amount numeric NOT NULL DEFAULT 0;
 ALTER TABLE detail_transaksi ADD COLUMN IF NOT EXISTS nama_obat text;
 UPDATE detail_transaksi d SET nama_obat=o.nama FROM obat o WHERE d.obat_id=o.id AND d.nama_obat IS NULL;
 CREATE TABLE IF NOT EXISTS payment_email_outbox (
   order_id uuid PRIMARY KEY REFERENCES transaksi(id) ON DELETE CASCADE,
   recipient text NOT NULL, payload jsonb NOT NULL,
   attempts integer NOT NULL DEFAULT 0, next_attempt_at timestamptz NOT NULL DEFAULT NOW(),
   locked_until timestamptz, sent_at timestamptz, last_error text,
   created_at timestamptz NOT NULL DEFAULT NOW()
 );
 CREATE INDEX IF NOT EXISTS payment_email_pending_idx ON payment_email_outbox(next_attempt_at) WHERE sent_at IS NULL;
 UPDATE transaksi SET reservation_state=CASE WHEN status='pending' THEN 'reserved'
   WHEN status IN ('paid','refunded','partially_refunded') THEN 'consumed' ELSE 'released' END
   WHERE reservation_state IS NULL;
 UPDATE transaksi SET expires_at=created_at+INTERVAL '7 days' WHERE expires_at IS NULL;
 ALTER TABLE transaksi ALTER COLUMN reservation_state SET DEFAULT 'reserved';
 ALTER TABLE transaksi ALTER COLUMN expires_at SET DEFAULT (NOW()+INTERVAL '7 days');
 UPDATE transaksi SET fulfillment_status='paid' WHERE status='paid' AND fulfillment_status='awaiting_payment';
 UPDATE transaksi SET refunded_amount=total_harga WHERE status='refunded' AND refunded_amount=0;
 CREATE INDEX IF NOT EXISTS transaksi_pending_user_idx ON transaksi(user_id) WHERE status='pending';
 `)
	return err
}
