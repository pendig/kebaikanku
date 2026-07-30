ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS subcategory varchar(100);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS campaign_type varchar(50) NOT NULL DEFAULT 'target_deadline';
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS banner_url text;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS location varchar(255);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS beneficiary_note text;
