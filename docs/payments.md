# Pembayaran MediFinder

## Konfigurasi backend

Salin pengaturan yang diperlukan dari `.env.payments.example` ke environment backend. Jangan masukkan Server Key atau kunci signature ke frontend/Git.

- `MIDTRANS_SERVER_KEY`: Server Key merchant untuk environment yang digunakan.
- `MIDTRANS_IS_PRODUCTION=false`: default tetap Sandbox. Nilai `true` menggunakan endpoint Production dan memerlukan kredensial Production. Perubahan kode ini tidak mengaktifkan Production dengan sendirinya.
- `MIDTRANS_SIGNATURE_KEY`: override kunci yang digunakan untuk memverifikasi signature webhook. Untuk akun/dashboard yang menyediakan Signature Secret Key terpisah, isi dengan kunci penandatangan yang sesuai dengan konfigurasi akun tersebut. Integrasi lama dapat mengosongkannya untuk memakai Server Key. Verifikasi konfigurasi signing akun sebelum aktivasi.
- Notification URL di dashboard: `https://DOMAIN-BACKEND/payment/notify`, dapat diakses publik dengan HTTPS.
- Finish Redirect URL di dashboard: `https://DOMAIN-WEB/profile`. Parameter status pada URL browser tidak pernah dipakai sebagai bukti pembayaran.
- Pada dashboard dengan scoped Server Keys, alur ini membutuhkan Snap Checkout **Write**, Core API **Read** untuk GET Status, dan capability Expire transaction **Write**. Refund diinisiasi melalui dashboard Midtrans, sehingga aplikasi tidak memerlukan izin untuk mengirim refund API.

Webhook memverifikasi signature SHA-512 dan mengambil GET Status yang terautentikasi sebelum mengubah database. Signature lama tidak mencakup transaction_status, sehingga status payload tidak dipercaya begitu saja. Webhook membalas 200 hanya setelah penyimpanan berhasil; kegagalan provider/database mendapat respons non-200 agar dapat dicoba ulang.

## Perilaku

1. Hanya satu checkout pending per pengguna. Klik ulang/tab lain melanjutkan pesanan itu. Setelah selesai atau pembatalan terkonfirmasi, pengguna dapat membuat pesanan berikutnya.
2. Harga disalin dari database dan divalidasi sebagai rupiah bulat. Stok yang tersedia adalah stok fisik dikurangi reservasi.
3. Order dan reservasi disimpan dahulu. HTTP Midtrans dijalankan di luar transaksi database. Keranjang baru dikurangi setelah tautan pembayaran disimpan; kuantitas tambahan yang dibuat saat menunggu tetap dipertahankan.
4. Penolakan pasti dari API pembuatan pembayaran melepaskan reservasi dan mempertahankan keranjang. Timeout/hasil tidak pasti tidak mengirim ulang permintaan pembayaran. Pesanan tetap terlihat di profil untuk diperiksa; jika token tidak sempat tersimpan, tunggu konfirmasi kedaluwarsa sebelum checkout berikutnya.
5. Sesi baru menggunakan blanket expiry 15 menit dengan start_time tetap. Worker memakai jeda 2 menit dan GET Status; pending di-expire lewat API lalu status diperiksa lagi. Gangguan jaringan tidak membatalkan pesanan lokal. Respons 404 setelah fixed expiry berarti sesi tidak memiliki pembayaran tercatat dan sudah tidak dapat dipakai.
6. Migrasi mempertahankan tenggat konservatif 7 hari untuk transaksi lama, karena sesi yang dibuat kode sebelumnya tidak memiliki expiry eksplisit. Stok transaksi lama tidak dilepas sepihak setelah 15 menit.
7. Perubahan status dan stok mengunci baris order serta obat, sehingga notifikasi bersamaan tidak mengurangi/melepas stok berulang. Pembayaran terlambat tetap dicatat lunas. Jika stok tidak cukup, fulfillment menjadi `needs_review`; admin dapat memperbarui stok fisik lalu menekan Periksa ulang stok, atau melakukan pengembalian dana sesuai kasus sebelum pesanan bisa dipenuhi.
8. Pembayaran dan fulfillment terpisah: `paid → preparing → ready → completed`. Hanya admin apotek pemilik dapat memproses pesanan lunas dan stoknya sudah tercatat consumed.
9. Refund penuh/sebagian mengikuti GET Status; `refunded_amount` menyimpan jumlah kumulatif, bukan penambahan tiap webhook. Refund tidak otomatis menambah stok fisik. Pengembalian barang harus diverifikasi terpisah.
10. Detail, refresh, dan tautan pembayaran dibatasi ke pemilik pesanan; admin hanya dapat mengakses transaksi apoteknya. Dashboard menampilkan pendapatan bersih untuk halaman yang sedang ditampilkan, tidak memasukkan pending/cancelled.

## Migrasi dan data lama

Saat backend baru dijalankan, `config.MigratePayments` melakukan migrasi kolom secara additive dan dapat dijalankan ulang. Backend memerlukan hak ALTER TABLE. Tidak ada akun/pesanan yang dihapus atau transaksi uang yang dibuat oleh migrasi.

Migrasi tidak menebak pembayaran historis yang pernah diabaikan kode lama, dan tidak mengubah stok fisik. Untuk memeriksa reservasi lama yang mungkin tertinggal, jalankan query baca berikut dan cocokkan dengan dashboard Midtrans sebelum mengoreksi data:

```sql
SELECT o.id, o.reserved_stock,
       COALESCE(SUM(d.jumlah) FILTER (WHERE t.reservation_state='reserved'),0) AS expected_reserved
FROM obat o
LEFT JOIN detail_transaksi d ON d.obat_id=o.id
LEFT JOIN transaksi t ON t.id=d.transaksi_id
GROUP BY o.id,o.reserved_stock
HAVING o.reserved_stock <> COALESCE(SUM(d.jumlah) FILTER (WHERE t.reservation_state='reserved'),0);
```

