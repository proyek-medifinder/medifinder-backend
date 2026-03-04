-- Pake UUID biar aman buat backend Go lu
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Tabel ROLES (Superadmin, Admin Apotek, User)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_name VARCHAR(50) NOT NULL UNIQUE
);

-- 2. Tabel USERS
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    google_id VARCHAR(255),
    profile_picture TEXT,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    status VARCHAR(50) DEFAULT 'Active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Ini jawaban buat BR-08 (Soft Delete)
    deleted_at TIMESTAMP NULL 
);

-- 3. Tabel APOTEK
CREATE TABLE apotek (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- UNIQUE di admin_id ini kunci buat BR-01 & BR-02 (1 Admin = 1 Apotek)
    admin_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE, 
    nama VARCHAR(100) NOT NULL,
    deskripsi TEXT,
    alamat TEXT NOT NULL,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    phone_number VARCHAR(20),
    photo_url TEXT,
    open_time TIME,
    close_time TIME,
    status VARCHAR(50) DEFAULT 'Pending',
    rating FLOAT DEFAULT 0,
    total_reviews INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Tabel OBAT
CREATE TABLE obat (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    apotek_id UUID NOT NULL REFERENCES apotek(id) ON DELETE CASCADE,
    nama VARCHAR(150) NOT NULL,
    stok INT NOT NULL DEFAULT 0,
    harga DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Tabel CART (Keranjang)
CREATE TABLE cart (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- UNIQUE di user_id ini kunci buat BR-03 (1 User = 1 Keranjang)
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE, 
    apotek_id UUID NOT NULL REFERENCES apotek(id) ON DELETE CASCADE
);

-- 6. Tabel CART_ITEM (Isi Keranjang)
CREATE TABLE cart_item (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_id UUID NOT NULL REFERENCES cart(id) ON DELETE CASCADE,
    obat_id UUID NOT NULL REFERENCES obat(id) ON DELETE CASCADE,
    jumlah INT NOT NULL DEFAULT 1
);

-- 7. Tabel TRANSAKSI
CREATE TABLE transaksi (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    apotek_id UUID NOT NULL REFERENCES apotek(id) ON DELETE RESTRICT,
    total DECIMAL(12, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'Pending',
    -- Ini tambahan krusial buat payment gateway (Midtrans/Xendit)
    snap_token VARCHAR(255), 
    payment_url TEXT,        
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. Tabel DETAIL_TRANSAKSI
CREATE TABLE detail_transaksi (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaksi_id UUID NOT NULL REFERENCES transaksi(id) ON DELETE CASCADE,
    obat_id UUID NOT NULL REFERENCES obat(id) ON DELETE RESTRICT,
    jumlah INT NOT NULL,
    harga DECIMAL(12, 2) NOT NULL
);

-- 9. Tabel RESEP (Sesuai SRS Bab V, nempel ke Transaksi)
CREATE TABLE resep (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- UNIQUE karena 1 transaksi biasanya cuma butuh 1 upload resep
    transaksi_id UUID NOT NULL UNIQUE REFERENCES transaksi(id) ON DELETE CASCADE, 
    file_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. Tabel VERIFICATION_LOGS
CREATE TABLE verification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    superadmin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);