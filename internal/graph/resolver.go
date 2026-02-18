package graph

import (
	"database/sql"

	"github.com/hutfut/vestigo/internal/audit"
	"github.com/hutfut/vestigo/internal/store"
)

type Resolver struct {
	DB               *sql.DB
	Companies        *store.CompanyStore
	Stakeholders     *store.StakeholderStore
	ShareClasses     *store.ShareClassStore
	VestingSchedules *store.VestingScheduleStore
	Grants           *store.GrantStore
	FundingRounds    *store.FundingRoundStore
	SAFENotes        *store.SAFENoteStore
	Audit            *audit.Logger
}
