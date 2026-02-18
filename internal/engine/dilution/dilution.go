package dilution

import (
	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

// StakeholderShares is a simplified input representing a stakeholder's current
// position on the cap table. This decouples the engine from the persistence layer.
type StakeholderShares struct {
	StakeholderID   string
	StakeholderName string
	ShareClassName  string
	Shares          decimal.Decimal
}

// RoundInput describes a hypothetical funding round for dilution modeling.
type RoundInput struct {
	RoundName     string
	PreMoneyVal   decimal.Decimal
	AmountRaised  decimal.Decimal
	NewShareClass string
	InvestorName  string
}

// Model calculates the dilution impact of a hypothetical funding round on the
// current cap table. It returns pre-round and post-round snapshots plus the
// new investor's entry.
func Model(existing []StakeholderShares, input RoundInput) domain.DilutionResult {
	hundred := decimal.NewFromInt(100)

	totalExisting := decimal.Zero
	for _, s := range existing {
		totalExisting = totalExisting.Add(s.Shares)
	}

	postMoneyVal := input.PreMoneyVal.Add(input.AmountRaised)
	pps := input.PreMoneyVal.Div(totalExisting)
	newShares := input.AmountRaised.Div(pps).RoundFloor(4)
	totalPost := totalExisting.Add(newShares)

	// Pre-round snapshot
	preEntries := make([]domain.CapTableEntry, len(existing))
	for i, s := range existing {
		preEntries[i] = domain.CapTableEntry{
			StakeholderID:   s.StakeholderID,
			StakeholderName: s.StakeholderName,
			ShareClassName:  s.ShareClassName,
			Shares:          s.Shares,
			OwnershipPct:    s.Shares.Div(totalExisting).Mul(hundred).RoundFloor(4),
		}
	}

	// Post-round snapshot
	postEntries := make([]domain.CapTableEntry, len(existing))
	for i, s := range existing {
		postEntries[i] = domain.CapTableEntry{
			StakeholderID:   s.StakeholderID,
			StakeholderName: s.StakeholderName,
			ShareClassName:  s.ShareClassName,
			Shares:          s.Shares,
			OwnershipPct:    s.Shares.Div(totalPost).Mul(hundred).RoundFloor(4),
		}
	}

	newInvestor := domain.CapTableEntry{
		StakeholderName: input.InvestorName,
		ShareClassName:  input.NewShareClass,
		Shares:          newShares,
		OwnershipPct:    newShares.Div(totalPost).Mul(hundred).RoundFloor(4),
	}

	_ = postMoneyVal

	return domain.DilutionResult{
		PreRound: domain.CapTableSnapshot{
			TotalShares: totalExisting,
			Entries:     preEntries,
		},
		PostRound: domain.CapTableSnapshot{
			TotalShares: totalPost,
			Entries:     append(postEntries, newInvestor),
		},
		NewInvestor: newInvestor,
		RoundName:   input.RoundName,
	}
}
