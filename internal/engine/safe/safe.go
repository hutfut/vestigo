package safe

import (
	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

// ConvertPreMoney converts a pre-money SAFE into shares at a priced round.
//
// Pre-money SAFE mechanics: the SAFE converts as if the investment happened
// just before the priced round. The conversion price is the lower of:
//   - cap price: valuation_cap / pre-money shares outstanding
//   - discount price: round PPS * (1 - discount_rate)
//   - round PPS (if neither cap nor discount applies)
//
// Pre-money shares outstanding includes all existing shares but NOT shares
// issued from other converting SAFEs (this is the key difference from post-money).
func ConvertPreMoney(safe domain.SAFENote, round domain.FundingRound, preMoneyShares decimal.Decimal) domain.SAFEConversionResult {
	roundPPS := round.PricePerShare
	effectivePPS := roundPPS
	method := "round_price"

	if safe.ValuationCap != nil && !safe.ValuationCap.IsZero() {
		capPPS := safe.ValuationCap.Div(preMoneyShares)
		if capPPS.LessThan(effectivePPS) {
			effectivePPS = capPPS
			method = "cap"
		}
	}

	if safe.DiscountRate != nil && !safe.DiscountRate.IsZero() {
		discountPPS := roundPPS.Mul(decimal.NewFromInt(1).Sub(*safe.DiscountRate))
		if discountPPS.LessThan(effectivePPS) {
			effectivePPS = discountPPS
			method = "discount"
		}
	}

	shares := safe.InvestmentAmount.Div(effectivePPS).RoundFloor(4)

	return domain.SAFEConversionResult{
		SAFEID:           safe.ID,
		SharesIssued:     shares,
		EffectivePPS:     effectivePPS,
		ConversionMethod: method,
	}
}

// ConvertPostMoney converts a post-money SAFE into shares at a priced round.
//
// Post-money SAFE mechanics: the valuation cap is a post-money cap, meaning
// the SAFE holder's ownership percentage is fixed at:
//   ownership = investment_amount / valuation_cap
//
// The conversion price is the lower of:
//   - cap price: valuation_cap / post-money capitalization (company ownership = cap - investment)
//   - discount price: round PPS * (1 - discount_rate)
//   - round PPS
//
// Post-money capitalization for cap price calculation uses the "company
// capitalization" which equals all shares outstanding as if the SAFE had
// already converted (the cap includes the SAFE itself).
func ConvertPostMoney(safe domain.SAFENote, round domain.FundingRound, preMoneyShares decimal.Decimal) domain.SAFEConversionResult {
	roundPPS := round.PricePerShare
	effectivePPS := roundPPS
	method := "round_price"

	if safe.ValuationCap != nil && !safe.ValuationCap.IsZero() {
		// Post-money cap: ownership% = investment / cap
		// Company capitalization for conversion = cap / round PPS
		// But the standard formula: shares = investment / (cap / capitalization)
		// Where capitalization = the post-money cap divided by round PPS
		// Simplifies to: capPPS = (cap - investment) / preMoneyShares
		// This ensures the SAFE holder gets exactly investment/cap ownership.
		companyCap := safe.ValuationCap.Sub(safe.InvestmentAmount)
		capPPS := companyCap.Div(preMoneyShares)
		if capPPS.LessThan(effectivePPS) {
			effectivePPS = capPPS
			method = "cap"
		}
	}

	if safe.DiscountRate != nil && !safe.DiscountRate.IsZero() {
		discountPPS := roundPPS.Mul(decimal.NewFromInt(1).Sub(*safe.DiscountRate))
		if discountPPS.LessThan(effectivePPS) {
			effectivePPS = discountPPS
			method = "discount"
		}
	}

	shares := safe.InvestmentAmount.Div(effectivePPS).RoundFloor(4)

	return domain.SAFEConversionResult{
		SAFEID:           safe.ID,
		SharesIssued:     shares,
		EffectivePPS:     effectivePPS,
		ConversionMethod: method,
	}
}

// Convert dispatches to the correct conversion logic based on SAFE type.
func Convert(safe domain.SAFENote, round domain.FundingRound, preMoneyShares decimal.Decimal) domain.SAFEConversionResult {
	switch safe.SAFEType {
	case domain.SAFEPostMoney:
		return ConvertPostMoney(safe, round, preMoneyShares)
	default:
		return ConvertPreMoney(safe, round, preMoneyShares)
	}
}
