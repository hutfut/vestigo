package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type VestingScheduleStore struct {
	db *sql.DB
}

func NewVestingScheduleStore(db *sql.DB) *VestingScheduleStore {
	return &VestingScheduleStore{db: db}
}

func (s *VestingScheduleStore) Create(ctx context.Context, vs *domain.VestingSchedule) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO vesting_schedules (cliff_months, total_months, frequency, acceleration_trigger)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		vs.CliffMonths, vs.TotalMonths, vs.Frequency, vs.AccelerationTrigger,
	).Scan(&vs.ID, &vs.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating vesting schedule: %w", err)
	}
	return nil
}

func (s *VestingScheduleStore) GetByID(ctx context.Context, id string) (*domain.VestingSchedule, error) {
	vs := &domain.VestingSchedule{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, cliff_months, total_months, frequency, acceleration_trigger, created_at
		 FROM vesting_schedules WHERE id = $1`, id,
	).Scan(&vs.ID, &vs.CliffMonths, &vs.TotalMonths, &vs.Frequency, &vs.AccelerationTrigger, &vs.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "vesting_schedule", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting vesting schedule: %w", err)
	}
	return vs, nil
}
