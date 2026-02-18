package waterfall

import (
	"testing"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

func dec(v string) decimal.Decimal {
	d, _ := decimal.NewFromString(v)
	return d
}

func decPtr(v string) *decimal.Decimal {
	d := dec(v)
	return &d
}

func ppsPtr(v string) *decimal.Decimal {
	return decPtr(v)
}

func TestCalculate_SimplePreferred(t *testing.T) {
	// Scenario: $10M exit, Series A has 1x non-participating preference on $3M invested.
	// Series A gets $3M preference, remaining $7M goes to common.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("3000000")},
			},
			TotalShares: dec("3000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("7000000")},
			},
			TotalShares: dec("7000000"),
		},
	}

	result := Calculate(positions, dec("10000000"))

	investorPayout := findPayout(result, "inv1")
	founderPayout := findPayout(result, "f1")

	if investorPayout == nil {
		t.Fatal("expected payout for investor")
	}
	if founderPayout == nil {
		t.Fatal("expected payout for founder")
	}

	if !investorPayout.Payout.Equal(dec("3000000")) {
		t.Errorf("investor payout = %s, want 3000000", investorPayout.Payout)
	}
	if !founderPayout.Payout.Equal(dec("7000000")) {
		t.Errorf("founder payout = %s, want 7000000", founderPayout.Payout)
	}
}

func TestCalculate_ParticipatingPreferred(t *testing.T) {
	// Scenario: $20M exit. Series A has 1x participating (uncapped) on $5M invested.
	// 5M preferred shares + 5M common = 10M total.
	// Step 1: Series A gets $5M preference.
	// Step 2: Remaining $15M split pro-rata. Each has 50% → $7.5M each.
	// Series A total: $5M + $7.5M = $12.5M. Common: $7.5M.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     true,
				ParticipationCap:    nil,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
	}

	result := Calculate(positions, dec("20000000"))

	investorPayout := findPayout(result, "inv1")
	founderPayout := findPayout(result, "f1")

	if investorPayout == nil {
		t.Fatal("expected payout for investor")
	}
	if founderPayout == nil {
		t.Fatal("expected payout for founder")
	}

	if !investorPayout.Payout.Equal(dec("12500000")) {
		t.Errorf("investor payout = %s, want 12500000", investorPayout.Payout)
	}
	if !founderPayout.Payout.Equal(dec("7500000")) {
		t.Errorf("founder payout = %s, want 7500000", founderPayout.Payout)
	}
}

func TestCalculate_InsufficientProceeds(t *testing.T) {
	// Exit at $2M, but Series A has $5M preference. Series A gets everything.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
	}

	result := Calculate(positions, dec("2000000"))

	investorPayout := findPayout(result, "inv1")
	founderPayout := findPayout(result, "f1")

	if investorPayout == nil || !investorPayout.Payout.Equal(dec("2000000")) {
		got := "nil"
		if investorPayout != nil {
			got = investorPayout.Payout.String()
		}
		t.Errorf("investor payout = %s, want 2000000", got)
	}
	if founderPayout != nil {
		t.Errorf("founder should get nothing, got %s", founderPayout.Payout)
	}
}

func TestCalculate_SeniorityOrder(t *testing.T) {
	// Series B (seniority 2) gets paid before Series A (seniority 1).
	// $8M exit. Series B preference = $5M, Series A preference = $5M.
	// Series B gets $5M, Series A gets remaining $3M, Common gets nothing.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv_a", StakeholderName: "Investor A", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred B",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("2.00"),
				Seniority:           2,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv_b", StakeholderName: "Investor B", Shares: dec("2500000")},
			},
			TotalShares: dec("2500000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
	}

	result := Calculate(positions, dec("8000000"))

	invB := findPayout(result, "inv_b")
	invA := findPayout(result, "inv_a")

	if invB == nil || !invB.Payout.Equal(dec("5000000")) {
		got := "nil"
		if invB != nil {
			got = invB.Payout.String()
		}
		t.Errorf("Series B payout = %s, want 5000000", got)
	}
	if invA == nil || !invA.Payout.Equal(dec("3000000")) {
		got := "nil"
		if invA != nil {
			got = invA.Payout.String()
		}
		t.Errorf("Series A payout = %s, want 3000000", got)
	}
}

