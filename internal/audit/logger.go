package audit

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hutfut/vestigo/internal/domain"
)

// Logger wraps the AuditRepository to provide a convenient API for recording
// audit events. It serializes before/after state as JSON and handles errors
// gracefully (logs but does not propagate) to avoid blocking mutations on
// audit failures.
type Logger struct {
	repo domain.AuditRepository
}

func NewLogger(repo domain.AuditRepository) *Logger {
	return &Logger{repo: repo}
}

func (l *Logger) Record(ctx context.Context, entityType, entityID, action string, before, after interface{}) {
	entry := &domain.AuditEntry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
	}

	if before != nil {
		bs, err := json.Marshal(before)
		if err != nil {
			log.Printf("audit: failed to marshal before state: %v", err)
		} else {
			entry.BeforeState = bs
		}
	}

	if after != nil {
		as, err := json.Marshal(after)
		if err != nil {
			log.Printf("audit: failed to marshal after state: %v", err)
		} else {
			entry.AfterState = as
		}
	}

	if err := l.repo.Log(ctx, entry); err != nil {
		log.Printf("audit: failed to write log entry: %v", err)
	}
}
