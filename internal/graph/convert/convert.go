package convert

import (
	"strings"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/hutfut/vestigo/internal/graph/model"
	"github.com/shopspring/decimal"
)

// ─── Domain → GraphQL ─────────────────────────────────────────────────────────

func ToGQLCompany(c *domain.Company) *model.Company {
	return &model.Company{
		ID:            c.ID,
		Name:          c.Name,
		Stakeholders:  []*model.Stakeholder{},
		ShareClasses:  []*model.ShareClass{},
		Grants:        []*model.Grant{},
		FundingRounds: []*model.FundingRound{},
		SafeNotes:     []*model.SAFENote{},
		CreatedAt:     model.DateTime(c.CreatedAt),
	}
}

func ToGQLStakeholder(sh *domain.Stakeholder) *model.Stakeholder {
	return &model.Stakeholder{
		ID:        sh.ID,
		CompanyID: sh.CompanyID,
		Name:      sh.Name,
		Email:     sh.Email,
		Role:      DomainRoleToGQL(sh.Role),
		Grants:    []*model.Grant{},
		CreatedAt: model.DateTime(sh.CreatedAt),
	}
}

func ToGQLShareClass(sc *domain.ShareClass) *model.ShareClass {
	return &model.ShareClass{
		ID:                  sc.ID,
		CompanyID:           sc.CompanyID,
		Name:                sc.Name,
		IsPreferred:         sc.IsPreferred,
		LiquidationMultiple: model.Decimal(sc.LiquidationMultiple),
		IsParticipating:     sc.IsParticipating,
		ParticipationCap:    DecPtrToGQLDecPtr(sc.ParticipationCap),
		PricePerShare:       DecPtrToGQLDecPtr(sc.PricePerShare),
		Seniority:           sc.Seniority,
		AuthorizedShares:    model.Decimal(sc.AuthorizedShares),
		CreatedAt:           model.DateTime(sc.CreatedAt),
	}
}

func ToGQLVestingSchedule(vs *domain.VestingSchedule) *model.VestingSchedule {
	return &model.VestingSchedule{
		ID:                  vs.ID,
		CliffMonths:         vs.CliffMonths,
		TotalMonths:         vs.TotalMonths,
		Frequency:           DomainFreqToGQL(vs.Frequency),
		AccelerationTrigger: DomainAccelToGQL(vs.AccelerationTrigger),
	}
}

func ToGQLGrant(g *domain.Grant) *model.Grant {
	mg := &model.Grant{
		ID:                g.ID,
		CompanyID:         g.CompanyID,
		StakeholderID:     g.StakeholderID,
		ShareClassID:      g.ShareClassID,
		VestingScheduleID: g.VestingScheduleID,
		Quantity:          model.Decimal(g.Quantity),
		GrantDate:         model.Date(g.GrantDate),
		ExercisePrice:     model.Decimal(g.ExercisePrice),
		IsExercised:       g.IsExercised,
		Notes:             g.Notes,
		CreatedAt:         model.DateTime(g.CreatedAt),
	}
	if g.VestingSchedule != nil {
		mg.VestingSchedule = ToGQLVestingSchedule(g.VestingSchedule)
	}
	return mg
}

func ToGQLFundingRound(fr *domain.FundingRound) *model.FundingRound {
	return &model.FundingRound{
		ID:                fr.ID,
		CompanyID:         fr.CompanyID,
		Name:              fr.Name,
		PreMoneyValuation: model.Decimal(fr.PreMoneyVal),
		AmountRaised:      model.Decimal(fr.AmountRaised),
		PricePerShare:     model.Decimal(fr.PricePerShare),
		ShareClassID:      fr.ShareClassID,
		RoundDate:         model.Date(fr.RoundDate),
		CreatedAt:         model.DateTime(fr.CreatedAt),
	}
}

func ToGQLSAFENote(sn *domain.SAFENote) *model.SAFENote {
	return &model.SAFENote{
		ID:               sn.ID,
		CompanyID:        sn.CompanyID,
		StakeholderID:    sn.StakeholderID,
		InvestmentAmount: model.Decimal(sn.InvestmentAmount),
		ValuationCap:     DecPtrToGQLDecPtr(sn.ValuationCap),
		DiscountRate:     DecPtrToGQLDecPtr(sn.DiscountRate),
		SafeType:         DomainSAFETypeToGQL(sn.SAFEType),
		IsConverted:      sn.IsConverted,
		ConvertedInRound: sn.ConvertedInRound,
		IssueDate:        model.Date(sn.IssueDate),
		CreatedAt:        model.DateTime(sn.CreatedAt),
	}
}

