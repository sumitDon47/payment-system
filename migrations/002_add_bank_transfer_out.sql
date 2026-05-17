-- Database Migration for Bank Transfer Out (Withdrawal) Functionality
-- Run this migration on your PostgreSQL database

-- ============================================================
-- 1. Create bank wallet transfers table (for transfers OUT)
-- ============================================================

CREATE TABLE IF NOT EXISTS bank_wallet_transfers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number    VARCHAR(20) NOT NULL,
    amount          NUMERIC(15, 2) NOT NULL,
    bank_account    VARCHAR(50) NOT NULL,
    account_holder  VARCHAR(100) NOT NULL,
    bank_code       VARCHAR(20) NOT NULL,
    bank_reference  VARCHAR(100) UNIQUE NOT NULL,
    description     VARCHAR(255),
    status          VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed, cancelled
    failure_reason  VARCHAR(255),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    completed_at    TIMESTAMP NULL
);

-- Create indexes for bank wallet transfers
CREATE INDEX IF NOT EXISTS idx_bank_transfer_user ON bank_wallet_transfers(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_transfer_phone ON bank_wallet_transfers(phone_number);
CREATE INDEX IF NOT EXISTS idx_bank_transfer_reference ON bank_wallet_transfers(bank_reference);
CREATE INDEX IF NOT EXISTS idx_bank_transfer_status ON bank_wallet_transfers(status);
CREATE INDEX IF NOT EXISTS idx_bank_transfer_created ON bank_wallet_transfers(created_at);
CREATE INDEX IF NOT EXISTS idx_bank_transfer_account ON bank_wallet_transfers(bank_account);

-- ============================================================
-- 2. Add user bank accounts table
-- ============================================================

CREATE TABLE IF NOT EXISTS user_bank_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_number  VARCHAR(50) NOT NULL,
    account_holder  VARCHAR(100) NOT NULL,
    bank_code       VARCHAR(20) NOT NULL,
    is_verified     BOOLEAN DEFAULT FALSE,
    is_default      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, account_number)
);

-- Create indexes for user bank accounts
CREATE INDEX IF NOT EXISTS idx_user_bank_account ON user_bank_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_bank_verified ON user_bank_accounts(is_verified);
CREATE INDEX IF NOT EXISTS idx_user_bank_default ON user_bank_accounts(is_default);