Semua pembayaran masih memakai satu merchant yang dikonfigurasi di backend. Pencatatan transaksi per apotek bukan pembagian dana otomatis. Settlement merchant, biaya layanan, dan pencairan ke apotek perlu mengikuti pengaturan bisnis/Midtrans tersendiri.

## Pengujian

Tes tidak memanggil Midtrans/Brevo asli dan tidak membaca `.env` aplikasi. Gunakan PostgreSQL disposable dengan URL eksplisit:

```powershell
$env:TEST_DATABASE_URL='postgres://USER@127.0.0.1:PORT/postgres?sslmode=disable'
go test ./... -count=1 -timeout=120s
```

Tes pembayaran membuat schema acak sendiri dan membersihkannya setelah selesai. Cakupan: checkout bersamaan, perubahan keranjang selama request, penolakan/timeout API, notifikasi paralel, nominal/mata uang tidak cocok, refund kumulatif, rekonsiliasi expiry, settlement terlambat, akses lintas akun dan urutan fulfillment. Tanpa TEST_DATABASE_URL, tes PostgreSQL dilewati; tes unit/signature/HTTP gateway tetap berjalan.

Rujukan Midtrans: https://docs.midtrans.com/reference/expire-a-snap-session, https://docs.midtrans.com/reference/get-transaction-status, https://docs.midtrans.com/docs/https-notification-webhooks.


## Melanjutkan pembayaran saat GET Status ditolak

Pembukaan ulang tautan Snap yang sudah tersimpan tidak memanggil GET Status. Backend tetap memeriksa pemilik transaksi, status lokal pending, tenggat, dan ketersediaan apotek. Tautan yang sudah dibayar/kedaluwarsa juga tetap ditangani oleh halaman Midtrans. Pembukaan tautan tidak pernah menandai transaksi lunas.

HTTP 403 saat Periksa status/webhook berarti akses API ditolak. Periksa izin key di Midtrans: Snap Checkout Write untuk membuat sesi; Core API Read untuk GET Status; Expire Write untuk menutup pembayaran pending. Setelah mengganti Server Key di environment backend, restart backend. Pengecekan status dan verifikasi webhook tetap memerlukan akses GET Status yang berfungsi.

## Diagnosis pesanan tertahan

`go run ./cmd/check-midtrans ORDER_UUID` memeriksa GET Status asli dan membaca metadata transaksi tanpa menulis database. Kredensial dan token pembayaran tidak dicetak. Ini membantu membedakan penolakan key dengan kegagalan query database.

`go run ./cmd/check-midtrans --reconcile ORDER_UUID` secara eksplisit menyinkronkan pesanan tersebut lewat PaymentService; perintah ini dapat memperbarui status/stok dan meminta expiry untuk pending yang sudah melewati tenggat. Gunakan hanya untuk pesanan yang memang sedang diperbaiki.

Checkout yang menemukan pending lama melakukan rekonsiliasi terlebih dahulu. Pesanan lama yang telah settlement atau dibatalkan tidak dipakai sebagai sesi checkout baru. Tombol Lanjutkan pembayaran tetap bisa membuka sesi pending yang tersimpan tanpa bergantung pada izin GET Status.

Waktu mulai expiry Snap dihitung dari expires_at (timestamptz) dikurangi 15 menit. created_at pada database lama dapat berupa timestamp tanpa zona waktu dan tidak dipakai sebagai waktu mulai Snap.

## Email konfirmasi pembayaran

Pembayaran yang pertama kali berpindah ke paid membuat satu entri payment_email_outbox dalam transaksi database yang sama. Email ditujukan ke email pengguna pemilik pesanan; detailnya mencakup nomor pesanan, waktu verifikasi (WIB), nama/alamat apotek, nama obat, jumlah, harga satuan, subtotal, dan total. Nama obat dan harga disalin saat checkout, sehingga perubahan produk berikutnya tidak mengubah rincian pembelian. Transaksi lama dibantu backfill nama obat saat migrasi.

Worker memeriksa antrean setiap 10 detik, mengunci pengambilan pesan dengan lease, dan mengirim melalui konfigurasi Brevo yang sama dengan email akun. Kegagalan pengiriman disimpan di last_error dan dicoba ulang dengan jeda bertambah hingga satu jam. Pembayaran tetap lunas meskipun Brevo gagal. Pengiriman yang telah diterima Brevo ditandai sent_at; ini berarti diterima provider, bukan jaminan sudah tiba di inbox.

Primary key order_id mencegah notifikasi pembayaran berulang membuat antrean ganda. Retry memakai idempotencyKey UUID yang sama pada Brevo. Proteksi idempotency Brevo memiliki TTL terbatas; pengiriman tepat satu kali ke inbox tidak dapat dijamin mutlak jika proses mati setelah provider menerima pesan tetapi sebelum database mencatatnya, lalu baru pulih setelah TTL berakhir.

Template payment_success.html memakai layout email bersama. Untuk pengecualian stok, email menyebut pembayaran diterima dan stok masih diperiksa, tanpa menjanjikan obat sudah siap diambil. Pending, capture challenge, pembayaran ditolak, dan refund tidak memicu email konfirmasi lunas. Tidak ada pengiriman massal untuk transaksi yang sudah paid sebelum fitur ini dipasang.

Migrasi antrean dan kolom nama obat berjalan otomatis saat backend dijalankan ulang. Pastikan BREVO_API_KEY dan EMAIL_SENDER tetap valid. Pengujian memakai pengirim tiruan; tidak ada email sungguhan yang dikirim saat tes.
