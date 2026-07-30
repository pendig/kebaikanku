CREATE TABLE IF NOT EXISTS payment_settings (
    id varchar(32) PRIMARY KEY,
    mode varchar(16) NOT NULL,
    server_key_cipher text,
    client_key_cipher text,
    created_at timestamptz,
    updated_at timestamptz
);