func TestCalculate_ZeroExit(t *testing.T) {
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{Name: "Common", IsPreferred: false},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("10000000")},
			},
			TotalShares: dec("10000000"),
		},
	}

	result := Calculate(positions, decimal.Zero)

	if len(result.Payouts) != 0 {
		t.Errorf("expected no payouts for $0 exit, got %d", len(result.Payouts))
	}
}

func TestCalculate_CommonOnly(t *testing.T) {
	// No preferred — all proceeds go pro-rata to common.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{Name: "Common", IsPreferred: false},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Alice", Shares: dec("6000000")},
				{StakeholderID: "f2", StakeholderName: "Bob", Shares: dec("4000000")},
			},
			TotalShares: dec("10000000"),
		},
	}

	result := Calculate(positions, dec("10000000"))

	alice := findPayout(result, "f1")
	bob := findPayout(result, "f2")

	if alice == nil || !alice.Payout.Equal(dec("6000000")) {
		t.Errorf("Alice payout wrong")
	}
	if bob == nil || !bob.Payout.Equal(dec("4000000")) {
		t.Errorf("Bob payout wrong")
	}

	if !result.TotalPayout.Equal(dec("10000000")) {
		t.Errorf("total payout = %s, want 10000000", result.TotalPayout)
	}
}

func TestCalculate_NonParticipatingConverts(t *testing.T) {
	// $20M exit. Series A (non-participating): 3M shares at $1, 1x pref = $3M.
	// Common: 7M shares.
	// Preference = $3M. As-converted = (3M/10M) * $20M = $6M. Conversion wins.
	// All 10M shares share $20M pro-rata.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("3000000")},
			},
			TotalShares: dec("3000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("7000000")},
			},
			TotalShares: dec("7000000"),
		},
	}

	result := Calculate(positions, dec("20000000"))

	inv := findPayout(result, "inv1")
	fdr := findPayout(result, "f1")

	if inv == nil || !inv.Payout.Equal(dec("6000000")) {
		got := "nil"
		if inv != nil {
			got = inv.Payout.String()
		}
		t.Errorf("investor payout = %s, want 6000000 (as-converted)", got)
	}
	if fdr == nil || !fdr.Payout.Equal(dec("14000000")) {
		got := "nil"
		if fdr != nil {
			got = fdr.Payout.String()
		}
		t.Errorf("founder payout = %s, want 14000000", got)
	}
}

func TestCalculate_NonParticipatingKeepsPreference(t *testing.T) {
	// $4M exit. Series A (non-participating): 3M shares at $1, 1x pref = $3M.
	// Common: 7M shares.
	// Preference = $3M. As-converted = (3M/10M) * $4M = $1.2M. Preference wins.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("3000000")},
			},
			TotalShares: dec("3000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("7000000")},
			},
			TotalShares: dec("7000000"),
		},
	}

	result := Calculate(positions, dec("4000000"))

	inv := findPayout(result, "inv1")
	fdr := findPayout(result, "f1")

	if inv == nil || !inv.Payout.Equal(dec("3000000")) {
		got := "nil"
		if inv != nil {
			got = inv.Payout.String()
		}
		t.Errorf("investor payout = %s, want 3000000 (preference)", got)
	}
	if fdr == nil || !fdr.Payout.Equal(dec("1000000")) {
		got := "nil"
		if fdr != nil {
			got = fdr.Payout.String()
		}
		t.Errorf("founder payout = %s, want 1000000", got)
	}
}

