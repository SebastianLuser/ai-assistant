package service

import (
	"database/sql"
	"time"

	"jarvis/pkg/domain"
	"jarvis/pkg/service/sqldata"

	"github.com/lib/pq"
)

type CatalogService interface {
	RecordUsage(name, entryType string, success bool) error
	GetAll() ([]domain.CatalogEntry, error)
	GetByName(name, entryType string) (*domain.CatalogEntry, error)
}

type PGCatalogService struct {
	db *sql.DB
}

func NewPGCatalogService(db *sql.DB) *PGCatalogService {
	return &PGCatalogService{db: db}
}

func (s *PGCatalogService) RecordUsage(name, entryType string, success bool) error {
	var successInc, errorInc int64
	if success {
		successInc = 1
	} else {
		errorInc = 1
	}
	_, err := s.db.Exec(sqldata.UpsertCatalog, name, entryType, successInc, errorInc)
	return err
}

func (s *PGCatalogService) GetAll() ([]domain.CatalogEntry, error) {
	rows, err := s.db.Query(sqldata.CatalogList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.CatalogEntry
	for rows.Next() {
		e, err := scanCatalogEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *PGCatalogService) GetByName(name, entryType string) (*domain.CatalogEntry, error) {
	row := s.db.QueryRow(sqldata.CatalogGet, name, entryType)

	var e domain.CatalogEntry
	var lastUsed sql.NullTime
	var tags []string

	err := row.Scan(&e.Name, &e.Type, &e.UsageCount, &lastUsed, &e.SuccessCount, &e.ErrorCount, pq.Array(&tags), &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastUsed.Valid {
		e.LastUsed = &lastUsed.Time
	}
	e.Tags = tags
	return &e, nil
}

type catalogScanner interface {
	Scan(dest ...any) error
}

func scanCatalogEntry(s catalogScanner) (domain.CatalogEntry, error) {
	var e domain.CatalogEntry
	var lastUsed sql.NullTime
	var tags []string

	err := s.Scan(&e.Name, &e.Type, &e.UsageCount, &lastUsed, &e.SuccessCount, &e.ErrorCount, pq.Array(&tags), &e.CreatedAt)
	if err != nil {
		return e, err
	}

	if lastUsed.Valid {
		t := lastUsed.Time
		e.LastUsed = &t
	}
	e.Tags = tags
	return e, nil
}

// NullCatalogService is a no-op implementation for when no database is available.
type NullCatalogService struct{}

func (NullCatalogService) RecordUsage(string, string, bool) error          { return nil }
func (NullCatalogService) GetAll() ([]domain.CatalogEntry, error)          { return nil, nil }
func (NullCatalogService) GetByName(string, string) (*domain.CatalogEntry, error) { return nil, nil }

var _ CatalogService = (*PGCatalogService)(nil)
var _ CatalogService = NullCatalogService{}

// NewCatalogServiceFromDSN creates a PGCatalogService reusing an existing *sql.DB connection.
func NewCatalogServiceFromDB(db *sql.DB) CatalogService {
	if db == nil {
		return NullCatalogService{}
	}
	return NewPGCatalogService(db)
}

// DB returns the underlying database connection for reuse.
func (s *PGMemoryService) DB() *sql.DB {
	return s.db
}

// LastUsedFormatted returns a human-readable timestamp, or "never" if nil.
func LastUsedFormatted(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.Format("2006-01-02 15:04")
}
