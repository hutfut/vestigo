package domain

import "context"

type CompanyRepository interface {
	Create(ctx context.Context, c *Company) error
	GetByID(ctx context.Context, id string) (*Company, error)
}

type StakeholderRepository interface {
	Create(ctx context.Context, s *Stakeholder) error
	GetByID(ctx context.Context, id string) (*Stakeholder, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]*Stakeholder, error)
	ListByCompany(ctx context.Context, companyID string) ([]Stakeholder, error)
}

type ShareClassRepository interface {
	Create(ctx context.Context, sc *ShareClass) error
	GetByID(ctx context.Context, id string) (*ShareClass, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]*ShareClass, error)
	ListByCompany(ctx context.Context, companyID string) ([]ShareClass, error)
}

type VestingScheduleRepository interface {
	Create(ctx context.Context, vs *VestingSchedule) error
	GetByID(ctx context.Context, id string) (*VestingSchedule, error)
}

type GrantRepository interface {
	Create(ctx context.Context, g *Grant) error
	GetByID(ctx context.Context, id string) (*Grant, error)
	ListByCompany(ctx context.Context, companyID string) ([]Grant, error)
	ListByStakeholder(ctx context.Context, stakeholderID string) ([]Grant, error)
}

type FundingRoundRepository interface {
	Create(ctx context.Context, fr *FundingRound) error
	GetByID(ctx context.Context, id string) (*FundingRound, error)
	ListByCompany(ctx context.Context, companyID string) ([]FundingRound, error)
}

type SAFENoteRepository interface {
	Create(ctx context.Context, s *SAFENote) error
	GetByID(ctx context.Context, id string) (*SAFENote, error)
	ListByCompany(ctx context.Context, companyID string) ([]SAFENote, error)
	MarkConverted(ctx context.Context, id string, roundID string) error
}

type AuditRepository interface {
	Log(ctx context.Context, entry *AuditEntry) error
	ListByEntity(ctx context.Context, entityType string, entityID string) ([]AuditEntry, error)
}
