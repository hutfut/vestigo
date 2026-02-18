package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type SAFENoteStore struct {
	db *sql.DB
}

func NewSAFENoteStore(db *sql.DB) *SAFENoteStore {
	return &SAFENoteStore{db: db}
}

func (s *SAFENoteStore) Create(ctx context.Context, sn *domain.SAFENote) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO safe_notes
		 (company_id, stakeholder_id, investment_amount, valuation_cap, discount_rate, safe_type, issue_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		sn.CompanyID, sn.StakeholderID, sn.InvestmentAmount,
		decimalPtrToNullString(sn.ValuationCap), decimalPtrToNullString(sn.DiscountRate),
		sn.SAFEType, sn.IssueDate,
	).Scan(&sn.ID, &sn.CreatedAt, &sn.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating SAFE note: %w", err)
	}
	return nil
}

func (s *SAFENoteStore) GetByID(ctx context.Context, id string) (*domain.SAFENote, error) {
	sn := &domain.SAFENote{}
	var valCap, discRate sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, stakeholder_id, investment_amount, valuation_cap, discount_rate,
		        safe_type, is_converted, converted_in_round, issue_date,
		        created_at, updated_at, deleted_at
		 FROM safe_notes WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&sn.ID, &sn.CompanyID, &sn.StakeholderID, &sn.InvestmentAmount, &valCap, &discRate,
		&sn.SAFEType, &sn.IsConverted, &sn.ConvertedInRound, &sn.IssueDate,
		&sn.CreatedAt, &sn.UpdatedAt, &sn.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "safe_note", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting SAFE note: %w", err)
	}
	sn.ValuationCap = nullStringToDecimalPtr(valCap)
	sn.DiscountRate = nullStringToDecimalPtr(discRate)
	return sn, nil
}

func (s *SAFENoteStore) ListByCompany(ctx context.Context, companyID string) ([]domain.SAFENote, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, stakeholder_id, investment_amount, valuation_cap, discount_rate,
		        safe_type, is_converted, converted_in_round, issue_date,
		        created_at, updated_at, deleted_at
		 FROM safe_notes WHERE company_id = $1 AND deleted_at IS NULL
		 ORDER BY issue_date`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing SAFE notes: %w", err)
	}
	defer rows.Close()

	var result []domain.SAFENote
	for rows.Next() {
		var sn domain.SAFENote
		var valCap, discRate sql.NullString
		if err := rows.Scan(&sn.ID, &sn.CompanyID, &sn.StakeholderID, &sn.InvestmentAmount, &valCap, &discRate,
			&sn.SAFEType, &sn.IsConverted, &sn.ConvertedInRound, &sn.IssueDate,
			&sn.CreatedAt, &sn.UpdatedAt, &sn.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning SAFE note: %w", err)
		}
		sn.ValuationCap = nullStringToDecimalPtr(valCap)
		sn.DiscountRate = nullStringToDecimalPtr(discRate)
		result = append(result, sn)
	}
	return result, rows.Err()
}

func (s *SAFENoteStore) MarkConverted(ctx context.Context, id string, roundID string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE safe_notes SET is_converted = true, converted_in_round = $2
		 WHERE id = $1 AND deleted_at IS NULL`, id, roundID,
	)
	if err != nil {
		return fmt.Errorf("marking SAFE as converted: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &domain.ErrNotFound{Entity: "safe_note", ID: id}
	}
	return nil
}
