package dilution

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(v string) decimal.Decimal {
	d, _ := decimal.NewFromString(v)
	return d
}

func TestModel(t *testing.T) {
	tests := []struct {
		name            string
		existing        []StakeholderShares
		input           RoundInput
		wantNewShares   string
		wantInvestorPct string
		wantTotalPost   string
	}{
		{
			name: "standard Series A: $5M at $15M pre on 10M shares",
			existing: []StakeholderShares{
				{StakeholderID: "f1", StakeholderName: "Alice", ShareClassName: "Common", Shares: dec("7000000")},
				{StakeholderID: "f2", StakeholderName: "Bob", ShareClassName: "Common", Shares: dec("3000000")},
			},
			input: RoundInput{
				RoundName:     "Series A",
				PreMoneyVal:   dec("15000000"),
				AmountRaised:  dec("5000000"),
				NewShareClass: "Preferred A",
				InvestorName:  "Acme VC",
			},
			// PPS = 15M / 10M = 1.50
			// New shares = 5M / 1.50 = 3,333,333.3333
			// Investor % = 3,333,333.3333 / 13,333,333.3333 = 25%
			wantNewShares:   "3333333.3333",
			wantInvestorPct: "24.9999",
			wantTotalPost:   "13333333.3333",
		},
		{
			name: "seed round: $1M at $4M pre on 10M shares (20% dilution)",
			existing: []StakeholderShares{
				{StakeholderID: "f1", StakeholderName: "Founder", ShareClassName: "Common", Shares: dec("10000000")},
			},
			input: RoundInput{
				RoundName:     "Seed",
				PreMoneyVal:   dec("4000000"),
				AmountRaised:  dec("1000000"),
				NewShareClass: "Common",
				InvestorName:  "Angel",
			},
			// PPS = 4M / 10M = 0.40
			// New shares = 1M / 0.40 = 2,500,000
			// Investor % = 2.5M / 12.5M = 20%
			wantNewShares:   "2500000",
			wantInvestorPct: "20",
			wantTotalPost:   "12500000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Model(tt.existing, tt.input)

			if !result.NewInvestor.Shares.Equal(dec(tt.wantNewShares)) {
				t.Errorf("NewInvestor.Shares = %s, want %s", result.NewInvestor.Shares, tt.wantNewShares)
			}
			if !result.NewInvestor.OwnershipPct.Equal(dec(tt.wantInvestorPct)) {
				t.Errorf("NewInvestor.OwnershipPct = %s, want %s", result.NewInvestor.OwnershipPct, tt.wantInvestorPct)
			}
			if !result.PostRound.TotalShares.Equal(dec(tt.wantTotalPost)) {
				t.Errorf("TotalShares = %s, want %s", result.PostRound.TotalShares, tt.wantTotalPost)
			}

			// Invariant: all post-round ownership percentages should sum to ~100%
			totalPct := decimal.Zero
			for _, e := range result.PostRound.Entries {
				totalPct = totalPct.Add(e.OwnershipPct)
			}
			hundred := dec("100")
			diff := totalPct.Sub(hundred).Abs()
			if diff.GreaterThan(dec("0.01")) {
				t.Errorf("post-round ownership sums to %s, want ~100", totalPct)
			}

			// Invariant: pre-round total equals sum of existing shares
			preTotal := decimal.Zero
			for _, e := range result.PreRound.Entries {
				preTotal = preTotal.Add(e.Shares)
			}
			if !preTotal.Equal(result.PreRound.TotalShares) {
				t.Errorf("pre-round shares sum %s != TotalShares %s", preTotal, result.PreRound.TotalShares)
			}
		})
	}
}

func TestModel_DilutionImpact(t *testing.T) {
	existing := []StakeholderShares{
		{StakeholderID: "f1", StakeholderName: "Founder", ShareClassName: "Common", Shares: dec("10000000")},
	}

	result := Model(existing, RoundInput{
		RoundName:     "Series A",
		PreMoneyVal:   dec("10000000"),
		AmountRaised:  dec("5000000"),
		NewShareClass: "Preferred A",
		InvestorName:  "VC Fund",
	})

	founderPre := result.PreRound.Entries[0].OwnershipPct
	founderPost := result.PostRound.Entries[0].OwnershipPct

	if !founderPre.Equal(dec("100")) {
		t.Errorf("founder pre-round ownership = %s, want 100", founderPre)
	}

	// $5M on $10M pre = 33.33% to investor, founder diluted to 66.66%
	expectedPostPct := dec("66.6666")
	if !founderPost.Equal(expectedPostPct) {
		t.Errorf("founder post-round ownership = %s, want %s", founderPost, expectedPostPct)
	}
}
