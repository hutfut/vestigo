package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/lib/pq"
)

type StakeholderStore struct {
	db *sql.DB
}

func NewStakeholderStore(db *sql.DB) *StakeholderStore {
	return &StakeholderStore{db: db}
}

func (s *StakeholderStore) Create(ctx context.Context, sh *domain.Stakeholder) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO stakeholders (company_id, name, email, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		sh.CompanyID, sh.Name, sh.Email, sh.Role,
	).Scan(&sh.ID, &sh.CreatedAt, &sh.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating stakeholder: %w", err)
	}
	return nil
}

func (s *StakeholderStore) GetByID(ctx context.Context, id string) (*domain.Stakeholder, error) {
	sh := &domain.Stakeholder{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, name, email, role, created_at, updated_at, deleted_at
		 FROM stakeholders WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&sh.ID, &sh.CompanyID, &sh.Name, &sh.Email, &sh.Role, &sh.CreatedAt, &sh.UpdatedAt, &sh.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "stakeholder", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting stakeholder: %w", err)
	}
	return sh, nil
}

func (s *StakeholderStore) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.Stakeholder, error) {
	if len(ids) == 0 {
		return map[string]*domain.Stakeholder{}, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, email, role, created_at, updated_at, deleted_at
		 FROM stakeholders WHERE id = ANY($1) AND deleted_at IS NULL`, pq.Array(ids),
	)
	if err != nil {
		return nil, fmt.Errorf("batch-fetching stakeholders: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*domain.Stakeholder, len(ids))
	for rows.Next() {
		sh := &domain.Stakeholder{}
		if err := rows.Scan(&sh.ID, &sh.CompanyID, &sh.Name, &sh.Email, &sh.Role, &sh.CreatedAt, &sh.UpdatedAt, &sh.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning stakeholder: %w", err)
		}
		result[sh.ID] = sh
	}
	return result, rows.Err()
}

func (s *StakeholderStore) ListByCompany(ctx context.Context, companyID string) ([]domain.Stakeholder, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, email, role, created_at, updated_at, deleted_at
		 FROM stakeholders WHERE company_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing stakeholders: %w", err)
	}
	defer rows.Close()

	var result []domain.Stakeholder
	for rows.Next() {
		var sh domain.Stakeholder
		if err := rows.Scan(&sh.ID, &sh.CompanyID, &sh.Name, &sh.Email, &sh.Role, &sh.CreatedAt, &sh.UpdatedAt, &sh.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning stakeholder: %w", err)
		}
		result = append(result, sh)
	}
	return result, rows.Err()
}
