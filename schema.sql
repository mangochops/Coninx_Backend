-- ==========================
-- Coninx Backend Schema for Neon
-- ==========================

-- Use public schema
SET search_path TO public;

-- --------------------------
-- Admin Users Table
-- --------------------------
CREATE TABLE IF NOT EXISTS admin_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Drivers Table
-- --------------------------
CREATE TABLE IF NOT EXISTS drivers (
    id SERIAL PRIMARY KEY,
    phone_number BIGINT UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Vehicles Table
-- --------------------------
CREATE TABLE IF NOT EXISTS vehicles (
    id SERIAL PRIMARY KEY,
    make VARCHAR(100),
    model VARCHAR(100),
    year INTEGER,
    license_plate VARCHAR(50) UNIQUE,
    driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Dispatches Table
-- --------------------------
CREATE TABLE IF NOT EXISTS dispatches (
    id SERIAL PRIMARY KEY,
    recipient VARCHAR(255) NOT NULL,
    location VARCHAR(500) NOT NULL,
    driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL,
    vehicle_id INTEGER REFERENCES vehicles(id) ON DELETE SET NULL,
    invoice INTEGER,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Trips Table
-- --------------------------
CREATE TABLE IF NOT EXISTS trips (
    id SERIAL PRIMARY KEY,
    destination VARCHAR(500) NOT NULL,
    driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL,
    recipient_name VARCHAR(255) NOT NULL,
    invoice INTEGER,
    status VARCHAR(50) DEFAULT 'requested',
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    dispatch_id INTEGER REFERENCES dispatches(id) ON DELETE SET NULL,
    vehicle_id INTEGER REFERENCES vehicles(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Deliveries Table
-- --------------------------
CREATE TABLE IF NOT EXISTS deliveries (
    id SERIAL PRIMARY KEY,
    recipient VARCHAR(255) NOT NULL,
    condition VARCHAR(255),
    delivery_note TEXT,
    trip_id INTEGER REFERENCES trips(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Indexes for performance
-- --------------------------
CREATE INDEX IF NOT EXISTS idx_trips_driver_id ON trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX IF NOT EXISTS idx_dispatches_driver_id ON dispatches(driver_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_driver_id ON vehicles(driver_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_trip_id ON deliveries(trip_id);

-- --------------------------
-- Optional: Test insert
-- --------------------------
-- INSERT INTO admin_users (username, email, password) VALUES ('Admin Test', 'admin@test.com', '123456');

-- INSERT INTO drivers (phone_number, password, name) VALUES (1234567890, 'driverpass', 'John Doe');