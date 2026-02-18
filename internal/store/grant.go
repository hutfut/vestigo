package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type GrantStore struct {
	db *sql.DB
}

func NewGrantStore(db *sql.DB) *GrantStore {
	return &GrantStore{db: db}
}

func (s *GrantStore) Create(ctx context.Context, g *domain.Grant) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO grants
		 (company_id, stakeholder_id, share_class_id, vesting_schedule_id, quantity, grant_date, exercise_price, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		g.CompanyID, g.StakeholderID, g.ShareClassID, g.VestingScheduleID,
		g.Quantity, g.GrantDate, g.ExercisePrice, g.Notes,
	).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating grant: %w", err)
	}
	return nil
}

func (s *GrantStore) GetByID(ctx context.Context, id string) (*domain.Grant, error) {
	g := &domain.Grant{}
	err := s.db.QueryRowContext(ctx,
		`SELECT g.id, g.company_id, g.stakeholder_id, g.share_class_id, g.vesting_schedule_id,
		        g.quantity, g.grant_date, g.exercise_price, g.is_exercised, g.notes,
		        g.created_at, g.updated_at, g.deleted_at
		 FROM grants g WHERE g.id = $1 AND g.deleted_at IS NULL`, id,
	).Scan(&g.ID, &g.CompanyID, &g.StakeholderID, &g.ShareClassID, &g.VestingScheduleID,
		&g.Quantity, &g.GrantDate, &g.ExercisePrice, &g.IsExercised, &g.Notes,
		&g.CreatedAt, &g.UpdatedAt, &g.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "grant", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting grant: %w", err)
	}
	return g, nil
}

func (s *GrantStore) ListByCompany(ctx context.Context, companyID string) ([]domain.Grant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.id, g.company_id, g.stakeholder_id, g.share_class_id, g.vesting_schedule_id,
		        g.quantity, g.grant_date, g.exercise_price, g.is_exercised, g.notes,
		        g.created_at, g.updated_at, g.deleted_at
		 FROM grants g WHERE g.company_id = $1 AND g.deleted_at IS NULL
		 ORDER BY g.grant_date`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing grants: %w", err)
	}
	defer rows.Close()

	var result []domain.Grant
	for rows.Next() {
		var g domain.Grant
		if err := rows.Scan(&g.ID, &g.CompanyID, &g.StakeholderID, &g.ShareClassID, &g.VestingScheduleID,
			&g.Quantity, &g.GrantDate, &g.ExercisePrice, &g.IsExercised, &g.Notes,
			&g.CreatedAt, &g.UpdatedAt, &g.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning grant: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (s *GrantStore) ListByStakeholder(ctx context.Context, stakeholderID string) ([]domain.Grant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.id, g.company_id, g.stakeholder_id, g.share_class_id, g.vesting_schedule_id,
		        g.quantity, g.grant_date, g.exercise_price, g.is_exercised, g.notes,
		        g.created_at, g.updated_at, g.deleted_at
		 FROM grants g WHERE g.stakeholder_id = $1 AND g.deleted_at IS NULL
		 ORDER BY g.grant_date`, stakeholderID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing grants by stakeholder: %w", err)
	}
	defer rows.Close()

	var result []domain.Grant
	for rows.Next() {
		var g domain.Grant
		if err := rows.Scan(&g.ID, &g.CompanyID, &g.StakeholderID, &g.ShareClassID, &g.VestingScheduleID,
			&g.Quantity, &g.GrantDate, &g.ExercisePrice, &g.IsExercised, &g.Notes,
			&g.CreatedAt, &g.UpdatedAt, &g.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning grant: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}
