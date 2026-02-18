package safe

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

func TestConvertPreMoney(t *testing.T) {
	tests := []struct {
		name           string
		safe           domain.SAFENote
		round          domain.FundingRound
		preMoneyShares decimal.Decimal
		wantShares     string
		wantPPS        string
		wantMethod     string
	}{
		{
			name: "cap is binding (cap PPS < round PPS)",
			safe: domain.SAFENote{
				ID:               "s1",
				InvestmentAmount: dec("500000"),
				ValuationCap:     decPtr("5000000"),
				DiscountRate:     decPtr("0.20"),
				SAFEType:         domain.SAFEPreMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("1.50"),
			},
			preMoneyShares: dec("5000000"),
			wantShares:     "500000",
			wantPPS:        "1",
			wantMethod:     "cap",
		},
		{
			name: "discount is binding (discount PPS < cap PPS)",
			safe: domain.SAFENote{
				ID:               "s2",
				InvestmentAmount: dec("500000"),
				ValuationCap:     decPtr("10000000"),
				DiscountRate:     decPtr("0.20"),
				SAFEType:         domain.SAFEPreMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("1.50"),
			},
			preMoneyShares: dec("5000000"),
			wantShares:     "416666.6666",
			wantPPS:        "1.2",
			wantMethod:     "discount",
		},
		{
			name: "round price is binding (no cap, no discount)",
			safe: domain.SAFENote{
				ID:               "s3",
				InvestmentAmount: dec("500000"),
				SAFEType:         domain.SAFEPreMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("1.00"),
			},
			preMoneyShares: dec("10000000"),
			wantShares:     "500000",
			wantPPS:        "1",
			wantMethod:     "round_price",
		},
		{
			name: "cap only, no discount",
			safe: domain.SAFENote{
				ID:               "s4",
				InvestmentAmount: dec("250000"),
				ValuationCap:     decPtr("4000000"),
				SAFEType:         domain.SAFEPreMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("2.00"),
			},
			preMoneyShares: dec("5000000"),
			wantShares:     "312500",
			wantPPS:        "0.8",
			wantMethod:     "cap",
		},
		{
			name: "discount only, no cap",
			safe: domain.SAFENote{
				ID:               "s5",
				InvestmentAmount: dec("100000"),
				DiscountRate:     decPtr("0.15"),
				SAFEType:         domain.SAFEPreMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("2.00"),
			},
			preMoneyShares: dec("5000000"),
			wantShares:     "58823.5294",
			wantPPS:        "1.7",
			wantMethod:     "discount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertPreMoney(tt.safe, tt.round, tt.preMoneyShares)

			if !got.SharesIssued.Equal(dec(tt.wantShares)) {
				t.Errorf("SharesIssued = %s, want %s", got.SharesIssued, tt.wantShares)
			}
			if !got.EffectivePPS.Equal(dec(tt.wantPPS)) {
				t.Errorf("EffectivePPS = %s, want %s", got.EffectivePPS, tt.wantPPS)
			}
			if got.ConversionMethod != tt.wantMethod {
				t.Errorf("ConversionMethod = %s, want %s", got.ConversionMethod, tt.wantMethod)
			}
		})
	}
}

func TestConvertPostMoney(t *testing.T) {
	tests := []struct {
		name           string
		safe           domain.SAFENote
		round          domain.FundingRound
		preMoneyShares decimal.Decimal
		wantShares     string
		wantPPS        string
		wantMethod     string
	}{
		{
			name: "post-money cap is binding",
			safe: domain.SAFENote{
				ID:               "s1",
				InvestmentAmount: dec("500000"),
				ValuationCap:     decPtr("5000000"),
				SAFEType:         domain.SAFEPostMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("2.00"),
			},
			preMoneyShares: dec("5000000"),
			// capPPS = (5M - 500K) / 5M = 0.9
			// shares = 500K / 0.9 = 555555.5555
			wantShares: "555555.5555",
			wantPPS:    "0.9",
			wantMethod: "cap",
		},
		{
			name: "post-money discount is binding over cap",
			safe: domain.SAFENote{
				ID:               "s2",
				InvestmentAmount: dec("500000"),
				ValuationCap:     decPtr("20000000"),
				DiscountRate:     decPtr("0.25"),
				SAFEType:         domain.SAFEPostMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("2.00"),
			},
			preMoneyShares: dec("5000000"),
			// capPPS = (20M - 500K) / 5M = 3.9 (higher than round PPS, not binding)
			// discountPPS = 2.00 * 0.75 = 1.50
			// shares = 500K / 1.50 = 333333.3333
			wantShares: "333333.3333",
			wantPPS:    "1.5",
			wantMethod: "discount",
		},
		{
			name: "round price wins when cap and discount are both higher",
			safe: domain.SAFENote{
				ID:               "s3",
				InvestmentAmount: dec("100000"),
				ValuationCap:     decPtr("100000000"),
				DiscountRate:     decPtr("0.05"),
				SAFEType:         domain.SAFEPostMoney,
			},
			round: domain.FundingRound{
				PricePerShare: dec("0.50"),
			},
			preMoneyShares: dec("10000000"),
			// capPPS = (100M - 100K) / 10M = 9.99 (way higher)
			// discountPPS = 0.50 * 0.95 = 0.475
			// round PPS = 0.50
			// discount PPS 0.475 < 0.50 so discount wins
			wantShares: "210526.3157",
			wantPPS:    "0.475",
			wantMethod: "discount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertPostMoney(tt.safe, tt.round, tt.preMoneyShares)

			if !got.SharesIssued.Equal(dec(tt.wantShares)) {
				t.Errorf("SharesIssued = %s, want %s", got.SharesIssued, tt.wantShares)
			}
			if !got.EffectivePPS.Equal(dec(tt.wantPPS)) {
				t.Errorf("EffectivePPS = %s, want %s", got.EffectivePPS, tt.wantPPS)
			}
			if got.ConversionMethod != tt.wantMethod {
				t.Errorf("ConversionMethod = %s, want %s", got.ConversionMethod, tt.wantMethod)
			}
		})
	}
}

func TestConvert_DispatchesByType(t *testing.T) {
	preMoneyShares := dec("10000000")
	round := domain.FundingRound{PricePerShare: dec("1.00")}

	preMoney := domain.SAFENote{
		ID:               "pre",
		InvestmentAmount: dec("100000"),
		ValuationCap:     decPtr("5000000"),
		SAFEType:         domain.SAFEPreMoney,
	}
	postMoney := domain.SAFENote{
		ID:               "post",
		InvestmentAmount: dec("100000"),
		ValuationCap:     decPtr("5000000"),
		SAFEType:         domain.SAFEPostMoney,
	}

	preResult := Convert(preMoney, round, preMoneyShares)
	postResult := Convert(postMoney, round, preMoneyShares)

	// Pre-money and post-money should yield different share counts for the
	// same inputs because the cap calculation differs.
	if preResult.SharesIssued.Equal(postResult.SharesIssued) {
		t.Error("expected different share counts for pre-money vs post-money SAFE")
	}
}
