-- ==========================
-- Coninx Backend Schema for Neon
-- ==========================

SET search_path TO public;

-- --------------------------
-- Admin Users Table
-- --------------------------
CREATE TABLE IF NOT EXISTS admin_users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL
);

-- --------------------------
-- Drivers Table
-- --------------------------
CREATE TABLE IF NOT EXISTS drivers (
    id SERIAL PRIMARY KEY,
    id_number BIGINT UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Vehicles Table
-- --------------------------
CREATE TABLE IF NOT EXISTS vehicles (
    id SERIAL PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    reg_no VARCHAR(50) UNIQUE NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Dispatches Table
-- --------------------------
CREATE TABLE IF NOT EXISTS dispatches (
    id SERIAL PRIMARY KEY,
    recipient VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,           -- Added for Twilio OTP
    location VARCHAR(500) NOT NULL,
    driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL,
    vehicle_id INTEGER REFERENCES vehicles(id) ON DELETE SET NULL,
    invoice INTEGER,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE,       -- Added for OTP confirmation
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------
-- Trips Table
-- --------------------------
CREATE TABLE IF NOT EXISTS trips (
    id SERIAL PRIMARY KEY,
    dispatch_id INT REFERENCES dispatches(id) ON DELETE CASCADE,
    driver_id INT,
    vehicle_id INT,
    status VARCHAR(50) DEFAULT 'started',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    last_updated TIMESTAMP DEFAULT NOW()
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
CREATE INDEX IF NOT EXISTS idx_vehicles_reg_no ON vehicles(reg_no);
CREATE INDEX IF NOT EXISTS idx_deliveries_trip_id ON deliveries(trip_id);
