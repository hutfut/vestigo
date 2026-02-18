package waterfall

import (
	"sort"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/shopspring/decimal"
)

// ShareClassPosition represents a share class and its holders for waterfall purposes.
type ShareClassPosition struct {
	ShareClass  domain.ShareClass
	Holders     []HolderPosition
	TotalShares decimal.Decimal
}

// HolderPosition represents one stakeholder's position within a share class.
type HolderPosition struct {
	StakeholderID   string
	StakeholderName string
	Shares          decimal.Decimal
}

// Calculate computes the liquidation waterfall for a given exit valuation.
//
// The algorithm proceeds in seniority order (highest first):
//  1. Each preferred class receives its liquidation preference (multiple * invested).
//  2. If participating, preferred also shares in remaining proceeds pro-rata
//     with common, subject to the participation cap.
//  3. Non-participating preferred compares its liquidation preference to its
//     as-converted common payout and takes whichever is higher. When a class
//     converts, the waterfall is recalculated with that class in the common pool.
//  4. Common shares receive whatever remains after all preferences are satisfied.
func Calculate(positions []ShareClassPosition, exitValuation decimal.Decimal) domain.WaterfallResult {
	result := domain.WaterfallResult{
		ExitValuation: exitValuation,
		Payouts:       []domain.WaterfallPayout{},
	}

	if exitValuation.LessThanOrEqual(decimal.Zero) {
		return result
	}

	converting := resolveConversions(positions, exitValuation)
	payoutMap := distribute(positions, exitValuation, converting)

	totalPayout := decimal.Zero
	for _, pos := range positions {
		for _, h := range pos.Holders {
			payout := payoutMap[h.StakeholderID]
			if payout.GreaterThan(decimal.Zero) {
				perShare := decimal.Zero
				if h.Shares.GreaterThan(decimal.Zero) {
					perShare = payout.Div(h.Shares).RoundFloor(4)
				}
				result.Payouts = append(result.Payouts, domain.WaterfallPayout{
					StakeholderID:   h.StakeholderID,
					StakeholderName: h.StakeholderName,
					ShareClassName:  pos.ShareClass.Name,
					Shares:          h.Shares,
					Payout:          payout,
					PayoutPerShare:  perShare,
				})
				totalPayout = totalPayout.Add(payout)
			}
		}
	}

	result.TotalPayout = totalPayout
	return result
}

