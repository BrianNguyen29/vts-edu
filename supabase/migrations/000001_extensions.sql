-- 000001_extensions.sql
-- Baseline extensions and helpers for PostgreSQL 15+.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Simple updated_at trigger helper.
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
