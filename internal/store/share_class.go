package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type ShareClassStore struct {
	db *sql.DB
}

func NewShareClassStore(db *sql.DB) *ShareClassStore {
	return &ShareClassStore{db: db}
}

func (s *ShareClassStore) Create(ctx context.Context, sc *domain.ShareClass) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO share_classes
		 (company_id, name, is_preferred, liquidation_multiple, is_participating,
		  participation_cap, price_per_share, seniority, authorized_shares)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		sc.CompanyID, sc.Name, sc.IsPreferred, sc.LiquidationMultiple, sc.IsParticipating,
		decimalPtrToNullString(sc.ParticipationCap), decimalPtrToNullString(sc.PricePerShare),
		sc.Seniority, sc.AuthorizedShares,
	).Scan(&sc.ID, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating share class: %w", err)
	}
	return nil
}

func (s *ShareClassStore) GetByID(ctx context.Context, id string) (*domain.ShareClass, error) {
	sc := &domain.ShareClass{}
	var participationCap, pricePerShare sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, name, is_preferred, liquidation_multiple, is_participating,
		        participation_cap, price_per_share, seniority, authorized_shares,
		        created_at, updated_at, deleted_at
		 FROM share_classes WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&sc.ID, &sc.CompanyID, &sc.Name, &sc.IsPreferred, &sc.LiquidationMultiple,
		&sc.IsParticipating, &participationCap, &pricePerShare, &sc.Seniority,
		&sc.AuthorizedShares, &sc.CreatedAt, &sc.UpdatedAt, &sc.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrNotFound{Entity: "share_class", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("getting share class: %w", err)
	}
	sc.ParticipationCap = nullStringToDecimalPtr(participationCap)
	sc.PricePerShare = nullStringToDecimalPtr(pricePerShare)
	return sc, nil
}

func (s *ShareClassStore) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.ShareClass, error) {
	if len(ids) == 0 {
		return map[string]*domain.ShareClass{}, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, is_preferred, liquidation_multiple, is_participating,
		        participation_cap, price_per_share, seniority, authorized_shares,
		        created_at, updated_at, deleted_at
		 FROM share_classes WHERE id = ANY($1) AND deleted_at IS NULL`, pq.Array(ids),
	)
	if err != nil {
		return nil, fmt.Errorf("batch-fetching share classes: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*domain.ShareClass, len(ids))
	for rows.Next() {
		sc := &domain.ShareClass{}
		var participationCap, pricePerShare sql.NullString
		if err := rows.Scan(&sc.ID, &sc.CompanyID, &sc.Name, &sc.IsPreferred, &sc.LiquidationMultiple,
			&sc.IsParticipating, &participationCap, &pricePerShare, &sc.Seniority,
			&sc.AuthorizedShares, &sc.CreatedAt, &sc.UpdatedAt, &sc.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning share class: %w", err)
		}
		sc.ParticipationCap = nullStringToDecimalPtr(participationCap)
		sc.PricePerShare = nullStringToDecimalPtr(pricePerShare)
		result[sc.ID] = sc
	}
	return result, rows.Err()
}

func (s *ShareClassStore) ListByCompany(ctx context.Context, companyID string) ([]domain.ShareClass, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, is_preferred, liquidation_multiple, is_participating,
		        participation_cap, price_per_share, seniority, authorized_shares,
		        created_at, updated_at, deleted_at
		 FROM share_classes WHERE company_id = $1 AND deleted_at IS NULL
		 ORDER BY seniority, created_at`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing share classes: %w", err)
	}
	defer rows.Close()

	var result []domain.ShareClass
	for rows.Next() {
		var sc domain.ShareClass
		var participationCap, pricePerShare sql.NullString
		if err := rows.Scan(&sc.ID, &sc.CompanyID, &sc.Name, &sc.IsPreferred, &sc.LiquidationMultiple,
			&sc.IsParticipating, &participationCap, &pricePerShare, &sc.Seniority,
			&sc.AuthorizedShares, &sc.CreatedAt, &sc.UpdatedAt, &sc.DeletedAt); err != nil {
			return nil, fmt.Errorf("scanning share class: %w", err)
		}
		sc.ParticipationCap = nullStringToDecimalPtr(participationCap)
		sc.PricePerShare = nullStringToDecimalPtr(pricePerShare)
		result = append(result, sc)
	}
	return result, rows.Err()
}

func decimalPtrToNullString(d *decimal.Decimal) sql.NullString {
	if d == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: d.String(), Valid: true}
}

func nullStringToDecimalPtr(ns sql.NullString) *decimal.Decimal {
	if !ns.Valid {
		return nil
	}
	d, _ := decimal.NewFromString(ns.String)
	return &d
}
