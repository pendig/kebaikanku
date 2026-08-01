CREATE TABLE IF NOT EXISTS schema_migrations (
    version varchar(191) PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organizations (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL,
    email varchar(191) NOT NULL,
    password_hash varchar(255) NOT NULL,
    address text,
    status varchar(50) NOT NULL DEFAULT 'pending',
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_email ON organizations (email);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations (deleted_at);

CREATE TABLE IF NOT EXISTS donors (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL,
    phone_number varchar(50) NOT NULL,
    email varchar(191),
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_donors_phone_number ON donors (phone_number);
CREATE INDEX IF NOT EXISTS idx_donors_deleted_at ON donors (deleted_at);

CREATE TABLE IF NOT EXISTS waitlists (
    id char(36) PRIMARY KEY,
    email varchar(191) NOT NULL,
    source varchar(255),
    ip_address varchar(64),
    user_agent text,
    created_at timestamptz,
    updated_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_waitlists_email ON waitlists (email);

CREATE TABLE IF NOT EXISTS campaigns (
    id char(36) PRIMARY KEY,
    organization_id char(36) NOT NULL REFERENCES organizations (id),
    title varchar(255) NOT NULL,
    slug varchar(255) NOT NULL,
    description text,
    category varchar(100) NOT NULL,
    subcategory varchar(100),
    campaign_type varchar(50) NOT NULL DEFAULT 'target_deadline',
    banner_url text,
    location varchar(255),
    beneficiary_note text,
    target_amount decimal(15,2) NOT NULL,
    collected_amount decimal(15,2) NOT NULL DEFAULT 0,
    end_date timestamptz,
    status varchar(50) NOT NULL DEFAULT 'active',
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_campaigns_organization_id ON campaigns (organization_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_campaigns_slug ON campaigns (slug);
CREATE INDEX IF NOT EXISTS idx_campaigns_deleted_at ON campaigns (deleted_at);

CREATE TABLE IF NOT EXISTS donations (
    id char(36) PRIMARY KEY,
    campaign_id char(36) NOT NULL REFERENCES campaigns (id),
    donor_id char(36) NOT NULL REFERENCES donors (id),
    amount decimal(15,2) NOT NULL,
    platform_tip decimal(15,2) DEFAULT 0,
    pg_fee decimal(15,2) DEFAULT 0,
    status varchar(50) NOT NULL DEFAULT 'pending',
    payment_method varchar(100),
    provider varchar(100),
    provider_order_id varchar(191),
    provider_transaction_id varchar(191),
    provider_status varchar(100),
    provider_payload text,
    paid_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_donations_campaign_id ON donations (campaign_id);
CREATE INDEX IF NOT EXISTS idx_donations_donor_id ON donations (donor_id);
CREATE INDEX IF NOT EXISTS idx_donations_provider_order_id ON donations (provider_order_id);
CREATE INDEX IF NOT EXISTS idx_donations_provider_transaction_id ON donations (provider_transaction_id);
