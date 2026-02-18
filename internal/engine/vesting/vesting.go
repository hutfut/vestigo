package vesting

import (
	"time"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

// Calculate computes the vesting status for a grant at a given point in time.
// The grant must have a VestingSchedule attached. If it does not, the entire
// grant is treated as fully vested (immediate grant with no schedule).
func Calculate(grant domain.Grant, asOf time.Time) domain.VestingStatus {
	result := domain.VestingStatus{
		GrantID:     grant.ID,
		AsOfDate:    asOf,
		TotalShares: grant.Quantity,
	}

	if grant.VestingSchedule == nil {
		result.VestedShares = grant.Quantity
		result.UnvestedShares = decimal.Zero
		result.PercentVested = decimal.NewFromInt(100)
		result.CliffDate = grant.GrantDate
		result.FullyVestedAt = grant.GrantDate
		result.IsFullyVested = true
		return result
	}

	vs := grant.VestingSchedule
	grantDate := grant.GrantDate
	cliffDate := addMonths(grantDate, vs.CliffMonths)
	fullyVestedAt := addMonths(grantDate, vs.TotalMonths)

	result.CliffDate = cliffDate
	result.FullyVestedAt = fullyVestedAt

	if asOf.Before(cliffDate) {
		result.VestedShares = decimal.Zero
		result.UnvestedShares = grant.Quantity
		result.PercentVested = decimal.Zero
		result.IsFullyVested = false
		return result
	}

	if !asOf.Before(fullyVestedAt) {
		result.VestedShares = grant.Quantity
		result.UnvestedShares = decimal.Zero
		result.PercentVested = decimal.NewFromInt(100)
		result.IsFullyVested = true
		return result
	}

	periodsElapsed := countPeriods(grantDate, asOf, vs.Frequency)
	totalPeriods := countPeriods(grantDate, fullyVestedAt, vs.Frequency)

	if totalPeriods == 0 {
		result.VestedShares = grant.Quantity
		result.UnvestedShares = decimal.Zero
		result.PercentVested = decimal.NewFromInt(100)
		result.IsFullyVested = true
		return result
	}

	vestedFraction := decimal.NewFromInt(int64(periodsElapsed)).
		Div(decimal.NewFromInt(int64(totalPeriods)))

	result.VestedShares = grant.Quantity.Mul(vestedFraction).RoundFloor(4)
	result.UnvestedShares = grant.Quantity.Sub(result.VestedShares)
	result.PercentVested = vestedFraction.Mul(decimal.NewFromInt(100)).RoundFloor(2)
	result.IsFullyVested = false

	return result
}

// CalculateAccelerated returns the vesting status as if single-trigger
// acceleration fires on the given date (100% immediate vesting).
func CalculateAccelerated(grant domain.Grant, triggerDate time.Time) domain.VestingStatus {
	return domain.VestingStatus{
		GrantID:        grant.ID,
		AsOfDate:       triggerDate,
		TotalShares:    grant.Quantity,
		VestedShares:   grant.Quantity,
		UnvestedShares: decimal.Zero,
		PercentVested:  decimal.NewFromInt(100),
		CliffDate:      grant.GrantDate,
		FullyVestedAt:  triggerDate,
		IsFullyVested:  true,
	}
}

func addMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// countPeriods counts how many complete vesting periods have elapsed between
// start and asOf for a given frequency.
func countPeriods(start, asOf time.Time, freq domain.VestingFrequency) int {
	monthsBetween := monthsDiff(start, asOf)

	switch freq {
	case domain.FrequencyMonthly:
		return monthsBetween
	case domain.FrequencyQuarterly:
		return monthsBetween / 3
	case domain.FrequencyAnnually:
		return monthsBetween / 12
	default:
		return monthsBetween
	}
}

// monthsDiff calculates the number of full months between two dates.
func monthsDiff(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	total := years*12 + months

	if to.Day() < from.Day() {
		total--
	}
	if total < 0 {
		return 0
	}
	return total
}