// resolveConversions determines which non-participating preferred classes should
// convert to common. For each such class it compares the preference payout to the
// as-converted payout and picks whichever is higher. Because one class converting
// changes the pool for others, this iterates until all decisions stabilize.
func resolveConversions(positions []ShareClassPosition, exitValuation decimal.Decimal) map[int]bool {
	var npIndices []int
	for i, p := range positions {
		if p.ShareClass.IsPreferred && !p.ShareClass.IsParticipating {
			npIndices = append(npIndices, i)
		}
	}
	if len(npIndices) == 0 {
		return nil
	}

	converting := make(map[int]bool)

	for iter := 0; iter < 50; iter++ {
		changed := false
		for _, idx := range npIndices {
			withPref := cloneIntSet(converting)
			delete(withPref, idx)
			prefPayouts := distribute(positions, exitValuation, withPref)
			prefTotal := holderPayoutSum(positions[idx].Holders, prefPayouts)

			withConv := cloneIntSet(converting)
			withConv[idx] = true
			convPayouts := distribute(positions, exitValuation, withConv)
			convTotal := holderPayoutSum(positions[idx].Holders, convPayouts)

			shouldConvert := convTotal.GreaterThan(prefTotal)
			if shouldConvert != converting[idx] {
				converting[idx] = shouldConvert
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	for k, v := range converting {
		if !v {
			delete(converting, k)
		}
	}
	return converting
}

// distribute runs the waterfall payout with the given conversion decisions.
// Classes whose index appears in converting forfeit their liquidation preference
// and are treated as common for pro-rata distribution.
func distribute(positions []ShareClassPosition, exitValuation decimal.Decimal, converting map[int]bool) map[string]decimal.Decimal {
	type indexedPos struct {
		pos ShareClassPosition
	}

	var preferred []ShareClassPosition
	var commonPool []ShareClassPosition

	for i, p := range positions {
		if p.ShareClass.IsPreferred && !converting[i] {
			preferred = append(preferred, p)
		} else {
			commonPool = append(commonPool, p)
		}
	}

	sort.Slice(preferred, func(i, j int) bool {
		return preferred[i].ShareClass.Seniority > preferred[j].ShareClass.Seniority
	})

	remaining := exitValuation
	payoutMap := make(map[string]decimal.Decimal)
	var participatingClasses []ShareClassPosition

	// Phase 1: Pay liquidation preferences to non-converting preferred shareholders.
	for _, pref := range preferred {
		if remaining.LessThanOrEqual(decimal.Zero) {
			break
		}

		investedAmount := pref.TotalShares.Mul(derefOrOne(pref.ShareClass.PricePerShare))
		preference := investedAmount.Mul(pref.ShareClass.LiquidationMultiple)

		paid := decimal.Min(preference, remaining)
		remaining = remaining.Sub(paid)

		for _, h := range pref.Holders {
			holderFraction := h.Shares.Div(pref.TotalShares)
			holderPayout := paid.Mul(holderFraction).RoundFloor(4)
			payoutMap[h.StakeholderID] = payoutMap[h.StakeholderID].Add(holderPayout)
		}

		if pref.ShareClass.IsParticipating {
			participatingClasses = append(participatingClasses, pref)
		}
	}

	// Phase 2: Distribute remaining proceeds among common pool + participating preferred.
	if remaining.GreaterThan(decimal.Zero) {
		totalParticipating := decimal.Zero

		for _, c := range commonPool {
			totalParticipating = totalParticipating.Add(c.TotalShares)
		}
		for _, p := range participatingClasses {
			totalParticipating = totalParticipating.Add(p.TotalShares)
		}

		if totalParticipating.GreaterThan(decimal.Zero) {
			distributeProRata(commonPool, remaining, totalParticipating, payoutMap, nil)
			distributeProRata(participatingClasses, remaining, totalParticipating, payoutMap, capFor)
		}
	}

	return payoutMap
}

type capFunc func(ShareClassPosition) *decimal.Decimal

func capFor(pos ShareClassPosition) *decimal.Decimal {
	if !pos.ShareClass.IsParticipating {
		return nil
	}
	if pos.ShareClass.ParticipationCap == nil {
		return nil
	}
	investedAmount := pos.TotalShares.Mul(derefOrOne(pos.ShareClass.PricePerShare))
	totalCap := investedAmount.Mul(*pos.ShareClass.ParticipationCap)
	return &totalCap
}

func distributeProRata(classes []ShareClassPosition, pool, totalShares decimal.Decimal, payoutMap map[string]decimal.Decimal, capFn capFunc) {
	for _, cls := range classes {
		classFraction := cls.TotalShares.Div(totalShares)
		classPool := pool.Mul(classFraction)

		if capFn != nil {
			if cap := capFn(cls); cap != nil {
				alreadyPaid := decimal.Zero
				for _, h := range cls.Holders {
					alreadyPaid = alreadyPaid.Add(payoutMap[h.StakeholderID])
				}
				maxAdditional := cap.Sub(alreadyPaid)
				if maxAdditional.LessThanOrEqual(decimal.Zero) {
					continue
				}
				classPool = decimal.Min(classPool, maxAdditional)
			}
		}

		for _, h := range cls.Holders {
			holderFraction := h.Shares.Div(cls.TotalShares)
			holderPayout := classPool.Mul(holderFraction).RoundFloor(4)
			payoutMap[h.StakeholderID] = payoutMap[h.StakeholderID].Add(holderPayout)
		}
	}
}

func derefOrOne(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.NewFromInt(1)
	}
	return *d
}

func cloneIntSet(m map[int]bool) map[int]bool {
	c := make(map[int]bool, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func holderPayoutSum(holders []HolderPosition, payouts map[string]decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, h := range holders {
		total = total.Add(payouts[h.StakeholderID])
	}
	return total
}
