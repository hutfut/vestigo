package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type StakeholderRole string

const (
	RoleFounder    StakeholderRole = "founder"
	RoleEmployee   StakeholderRole = "employee"
	RoleInvestor   StakeholderRole = "investor"
	RoleAdvisor    StakeholderRole = "advisor"
	RoleConsultant StakeholderRole = "consultant"
)

type SAFEType string

const (
	SAFEPreMoney  SAFEType = "pre_money"
	SAFEPostMoney SAFEType = "post_money"
)

type VestingFrequency string

const (
	FrequencyMonthly   VestingFrequency = "monthly"
	FrequencyQuarterly VestingFrequency = "quarterly"
	FrequencyAnnually  VestingFrequency = "annually"
)

type AccelerationTrigger string

const (
	AccelerationNone          AccelerationTrigger = "none"
	AccelerationSingleTrigger AccelerationTrigger = "single_trigger"
	AccelerationDoubleTrigger AccelerationTrigger = "double_trigger"
)

type Company struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Stakeholder struct {
	ID        string
	CompanyID string
	Name      string
	Email     string
	Role      StakeholderRole
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ShareClass struct {
	ID                   string
	CompanyID            string
	Name                 string
	IsPreferred          bool
	LiquidationMultiple  decimal.Decimal
	IsParticipating      bool
	ParticipationCap     *decimal.Decimal // nil = uncapped
	PricePerShare        *decimal.Decimal
	Seniority            int
	AuthorizedShares     decimal.Decimal
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

type VestingSchedule struct {
	ID                  string
	CliffMonths         int
	TotalMonths         int
	Frequency           VestingFrequency
	AccelerationTrigger AccelerationTrigger
	CreatedAt           time.Time
}

type Grant struct {
	ID                string
	CompanyID         string
	StakeholderID     string
	ShareClassID      string
	VestingScheduleID *string
	Quantity          decimal.Decimal
	GrantDate         time.Time
	ExercisePrice     decimal.Decimal
	IsExercised       bool
	Notes             *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time

	VestingSchedule *VestingSchedule
}

type FundingRound struct {
	ID               string
	CompanyID        string
	Name             string
	PreMoneyVal      decimal.Decimal
	AmountRaised     decimal.Decimal
	PricePerShare    decimal.Decimal
	ShareClassID     string
	RoundDate        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type SAFENote struct {
	ID               string
	CompanyID        string
	StakeholderID    string
	InvestmentAmount decimal.Decimal
	ValuationCap     *decimal.Decimal
	DiscountRate     *decimal.Decimal // 0.20 = 20%
	SAFEType         SAFEType
	IsConverted      bool
	ConvertedInRound *string
	IssueDate        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type AuditEntry struct {
	ID          string
	EntityType  string
	EntityID    string
	Action      string
	ActorID     *string
	BeforeState []byte // JSON
	AfterState  []byte // JSON
	CreatedAt   time.Time
}

// --- Computed types (not persisted, returned by engines) ---

type VestingStatus struct {
	GrantID        string
	AsOfDate       time.Time
	TotalShares    decimal.Decimal
	VestedShares   decimal.Decimal
	UnvestedShares decimal.Decimal
	PercentVested  decimal.Decimal
	CliffDate      time.Time
	FullyVestedAt  time.Time
	IsFullyVested  bool
}

type SAFEConversionResult struct {
	SAFEID            string
	SharesIssued      decimal.Decimal
	EffectivePPS      decimal.Decimal
	ConversionMethod  string // "cap", "discount", or "round_price"
}

type CapTableEntry struct {
	StakeholderID   string
	StakeholderName string
	ShareClassName  string
	Shares          decimal.Decimal
	OwnershipPct    decimal.Decimal
}

type CapTableSnapshot struct {
	CompanyID        string
	AsOfDate         time.Time
	TotalShares      decimal.Decimal
	Entries          []CapTableEntry
}

type DilutionResult struct {
	PreRound     CapTableSnapshot
	PostRound    CapTableSnapshot
	NewInvestor  CapTableEntry
	RoundName    string
}

type WaterfallPayout struct {
	StakeholderID   string
	StakeholderName string
	ShareClassName  string
	Shares          decimal.Decimal
	Payout          decimal.Decimal
	PayoutPerShare  decimal.Decimal
}

type WaterfallResult struct {
	ExitValuation  decimal.Decimal
	TotalPayout    decimal.Decimal
	Payouts        []WaterfallPayout
}
