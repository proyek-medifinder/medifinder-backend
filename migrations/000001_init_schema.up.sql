CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ROLES
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL
);

-- USERS
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role_id INT REFERENCES roles(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- APOTEK
CREATE TABLE apotek (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID UNIQUE REFERENCES users(id),
    nama VARCHAR(150) NOT NULL,
    alamat TEXT NOT NULL,
    latitude DECIMAL(9,6) NOT NULL,
    longitude DECIMAL(9,6) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- OBAT
CREATE TABLE obat (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    apotek_id UUID REFERENCES apotek(id) ON DELETE CASCADE,
    nama VARCHAR(150) NOT NULL,
    stok INT NOT NULL CHECK (stok >= 0),
    harga BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- CART
CREATE TABLE cart (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE REFERENCES users(id),
    apotek_id UUID REFERENCES apotek(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- CART ITEMS
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_id UUID REFERENCES cart(id) ON DELETE CASCADE,
    obat_id UUID REFERENCES obat(id),
    jumlah INT NOT NULL CHECK (jumlah > 0)
);

-- TRANSAKSI
CREATE TYPE transaksi_status AS ENUM (
    'pending',
    'paid',
    'cancelled',
    'confirmed'
);

CREATE TABLE transaksi (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    apotek_id UUID REFERENCES apotek(id),
    total BIGINT NOT NULL,
    status transaksi_status DEFAULT 'pending',
    expired_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- DETAIL TRANSAKSI
CREATE TABLE detail_transaksi (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaksi_id UUID REFERENCES transaksi(id) ON DELETE CASCADE,
    obat_id UUID REFERENCES obat(id),
    jumlah INT NOT NULL,
    harga BIGINT NOT NULL
);

-- RESEP
CREATE TABLE resep (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaksi_id UUID UNIQUE REFERENCES transaksi(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL
);

-- SEED ROLES
INSERT INTO roles (role_name) VALUES
('user'),
('admin_apotek'),
('super_admin');
