package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type FundingRoundStore struct {
	db *sql.DB
}

func NewFundingRoundStore(db *sql.DB) *FundingRoundStore {
	return &FundingRoundStore{db: db}
}

func (s *FundingRoundStore) Create(ctx context.Context, fr *domain.FundingRound) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO funding_rounds
		 (company_id, name, pre_money_valuation, amount_raised, price_per_share, share_class_id, round_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		fr.CompanyID, fr.Name, fr.PreMoneyVal, fr.AmountRaised, fr.PricePerShare, fr.ShareClassID, fr.RoundDate,
	).Scan(&fr.ID, &fr.CreatedAt, &fr.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating funding round: %w", err)
	}
	return nil
}

func (s *FundingRoundStore) GetByID(ctx context.Context, id string) (*domain.FundingRound, error) {
	fr := &domain.FundingRound{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, name, pre_money_valuation, amount_raised, price_per_share,
		        share_class_id, round_date, created_at, updated_at, deleted_at
		 FROM funding_rounds WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&fr.ID, &fr.CompanyID, &fr.Name, &fr.PreMoneyVal, &fr.AmountRaised, &fr.PricePerShare,
		&fr.ShareClassID, &fr.RoundDate, &fr.CreatedAt, &fr.UpdatedAt, &fr.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "funding_round", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting funding round: %w", err)
	}
	return fr, nil
}

func (s *FundingRoundStore) ListByCompany(ctx context.Context, companyID string) ([]domain.FundingRound, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, pre_money_valuation, amount_raised, price_per_share,
		        share_class_id, round_date, created_at, updated_at, deleted_at
		 FROM funding_rounds WHERE company_id = $1 AND deleted_at IS NULL
		 ORDER BY round_date`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing funding rounds: %w", err)
	}
	defer rows.Close()

	var result []domain.FundingRound
	for rows.Next() {
		var fr domain.FundingRound
		if err := rows.Scan(&fr.ID, &fr.CompanyID, &fr.Name, &fr.PreMoneyVal, &fr.AmountRaised,
			&fr.PricePerShare, &fr.ShareClassID, &fr.RoundDate, &fr.CreatedAt, &fr.UpdatedAt, &fr.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning funding round: %w", err)
		}
		result = append(result, fr)
	}
	return result, rows.Err()
}
