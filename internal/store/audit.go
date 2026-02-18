package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hutfut/vestigo/internal/domain"
)

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

func (s *AuditStore) Log(ctx context.Context, entry *domain.AuditEntry) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO audit_log (entity_type, entity_id, action, actor_id, before_state, after_state)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		entry.EntityType, entry.EntityID, entry.Action, entry.ActorID,
		entry.BeforeState, entry.AfterState,
	).Scan(&entry.ID, &entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("writing audit log: %w", err)
	}
	return nil
}

func (s *AuditStore) ListByEntity(ctx context.Context, entityType string, entityID string) ([]domain.AuditEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, entity_type, entity_id, action, actor_id, before_state, after_state, created_at
		 FROM audit_log WHERE entity_type = $1 AND entity_id = $2
		 ORDER BY created_at DESC`, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing audit log: %w", err)
	}
	defer rows.Close()

	var result []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &e.Action, &e.ActorID,
			&e.BeforeState, &e.AfterState, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning audit entry: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
