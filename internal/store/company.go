package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type CompanyStore struct {
	db *sql.DB
}

func NewCompanyStore(db *sql.DB) *CompanyStore {
	return &CompanyStore{db: db}
}

func (s *CompanyStore) Create(ctx context.Context, c *domain.Company) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO companies (name) VALUES ($1) RETURNING id, created_at, updated_at`,
		c.Name,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating company: %w", err)
	}
	return nil
}

func (s *CompanyStore) GetByID(ctx context.Context, id string) (*domain.Company, error) {
	c := &domain.Company{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at, deleted_at
		 FROM companies WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "company", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting company: %w", err)
	}
	return c, nil
}
