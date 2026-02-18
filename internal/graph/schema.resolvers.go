package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/hutfut/vestigo/internal/engine/dilution"
	safeengine "github.com/hutfut/vestigo/internal/engine/safe"
	vestingengine "github.com/hutfut/vestigo/internal/engine/vesting"
	waterfallengine "github.com/hutfut/vestigo/internal/engine/waterfall"
	"github.com/hutfut/vestigo/internal/graph/convert"
	"github.com/hutfut/vestigo/internal/graph/model"
	"github.com/shopspring/decimal"
)

// ─── Mutations ────────────────────────────────────────────────────────────────

func (r *mutationResolver) CreateCompany(ctx context.Context, input model.CreateCompanyInput) (*model.Company, error) {
	c := &domain.Company{Name: input.Name}
	if err := r.Companies.Create(ctx, c); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "company", c.ID, "create", nil, c)
	return convert.ToGQLCompany(c), nil
}

func (r *mutationResolver) AddStakeholder(ctx context.Context, input model.AddStakeholderInput) (*model.Stakeholder, error) {
	sh := &domain.Stakeholder{
		CompanyID: input.CompanyID,
		Name:      input.Name,
		Email:     input.Email,
		Role:      convert.GQLRoleToDomain(input.Role),
	}
	if err := r.Stakeholders.Create(ctx, sh); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "stakeholder", sh.ID, "create", nil, sh)
	return convert.ToGQLStakeholder(sh), nil
}

func (r *mutationResolver) CreateShareClass(ctx context.Context, input model.CreateShareClassInput) (*model.ShareClass, error) {
	sc := &domain.ShareClass{
		CompanyID:           input.CompanyID,
		Name:                input.Name,
		IsPreferred:         input.IsPreferred,
		LiquidationMultiple: convert.DecOrDefault(input.LiquidationMultiple, decimal.NewFromInt(1)),
		IsParticipating:     convert.BoolOrDefault(input.IsParticipating, false),
		ParticipationCap:    convert.GQLDecToDecPtr(input.ParticipationCap),
		PricePerShare:       convert.GQLDecToDecPtr(input.PricePerShare),
		Seniority:           convert.IntOrDefault(input.Seniority, 0),
		AuthorizedShares:    decimal.Decimal(input.AuthorizedShares),
	}
	if err := r.ShareClasses.Create(ctx, sc); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "share_class", sc.ID, "create", nil, sc)
	return convert.ToGQLShareClass(sc), nil
}

func (r *mutationResolver) CreateVestingSchedule(ctx context.Context, input model.CreateVestingScheduleInput) (*model.VestingSchedule, error) {
	accel := domain.AccelerationNone
	if input.AccelerationTrigger != nil {
		accel = convert.GQLAccelToDomain(*input.AccelerationTrigger)
	}
	vs := &domain.VestingSchedule{
		CliffMonths:         input.CliffMonths,
		TotalMonths:         input.TotalMonths,
		Frequency:           convert.GQLFreqToDomain(input.Frequency),
		AccelerationTrigger: accel,
	}
	if err := r.VestingSchedules.Create(ctx, vs); err != nil {
		return nil, err
	}
	return convert.ToGQLVestingSchedule(vs), nil
}

func (r *mutationResolver) IssueGrant(ctx context.Context, input model.IssueGrantInput) (*model.Grant, error) {
	g := &domain.Grant{
		CompanyID:         input.CompanyID,
		StakeholderID:     input.StakeholderID,
		ShareClassID:      input.ShareClassID,
		VestingScheduleID: input.VestingScheduleID,
		Quantity:          decimal.Decimal(input.Quantity),
		GrantDate:         time.Time(input.GrantDate),
		ExercisePrice:     convert.DecOrDefault(input.ExercisePrice, decimal.Zero),
		Notes:             input.Notes,
	}
	if err := r.Grants.Create(ctx, g); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "grant", g.ID, "create", nil, g)
	return convert.ToGQLGrant(g), nil
}

