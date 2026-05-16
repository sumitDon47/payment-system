-- Database Migration for Phone-Based Login and Bank API Integration
-- Run this migration on your PostgreSQL database

-- ============================================================
-- 1. Add phone number fields to users table
-- ============================================================

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20) UNIQUE;

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN DEFAULT FALSE;

-- Create index for faster phone lookups
CREATE INDEX IF NOT EXISTS idx_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_phone_verified ON users(phone_verified);

-- ============================================================
-- 2. Create phone OTP table
-- ============================================================

CREATE TABLE IF NOT EXISTS phone_otps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number    VARCHAR(20) NOT NULL,
    code            VARCHAR(6) NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    attempts        INTEGER DEFAULT 0,
    verified        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW(),
    UNIQUE(phone_number, code)
);

CREATE INDEX IF NOT EXISTS idx_phone_otp_phone ON phone_otps(phone_number);
CREATE INDEX IF NOT EXISTS idx_phone_otp_expires ON phone_otps(expires_at);

-- ============================================================
-- 3. Create bank wallet loads table
-- ============================================================

CREATE TABLE IF NOT EXISTS bank_wallet_loads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number    VARCHAR(20) NOT NULL,
    amount          NUMERIC(15, 2) NOT NULL,
    bank_reference  VARCHAR(100) UNIQUE NOT NULL,
    bank_code       VARCHAR(20) NOT NULL,
    status          VARCHAR(20) DEFAULT 'pending',
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    completed_at    TIMESTAMP NULL
);

-- Create indexes for bank wallet loads
CREATE INDEX IF NOT EXISTS idx_bank_wallet_user ON bank_wallet_loads(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_wallet_phone ON bank_wallet_loads(phone_number);
CREATE INDEX IF NOT EXISTS idx_bank_wallet_reference ON bank_wallet_loads(bank_reference);
CREATE INDEX IF NOT EXISTS idx_bank_wallet_status ON bank_wallet_loads(status);
CREATE INDEX IF NOT EXISTS idx_bank_wallet_created ON bank_wallet_loads(created_at);

-- ============================================================
-- 4. Add audit columns to transactions table (optional but recommended)
-- ============================================================

ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'NPR';

ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS description VARCHAR(255);

-- Create index for faster transaction lookups
CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender_id);
CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created ON transactions(created_at);

-- ============================================================
-- 5. Create phone-based transaction audit log (optional)
-- ============================================================

CREATE TABLE IF NOT EXISTS phone_transfer_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_phone    VARCHAR(20) NOT NULL,
    receiver_phone  VARCHAR(20) NOT NULL,
    sender_id       UUID REFERENCES users(id),
    receiver_id     UUID REFERENCES users(id),
    transaction_id  UUID REFERENCES transactions(id),
    amount          NUMERIC(15, 2) NOT NULL,
    status          VARCHAR(20),
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_phone_transfer_sender ON phone_transfer_logs(sender_phone);
CREATE INDEX IF NOT EXISTS idx_phone_transfer_receiver ON phone_transfer_logs(receiver_phone);
CREATE INDEX IF NOT EXISTS idx_phone_transfer_created ON phone_transfer_logs(created_at);

-- ============================================================
-- 6. Create bank API audit table (optional but recommended for compliance)
-- ============================================================

CREATE TABLE IF NOT EXISTS bank_api_audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_code       VARCHAR(20) NOT NULL,
    request_type    VARCHAR(50) NOT NULL,
    bank_reference  VARCHAR(100),
    amount          NUMERIC(15, 2),
    phone_number    VARCHAR(20),
    status          VARCHAR(20),
    request_at      TIMESTAMP NOT NULL,
    response_status INTEGER,
    error_message   VARCHAR(500),
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_audit_bank ON bank_api_audit_logs(bank_code);
CREATE INDEX IF NOT EXISTS idx_bank_audit_reference ON bank_api_audit_logs(bank_reference);
CREATE INDEX IF NOT EXISTS idx_bank_audit_created ON bank_api_audit_logs(created_at);

-- ============================================================
-- 7. Verify required extensions
-- ============================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- Sample queries to verify the migration
-- ============================================================

-- Check if phone_number column exists in users
-- SELECT column_name FROM information_schema.columns WHERE table_name='users' AND column_name='phone_number';

-- List all bank wallet loads
-- SELECT id, phone_number, amount, status, created_at FROM bank_wallet_loads;

-- Check pending wallet loads
-- SELECT id, phone_number, amount, created_at FROM bank_wallet_loads WHERE status = 'pending' AND created_at > NOW() - INTERVAL '24 hours';
