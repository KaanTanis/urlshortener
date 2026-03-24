package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kaantanis/urlshortener/internal/model"
	"github.com/mattn/go-sqlite3"
)

var ErrCodeExists = errors.New("code already exists")

type URLRepository struct {
	insertStmt        *sql.Stmt
	findByCodeStmt    *sql.Stmt
	incrementHitsStmt *sql.Stmt
}

func NewURLRepository(db *sql.DB) (*URLRepository, error) {
	insertStmt, err := db.Prepare(`
		INSERT INTO urls (code, original) VALUES (?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare insert url: %w", err)
	}

	findByCodeStmt, err := db.Prepare(`
		SELECT id, code, original, created_at, hit_count
		FROM urls
		WHERE code = ?
	`)
	if err != nil {
		_ = insertStmt.Close()
		return nil, fmt.Errorf("prepare find url by code: %w", err)
	}

	incrementHitsStmt, err := db.Prepare(`
		UPDATE urls
		SET hit_count = hit_count + 1
		WHERE code = ?
	`)
	if err != nil {
		_ = insertStmt.Close()
		_ = findByCodeStmt.Close()
		return nil, fmt.Errorf("prepare increment hits: %w", err)
	}

	return &URLRepository{
		insertStmt:        insertStmt,
		findByCodeStmt:    findByCodeStmt,
		incrementHitsStmt: incrementHitsStmt,
	}, nil
}

func (r *URLRepository) Create(code string, original string) (model.URL, error) {
	result, err := r.insertStmt.Exec(code, original)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return model.URL{}, ErrCodeExists
		}
		return model.URL{}, fmt.Errorf("insert url: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.URL{}, fmt.Errorf("read inserted id: %w", err)
	}

	created, err := r.FindByCode(code)
	if err != nil {
		return model.URL{}, fmt.Errorf("find created url: %w", err)
	}
	created.ID = id

	return created, nil
}

func (r *URLRepository) FindByCode(code string) (model.URL, error) {
	var row model.URL
	err := r.findByCodeStmt.QueryRow(code).Scan(
		&row.ID,
		&row.Code,
		&row.Original,
		&row.CreatedAt,
		&row.HitCount,
	)
	if err != nil {
		return model.URL{}, err
	}
	return row, nil
}

func (r *URLRepository) IncrementHitCount(code string) error {
	_, err := r.incrementHitsStmt.Exec(code)
	if err != nil {
		return fmt.Errorf("increment hit_count: %w", err)
	}
	return nil
}

func (r *URLRepository) Close() error {
	if err := r.insertStmt.Close(); err != nil {
		return err
	}
	if err := r.findByCodeStmt.Close(); err != nil {
		return err
	}
	return r.incrementHitsStmt.Close()
}
