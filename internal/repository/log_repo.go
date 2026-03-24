package repository

import (
	"database/sql"
	"fmt"

	"github.com/kaantanis/urlshortener/internal/model"
)

type VisitLogRepository struct {
	insertStmt     *sql.Stmt
	recentByCodeSt *sql.Stmt
}

func NewVisitLogRepository(db *sql.DB) (*VisitLogRepository, error) {
	insertStmt, err := db.Prepare(`
		INSERT INTO visit_logs (code, ip, user_agent, referer, accept_lang, origin, host)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare insert visit log: %w", err)
	}

	recentByCodeSt, err := db.Prepare(`
		SELECT id, code, visited_at, ip, user_agent, referer, accept_lang, origin, host
		FROM visit_logs
		WHERE code = ?
		ORDER BY id DESC
		LIMIT ?
	`)
	if err != nil {
		_ = insertStmt.Close()
		return nil, fmt.Errorf("prepare recent visits query: %w", err)
	}

	return &VisitLogRepository{
		insertStmt:     insertStmt,
		recentByCodeSt: recentByCodeSt,
	}, nil
}

func (r *VisitLogRepository) Create(logEntry model.VisitLog) error {
	_, err := r.insertStmt.Exec(
		logEntry.Code,
		logEntry.IP,
		logEntry.UserAgent,
		logEntry.Referer,
		logEntry.AcceptLang,
		logEntry.Origin,
		logEntry.Host,
	)
	if err != nil {
		return fmt.Errorf("insert visit log: %w", err)
	}
	return nil
}

func (r *VisitLogRepository) FindRecentByCode(code string, limit int) ([]model.VisitLog, error) {
	rows, err := r.recentByCodeSt.Query(code, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent visits: %w", err)
	}
	defer rows.Close()

	logs := make([]model.VisitLog, 0, limit)
	for rows.Next() {
		var item model.VisitLog
		if scanErr := rows.Scan(
			&item.ID,
			&item.Code,
			&item.VisitedAt,
			&item.IP,
			&item.UserAgent,
			&item.Referer,
			&item.AcceptLang,
			&item.Origin,
			&item.Host,
		); scanErr != nil {
			return nil, fmt.Errorf("scan visit log row: %w", scanErr)
		}
		logs = append(logs, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visit log rows: %w", err)
	}

	return logs, nil
}

func (r *VisitLogRepository) Close() error {
	if err := r.insertStmt.Close(); err != nil {
		return err
	}
	return r.recentByCodeSt.Close()
}
