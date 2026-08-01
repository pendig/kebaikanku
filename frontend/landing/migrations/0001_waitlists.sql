CREATE TABLE IF NOT EXISTS waitlists (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	source TEXT,
	ip_address TEXT,
	user_agent TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_waitlists_ip_created_at ON waitlists (ip_address, created_at);
