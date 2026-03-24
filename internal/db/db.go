package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite(dbPath string) (*sql.DB, error) {
	createdNow, err := ensureDBFile(dbPath)
	if err != nil {
		return nil, err
	}

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(0)
	database.SetConnMaxIdleTime(0)

	if _, err = database.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable wal: %w", err)
	}

	if _, err = database.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err = database.Exec(`PRAGMA synchronous=NORMAL;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("set synchronous normal: %w", err)
	}

	if _, err = database.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	if _, err = database.Exec(`PRAGMA temp_store=MEMORY;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("set temp store memory: %w", err)
	}

	if _, err = database.Exec(`PRAGMA wal_autocheckpoint=1000;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("set wal autocheckpoint: %w", err)
	}

	if _, err = database.Exec(`PRAGMA cache_size=-20000;`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("set cache size: %w", err)
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

	slog.Info("sqlite initialized", "db_path", dbPath, "created_now", createdNow)
	return database, nil
}

func ensureDBFile(dbPath string) (bool, error) {
	info, err := os.Stat(dbPath)
	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("db path is a directory: %s", dbPath)
		}
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat db path: %w", err)
	}

	dir := filepath.Dir(dbPath)
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return false, fmt.Errorf("create db directory: %w", mkErr)
	}

	file, createErr := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if createErr != nil {
		if os.IsExist(createErr) {
			return false, nil
		}
		return false, fmt.Errorf("create db file: %w", createErr)
	}

	if closeErr := file.Close(); closeErr != nil {
		return false, fmt.Errorf("close created db file: %w", closeErr)
	}

	return true, nil
}