func (r *mutationResolver) RecordFundingRound(ctx context.Context, input model.RecordFundingRoundInput) (*model.FundingRound, error) {
	fr := &domain.FundingRound{
		CompanyID:     input.CompanyID,
		Name:          input.Name,
		PreMoneyVal:   decimal.Decimal(input.PreMoneyValuation),
		AmountRaised:  decimal.Decimal(input.AmountRaised),
		PricePerShare: decimal.Decimal(input.PricePerShare),
		ShareClassID:  input.ShareClassID,
		RoundDate:     time.Time(input.RoundDate),
	}
	if err := r.FundingRounds.Create(ctx, fr); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "funding_round", fr.ID, "create", nil, fr)
	return convert.ToGQLFundingRound(fr), nil
}

func (r *mutationResolver) IssueSafe(ctx context.Context, input model.IssueSAFEInput) (*model.SAFENote, error) {
	sn := &domain.SAFENote{
		CompanyID:        input.CompanyID,
		StakeholderID:    input.StakeholderID,
		InvestmentAmount: decimal.Decimal(input.InvestmentAmount),
		ValuationCap:     convert.GQLDecToDecPtr(input.ValuationCap),
		DiscountRate:     convert.GQLDecToDecPtr(input.DiscountRate),
		SAFEType:         convert.GQLSAFETypeToDomain(input.SafeType),
		IssueDate:        time.Time(input.IssueDate),
	}
	if err := r.SAFENotes.Create(ctx, sn); err != nil {
		return nil, err
	}
	r.Audit.Record(ctx, "safe_note", sn.ID, "create", nil, sn)
	return convert.ToGQLSAFENote(sn), nil
}

func (r *mutationResolver) ConvertSafe(ctx context.Context, safeID string, roundID string) (*model.SAFEConversionResult, error) {
	sn, err := r.SAFENotes.GetByID(ctx, safeID)
	if err != nil {
		return nil, err
	}
	if sn.IsConverted {
		return nil, &domain.ErrConflict{Message: fmt.Sprintf("SAFE %s is already converted", safeID)}
	}

	round, err := r.FundingRounds.GetByID(ctx, roundID)
	if err != nil {
		return nil, err
	}

	grants, err := r.Grants.ListByCompany(ctx, sn.CompanyID)
	if err != nil {
		return nil, err
	}

	preMoneyShares := decimal.Zero
	for _, g := range grants {
		preMoneyShares = preMoneyShares.Add(g.Quantity)
	}

	result := safeengine.Convert(*sn, *round, preMoneyShares)

	if err := r.SAFENotes.MarkConverted(ctx, safeID, roundID); err != nil {
		return nil, err
	}

	r.Audit.Record(ctx, "safe_note", safeID, "convert", sn, map[string]interface{}{
		"round_id":      roundID,
		"shares_issued": result.SharesIssued.String(),
		"effective_pps": result.EffectivePPS.String(),
		"method":        result.ConversionMethod,
	})

	return &model.SAFEConversionResult{
		SafeID:           result.SAFEID,
		SharesIssued:     model.Decimal(result.SharesIssued),
		EffectivePps:     model.Decimal(result.EffectivePPS),
		ConversionMethod: result.ConversionMethod,
	}, nil
}

// ─── Queries ──────────────────────────────────────────────────────────────────

func (r *queryResolver) Company(ctx context.Context, id string) (*model.Company, error) {
	c, err := r.Companies.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	mc := convert.ToGQLCompany(c)

	stakeholders, err := r.Stakeholders.ListByCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range stakeholders {
		mc.Stakeholders = append(mc.Stakeholders, convert.ToGQLStakeholder(&stakeholders[i]))
	}

	classes, err := r.ShareClasses.ListByCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range classes {
		mc.ShareClasses = append(mc.ShareClasses, convert.ToGQLShareClass(&classes[i]))
	}

	grants, err := r.Grants.ListByCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range grants {
		mc.Grants = append(mc.Grants, convert.ToGQLGrant(&grants[i]))
	}

	rounds, err := r.FundingRounds.ListByCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range rounds {
		mc.FundingRounds = append(mc.FundingRounds, convert.ToGQLFundingRound(&rounds[i]))
	}

	safes, err := r.SAFENotes.ListByCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range safes {
		mc.SafeNotes = append(mc.SafeNotes, convert.ToGQLSAFENote(&safes[i]))
	}

	return mc, nil
}