func ToGQLVestingStatus(vs *domain.VestingStatus) *model.VestingStatus {
	return &model.VestingStatus{
		GrantID:        vs.GrantID,
		AsOfDate:       model.Date(vs.AsOfDate),
		TotalShares:    model.Decimal(vs.TotalShares),
		VestedShares:   model.Decimal(vs.VestedShares),
		UnvestedShares: model.Decimal(vs.UnvestedShares),
		PercentVested:  model.Decimal(vs.PercentVested),
		CliffDate:      model.Date(vs.CliffDate),
		FullyVestedAt:  model.Date(vs.FullyVestedAt),
		IsFullyVested:  vs.IsFullyVested,
	}
}

func ToGQLDilutionResult(r *domain.DilutionResult) *model.DilutionResult {
	return &model.DilutionResult{
		PreRound:    ToGQLCapTableSnapshot(&r.PreRound),
		PostRound:   ToGQLCapTableSnapshot(&r.PostRound),
		NewInvestor: ToGQLCapTableEntry(&r.NewInvestor),
		RoundName:   r.RoundName,
	}
}

func ToGQLCapTableSnapshot(s *domain.CapTableSnapshot) *model.CapTableSnapshot {
	entries := make([]*model.CapTableEntry, len(s.Entries))
	for i := range s.Entries {
		entries[i] = ToGQLCapTableEntry(&s.Entries[i])
	}
	return &model.CapTableSnapshot{
		CompanyID:   &s.CompanyID,
		TotalShares: model.Decimal(s.TotalShares),
		Entries:     entries,
	}
}

func ToGQLCapTableEntry(e *domain.CapTableEntry) *model.CapTableEntry {
	var shID *string
	if e.StakeholderID != "" {
		shID = &e.StakeholderID
	}
	return &model.CapTableEntry{
		StakeholderID:   shID,
		StakeholderName: e.StakeholderName,
		ShareClassName:  e.ShareClassName,
		Shares:          model.Decimal(e.Shares),
		OwnershipPct:    model.Decimal(e.OwnershipPct),
	}
}

// ─── Enum converters ──────────────────────────────────────────────────────────

func GQLRoleToDomain(r model.StakeholderRole) domain.StakeholderRole {
	return domain.StakeholderRole(strings.ToLower(string(r)))
}

func DomainRoleToGQL(r domain.StakeholderRole) model.StakeholderRole {
	return model.StakeholderRole(strings.ToUpper(string(r)))
}

func GQLFreqToDomain(f model.VestingFrequency) domain.VestingFrequency {
	return domain.VestingFrequency(strings.ToLower(string(f)))
}

func DomainFreqToGQL(f domain.VestingFrequency) model.VestingFrequency {
	return model.VestingFrequency(strings.ToUpper(string(f)))
}

func GQLAccelToDomain(a model.AccelerationTrigger) domain.AccelerationTrigger {
	return domain.AccelerationTrigger(strings.ToLower(string(a)))
}

func DomainAccelToGQL(a domain.AccelerationTrigger) model.AccelerationTrigger {
	return model.AccelerationTrigger(strings.ToUpper(string(a)))
}

func GQLSAFETypeToDomain(t model.SAFEType) domain.SAFEType {
	return domain.SAFEType(strings.ToLower(string(t)))
}

func DomainSAFETypeToGQL(t domain.SAFEType) model.SAFEType {
	return model.SAFEType(strings.ToUpper(string(t)))
}

// ─── Decimal / input helpers ──────────────────────────────────────────────────

func GQLDecToDecPtr(d *model.Decimal) *decimal.Decimal {
	if d == nil {
		return nil
	}
	v := decimal.Decimal(*d)
	return &v
}

func DecPtrToGQLDecPtr(d *decimal.Decimal) *model.Decimal {
	if d == nil {
		return nil
	}
	v := model.Decimal(*d)
	return &v
}

func DecOrDefault(d *model.Decimal, def decimal.Decimal) decimal.Decimal {
	if d == nil {
		return def
	}
	return decimal.Decimal(*d)
}

func BoolOrDefault(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func IntOrDefault(i *int, def int) int {
	if i == nil {
		return def
	}
	return *i
}
