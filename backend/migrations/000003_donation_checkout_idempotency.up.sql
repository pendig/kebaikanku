ALTER TABLE donations ADD COLUMN IF NOT EXISTS idempotency_key varchar(255);
ALTER TABLE donations ADD COLUMN IF NOT EXISTS checkout_token text;
ALTER TABLE donations ADD COLUMN IF NOT EXISTS checkout_redirect_url text;
CREATE UNIQUE INDEX IF NOT EXISTS idx_donations_idempotency_key ON donations (idempotency_key);