func (r *queryResolver) Stakeholder(ctx context.Context, id string) (*model.Stakeholder, error) {
	sh, err := r.Stakeholders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	msh := convert.ToGQLStakeholder(sh)

	grants, err := r.Grants.ListByStakeholder(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range grants {
		msh.Grants = append(msh.Grants, convert.ToGQLGrant(&grants[i]))
	}

	return msh, nil
}

func (r *queryResolver) VestingStatus(ctx context.Context, grantID string, asOfDate model.Date) (*model.VestingStatus, error) {
	g, err := r.Grants.GetByID(ctx, grantID)
	if err != nil {
		return nil, err
	}

	if g.VestingScheduleID != nil {
		vs, err := r.VestingSchedules.GetByID(ctx, *g.VestingScheduleID)
		if err != nil {
			return nil, err
		}
		g.VestingSchedule = vs
	}

	status := vestingengine.Calculate(*g, time.Time(asOfDate))
	return convert.ToGQLVestingStatus(&status), nil
}

func (r *queryResolver) CapTable(ctx context.Context, companyID string) (*model.CapTableSnapshot, error) {
	grants, err := r.Grants.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	shIDs, scIDs := collectGrantIDs(grants)
	shMap, err := r.Stakeholders.GetByIDs(ctx, shIDs)
	if err != nil {
		return nil, err
	}
	scMap, err := r.ShareClasses.GetByIDs(ctx, scIDs)
	if err != nil {
		return nil, err
	}

	type key struct{ shID, scID string }
	type entry struct {
		shName string
		scName string
		shares decimal.Decimal
	}
	agg := map[key]*entry{}
	totalShares := decimal.Zero

	for _, g := range grants {
		k := key{g.StakeholderID, g.ShareClassID}
		if _, ok := agg[k]; !ok {
			sh, sc := shMap[g.StakeholderID], scMap[g.ShareClassID]
			if sh == nil || sc == nil {
				return nil, fmt.Errorf("missing stakeholder %s or share class %s", g.StakeholderID, g.ShareClassID)
			}
			agg[k] = &entry{shName: sh.Name, scName: sc.Name, shares: decimal.Zero}
		}
		agg[k].shares = agg[k].shares.Add(g.Quantity)
		totalShares = totalShares.Add(g.Quantity)
	}

	hundred := decimal.NewFromInt(100)
	entries := make([]*model.CapTableEntry, 0, len(agg))
	for k, e := range agg {
		pct := decimal.Zero
		if totalShares.GreaterThan(decimal.Zero) {
			pct = e.shares.Div(totalShares).Mul(hundred).RoundFloor(4)
		}
		shID := k.shID
		entries = append(entries, &model.CapTableEntry{
			StakeholderID:   &shID,
			StakeholderName: e.shName,
			ShareClassName:  e.scName,
			Shares:          model.Decimal(e.shares),
			OwnershipPct:    model.Decimal(pct),
		})
	}

	cID := companyID
	return &model.CapTableSnapshot{
		CompanyID:   &cID,
		TotalShares: model.Decimal(totalShares),
		Entries:     entries,
	}, nil
}

func (r *queryResolver) ModelDilution(ctx context.Context, input model.DilutionModelInput) (*model.DilutionResult, error) {
	grants, err := r.Grants.ListByCompany(ctx, input.CompanyID)
	if err != nil {
		return nil, err
	}

	shIDs, scIDs := collectGrantIDs(grants)
	shMap, err := r.Stakeholders.GetByIDs(ctx, shIDs)
	if err != nil {
		return nil, err
	}
	scMap, err := r.ShareClasses.GetByIDs(ctx, scIDs)
	if err != nil {
		return nil, err
	}

	type key struct{ shID, scID string }
	type holder struct {
		shID, shName, scName string
		shares               decimal.Decimal
	}
	agg := map[key]*holder{}

	for _, g := range grants {
		k := key{g.StakeholderID, g.ShareClassID}
		if _, ok := agg[k]; !ok {
			sh, sc := shMap[g.StakeholderID], scMap[g.ShareClassID]
			if sh == nil || sc == nil {
				return nil, fmt.Errorf("missing stakeholder %s or share class %s", g.StakeholderID, g.ShareClassID)
			}
			agg[k] = &holder{shID: sh.ID, shName: sh.Name, scName: sc.Name, shares: decimal.Zero}
		}
		agg[k].shares = agg[k].shares.Add(g.Quantity)
	}

	existing := make([]dilution.StakeholderShares, 0, len(agg))
	for _, h := range agg {
		existing = append(existing, dilution.StakeholderShares{
			StakeholderID:   h.shID,
			StakeholderName: h.shName,
			ShareClassName:  h.scName,
			Shares:          h.shares,
		})
	}

	result := dilution.Model(existing, dilution.RoundInput{
		RoundName:     input.RoundName,
		PreMoneyVal:   decimal.Decimal(input.PreMoneyValuation),
		AmountRaised:  decimal.Decimal(input.AmountRaised),
		NewShareClass: input.NewShareClass,
		InvestorName:  input.InvestorName,
	})

	return convert.ToGQLDilutionResult(&result), nil
}

func (r *queryResolver) Waterfall(ctx context.Context, companyID string, exitValuation model.Decimal) (*model.WaterfallResult, error) {
	classes, err := r.ShareClasses.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	grants, err := r.Grants.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Build class-to-holders map
	classGrants := map[string][]domain.Grant{}
	for _, g := range grants {
		classGrants[g.ShareClassID] = append(classGrants[g.ShareClassID], g)
	}

	shIDs, _ := collectGrantIDs(grants)
	shMap, err := r.Stakeholders.GetByIDs(ctx, shIDs)
	if err != nil {
		return nil, err
	}

	positions := make([]waterfallengine.ShareClassPosition, 0, len(classes))
	for _, sc := range classes {
		gg := classGrants[sc.ID]
		if len(gg) == 0 {
			continue
		}

		totalShares := decimal.Zero
		holders := make([]waterfallengine.HolderPosition, 0, len(gg))

		for _, g := range gg {
			sh := shMap[g.StakeholderID]
			if sh == nil {
				return nil, fmt.Errorf("missing stakeholder %s", g.StakeholderID)
			}
			holders = append(holders, waterfallengine.HolderPosition{
				StakeholderID:   sh.ID,
				StakeholderName: sh.Name,
				Shares:          g.Quantity,
			})
			totalShares = totalShares.Add(g.Quantity)
		}

		positions = append(positions, waterfallengine.ShareClassPosition{
			ShareClass:  sc,
			Holders:     holders,
			TotalShares: totalShares,
		})
	}

	result := waterfallengine.Calculate(positions, decimal.Decimal(exitValuation))

	payouts := make([]*model.WaterfallPayout, len(result.Payouts))
	for i, p := range result.Payouts {
		payouts[i] = &model.WaterfallPayout{
			StakeholderID:   p.StakeholderID,
			StakeholderName: p.StakeholderName,
			ShareClassName:  p.ShareClassName,
			Shares:          model.Decimal(p.Shares),
			Payout:          model.Decimal(p.Payout),
			PayoutPerShare:  model.Decimal(p.PayoutPerShare),
		}
	}

	return &model.WaterfallResult{
		ExitValuation: model.Decimal(result.ExitValuation),
		TotalPayout:   model.Decimal(result.TotalPayout),
		Payouts:       payouts,
	}, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// collectGrantIDs returns deduplicated stakeholder and share-class IDs from a
// slice of grants, suitable for batch-fetching.
func collectGrantIDs(grants []domain.Grant) (stakeholderIDs, shareClassIDs []string) {
	shSeen := make(map[string]struct{}, len(grants))
	scSeen := make(map[string]struct{}, len(grants))
	for _, g := range grants {
		if _, ok := shSeen[g.StakeholderID]; !ok {
			shSeen[g.StakeholderID] = struct{}{}
			stakeholderIDs = append(stakeholderIDs, g.StakeholderID)
		}
		if _, ok := scSeen[g.ShareClassID]; !ok {
			scSeen[g.ShareClassID] = struct{}{}
			shareClassIDs = append(shareClassIDs, g.ShareClassID)
		}
	}
	return
}