func TestCalculate_NonParticipatingBreakpoint(t *testing.T) {
	// At exactly $10M: pref = $3M, as-converted = $3M. Equal → keep preference.
	// At $10,000,010: as-converted = 3M/10M * $10,000,010 = $3,000,003 > $3M → convert.
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Preferred A",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv1", StakeholderName: "Investor A", Shares: dec("3000000")},
			},
			TotalShares: dec("3000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("7000000")},
			},
			TotalShares: dec("7000000"),
		},
	}

	t.Run("at breakpoint, preference wins", func(t *testing.T) {
		result := Calculate(positions, dec("10000000"))
		inv := findPayout(result, "inv1")
		fdr := findPayout(result, "f1")

		if inv == nil || !inv.Payout.Equal(dec("3000000")) {
			t.Errorf("investor payout = %v, want 3000000", inv)
		}
		if fdr == nil || !fdr.Payout.Equal(dec("7000000")) {
			t.Errorf("founder payout = %v, want 7000000", fdr)
		}
	})

	t.Run("above breakpoint, conversion wins", func(t *testing.T) {
		result := Calculate(positions, dec("10000010"))
		inv := findPayout(result, "inv1")
		fdr := findPayout(result, "f1")

		// Converted: 3M/10M * 10,000,010 = 3,000,003
		if inv == nil || !inv.Payout.Equal(dec("3000003")) {
			got := "nil"
			if inv != nil {
				got = inv.Payout.String()
			}
			t.Errorf("investor payout = %s, want 3000003", got)
		}
		// Founder: 7M/10M * 10,000,010 = 7,000,007
		if fdr == nil || !fdr.Payout.Equal(dec("7000007")) {
			got := "nil"
			if fdr != nil {
				got = fdr.Payout.String()
			}
			t.Errorf("founder payout = %s, want 7000007", got)
		}
	})
}

func TestCalculate_MixedParticipatingAndNonParticipating(t *testing.T) {
	// $50M exit.
	// Series B (participating, uncapped, seniority 2): 2M shares at $2, pref = $4M.
	// Series A (non-participating, seniority 1): 3M shares at $1, pref = $3M.
	// Common: 5M shares.
	//
	// Series A should convert because as-converted ($13.8M) > preference ($3M).
	// After conversion:
	//   Phase 1: Series B pref = $4M. Remaining: $46M.
	//   Phase 2: Common(5M) + Series A(3M) + Series B(2M) = 10M shares share $46M.
	//     Series A: 3M/10M * $46M = $13,800,000
	//     Common:   5M/10M * $46M = $23,000,000
	//     Series B: $4M + 2M/10M * $46M = $4M + $9,200,000 = $13,200,000
	positions := []ShareClassPosition{
		{
			ShareClass: domain.ShareClass{
				Name:                "Series B Preferred",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     true,
				PricePerShare:       ppsPtr("2.00"),
				Seniority:           2,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv_b", StakeholderName: "Investor B", Shares: dec("2000000")},
			},
			TotalShares: dec("2000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:                "Series A Preferred",
				IsPreferred:         true,
				LiquidationMultiple: dec("1"),
				IsParticipating:     false,
				PricePerShare:       ppsPtr("1.00"),
				Seniority:           1,
			},
			Holders: []HolderPosition{
				{StakeholderID: "inv_a", StakeholderName: "Investor A", Shares: dec("3000000")},
			},
			TotalShares: dec("3000000"),
		},
		{
			ShareClass: domain.ShareClass{
				Name:        "Common",
				IsPreferred: false,
			},
			Holders: []HolderPosition{
				{StakeholderID: "f1", StakeholderName: "Founder", Shares: dec("5000000")},
			},
			TotalShares: dec("5000000"),
		},
	}

	result := Calculate(positions, dec("50000000"))

	invA := findPayout(result, "inv_a")
	invB := findPayout(result, "inv_b")
	fdr := findPayout(result, "f1")

	if invA == nil || !invA.Payout.Equal(dec("13800000")) {
		got := "nil"
		if invA != nil {
			got = invA.Payout.String()
		}
		t.Errorf("Series A (converted) payout = %s, want 13800000", got)
	}
	if invB == nil || !invB.Payout.Equal(dec("13200000")) {
		got := "nil"
		if invB != nil {
			got = invB.Payout.String()
		}
		t.Errorf("Series B payout = %s, want 13200000", got)
	}
	if fdr == nil || !fdr.Payout.Equal(dec("23000000")) {
		got := "nil"
		if fdr != nil {
			got = fdr.Payout.String()
		}
		t.Errorf("Founder payout = %s, want 23000000", got)
	}

	if !result.TotalPayout.Equal(dec("50000000")) {
		t.Errorf("total payout = %s, want 50000000", result.TotalPayout)
	}
}

func findPayout(result domain.WaterfallResult, stakeholderID string) *domain.WaterfallPayout {
	for _, p := range result.Payouts {
		if p.StakeholderID == stakeholderID {
			return &p
		}
	}
	return nil
}
