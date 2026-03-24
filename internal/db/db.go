package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite(dbPath string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err = database.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable wal: %w", err)
	}

	if _, err = database.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err = database.Exec(createURLsTable); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("create urls table: %w", err)
	}

	if _, err = database.Exec(createVisitLogsTable); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("create visit_logs table: %w", err)
	}

	if _, err = database.Exec(createVisitLogsCodeIDIndex); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("create visit_logs index: %w", err)
	}

	slog.Info("sqlite initialized", "db_path", dbPath)
	return database, nil
}
