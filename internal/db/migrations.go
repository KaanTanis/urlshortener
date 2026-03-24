package db

const createURLsTable = `
CREATE TABLE IF NOT EXISTS urls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	code TEXT UNIQUE NOT NULL,
	original TEXT NOT NULL,
	created_at TEXT DEFAULT (datetime('now')),
	hit_count INTEGER DEFAULT 0
);
`

const createVisitLogsTable = `
CREATE TABLE IF NOT EXISTS visit_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	code TEXT NOT NULL,
	visited_at TEXT DEFAULT (datetime('now')),
	ip TEXT,
	user_agent TEXT,
	referer TEXT,
	accept_lang TEXT,
	origin TEXT,
	host TEXT
);
`

const createVisitLogsCodeIDIndex = `
CREATE INDEX IF NOT EXISTS idx_visit_logs_code_id
ON visit_logs(code, id DESC);
`
