package vesting

import (
	"testing"
	"time"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func dec(v string) decimal.Decimal {
	d, _ := decimal.NewFromString(v)
	return d
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name            string
		grant           domain.Grant
		asOf            time.Time
		wantVested      string
		wantUnvested    string
		wantPct         string
		wantFullyVested bool
	}{
		{
			name: "before cliff, zero vested",
			grant: domain.Grant{
				ID:        "g1",
				Quantity:  dec("48000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2024, 6, 15),
			wantVested:      "0",
			wantUnvested:    "48000",
			wantPct:         "0",
			wantFullyVested: false,
		},
		{
			name: "exactly at cliff, 25% vested",
			grant: domain.Grant{
				ID:        "g2",
				Quantity:  dec("48000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2025, 1, 1),
			wantVested:      "12000",
			wantUnvested:    "36000",
			wantPct:         "25",
			wantFullyVested: false,
		},
		{
			name: "halfway through 4-year vest",
			grant: domain.Grant{
				ID:        "g3",
				Quantity:  dec("48000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2026, 1, 1),
			wantVested:      "24000",
			wantUnvested:    "24000",
			wantPct:         "50",
			wantFullyVested: false,
		},
		{
			name: "fully vested",
			grant: domain.Grant{
				ID:        "g4",
				Quantity:  dec("48000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2028, 1, 1),
			wantVested:      "48000",
			wantUnvested:    "0",
			wantPct:         "100",
			wantFullyVested: true,
		},
		{
			name: "past fully vested date",
			grant: domain.Grant{
				ID:        "g5",
				Quantity:  dec("48000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2030, 6, 1),
			wantVested:      "48000",
			wantUnvested:    "0",
			wantPct:         "100",
			wantFullyVested: true,
		},
		{
			name: "no vesting schedule, immediate grant",
			grant: domain.Grant{
				ID:              "g6",
				Quantity:        dec("10000"),
				GrantDate:       date(2024, 1, 1),
				VestingSchedule: nil,
			},
			asOf:            date(2024, 1, 1),
			wantVested:      "10000",
			wantUnvested:    "0",
			wantPct:         "100",
			wantFullyVested: true,
		},
		{
			name: "quarterly vesting at 6 months post-cliff",
			grant: domain.Grant{
				ID:        "g7",
				Quantity:  dec("16000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyQuarterly,
				},
			},
			asOf:            date(2025, 7, 1),
			wantVested:      "6000",
			wantUnvested:    "10000",
			wantPct:         "37.50",
			wantFullyVested: false,
		},
		{
			name: "zero cliff, monthly vesting, 3 months in",
			grant: domain.Grant{
				ID:        "g8",
				Quantity:  dec("12000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 0,
					TotalMonths: 12,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2024, 4, 1),
			wantVested:      "3000",
			wantUnvested:    "9000",
			wantPct:         "25",
			wantFullyVested: false,
		},
		{
			name: "annual vesting, 2 years into 4-year schedule",
			grant: domain.Grant{
				ID:        "g9",
				Quantity:  dec("40000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyAnnually,
				},
			},
			asOf:            date(2026, 1, 1),
			wantVested:      "20000",
			wantUnvested:    "20000",
			wantPct:         "50",
			wantFullyVested: false,
		},
		{
			name: "fractional shares round down",
			grant: domain.Grant{
				ID:        "g10",
				Quantity:  dec("10000"),
				GrantDate: date(2024, 1, 1),
				VestingSchedule: &domain.VestingSchedule{
					CliffMonths: 12,
					TotalMonths: 48,
					Frequency:   domain.FrequencyMonthly,
				},
			},
			asOf:            date(2025, 2, 1),
			wantVested:      "2708.3333",
			wantUnvested:    "7291.6667",
			wantPct:         "27.08",
			wantFullyVested: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.grant, tt.asOf)

			if !got.VestedShares.Equal(dec(tt.wantVested)) {
				t.Errorf("VestedShares = %s, want %s", got.VestedShares, tt.wantVested)
			}
			if !got.UnvestedShares.Equal(dec(tt.wantUnvested)) {
				t.Errorf("UnvestedShares = %s, want %s", got.UnvestedShares, tt.wantUnvested)
			}
			if !got.PercentVested.Equal(dec(tt.wantPct)) {
				t.Errorf("PercentVested = %s, want %s", got.PercentVested, tt.wantPct)
			}
			if got.IsFullyVested != tt.wantFullyVested {
				t.Errorf("IsFullyVested = %v, want %v", got.IsFullyVested, tt.wantFullyVested)
			}

			// Invariant: vested + unvested = total
			if !got.VestedShares.Add(got.UnvestedShares).Equal(got.TotalShares) {
				t.Errorf("invariant violation: vested(%s) + unvested(%s) != total(%s)",
					got.VestedShares, got.UnvestedShares, got.TotalShares)
			}
		})
	}
}

func TestCalculateAccelerated(t *testing.T) {
	grant := domain.Grant{
		ID:        "g1",
		Quantity:  dec("48000"),
		GrantDate: date(2024, 1, 1),
		VestingSchedule: &domain.VestingSchedule{
			CliffMonths:         12,
			TotalMonths:         48,
			Frequency:           domain.FrequencyMonthly,
			AccelerationTrigger: domain.AccelerationSingleTrigger,
		},
	}

	got := CalculateAccelerated(grant, date(2024, 6, 1))

	if !got.VestedShares.Equal(grant.Quantity) {
		t.Errorf("VestedShares = %s, want %s", got.VestedShares, grant.Quantity)
	}
	if !got.UnvestedShares.Equal(decimal.Zero) {
		t.Errorf("UnvestedShares = %s, want 0", got.UnvestedShares)
	}
	if !got.IsFullyVested {
		t.Error("expected IsFullyVested = true")
	}
}
