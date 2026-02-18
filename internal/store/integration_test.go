package store_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hutfut/vestigo/internal/domain"
	"github.com/hutfut/vestigo/internal/store"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("vestigo_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		t.Fatalf("failed to create migrate driver: %v", err)
	}

	migrationsPath := findMigrationsDir(t)
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		candidate := filepath.Join(dir, "migrations")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find migrations directory")
		}
		dir = parent
	}
}

func TestCompanyStore_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	s := store.NewCompanyStore(db)

	c := &domain.Company{Name: "Acme Corp"}
	if err := s.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == "" {
		t.Fatal("expected ID to be set after create")
	}

	got, err := s.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Acme Corp" {
		t.Errorf("Name = %s, want Acme Corp", got.Name)
	}
}

func TestStakeholderStore_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "Test Co"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatalf("create company: %v", err)
	}

	ss := store.NewStakeholderStore(db)
	sh := &domain.Stakeholder{
		CompanyID: company.ID,
		Name:      "Alice",
		Email:     "alice@test.co",
		Role:      domain.RoleFounder,
	}
	if err := ss.Create(ctx, sh); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := ss.ListByCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListByCompany: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 stakeholder, got %d", len(list))
	}
	if list[0].Name != "Alice" {
		t.Errorf("Name = %s, want Alice", list[0].Name)
	}
}

func TestGrantStore_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "GrantCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	scs := store.NewShareClassStore(db)
	sc := &domain.ShareClass{
		CompanyID:           company.ID,
		Name:                "Common",
		IsPreferred:         false,
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(10000000),
	}
	if err := scs.Create(ctx, sc); err != nil {
		t.Fatal(err)
	}

	ss := store.NewStakeholderStore(db)
	sh := &domain.Stakeholder{
		CompanyID: company.ID,
		Name:      "Bob",
		Email:     "bob@grantco.com",
		Role:      domain.RoleEmployee,
	}
	if err := ss.Create(ctx, sh); err != nil {
		t.Fatal(err)
	}

	vs := &domain.VestingSchedule{
		CliffMonths:         12,
		TotalMonths:         48,
		Frequency:           domain.FrequencyMonthly,
		AccelerationTrigger: domain.AccelerationNone,
	}
	vss := store.NewVestingScheduleStore(db)
	if err := vss.Create(ctx, vs); err != nil {
		t.Fatal(err)
	}

	gs := store.NewGrantStore(db)
	g := &domain.Grant{
		CompanyID:         company.ID,
		StakeholderID:     sh.ID,
		ShareClassID:      sc.ID,
		VestingScheduleID: &vs.ID,
		Quantity:          decimal.NewFromInt(48000),
		GrantDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExercisePrice:     decimal.NewFromFloat(0.10),
	}
	if err := gs.Create(ctx, g); err != nil {
		t.Fatal(err)
	}
	if g.ID == "" {
		t.Fatal("expected grant ID to be set")
	}

	grants, err := gs.ListByCompany(ctx, company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants))
	}
	if !grants[0].Quantity.Equal(decimal.NewFromInt(48000)) {
		t.Errorf("Quantity = %s, want 48000", grants[0].Quantity)
	}
}

func TestAuditStore_LogAndList(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "AuditCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	as := store.NewAuditStore(db)
	entry := &domain.AuditEntry{
		EntityType: "company",
		EntityID:   company.ID,
		Action:     "create",
		AfterState: []byte(`{"name":"AuditCo"}`),
	}
	if err := as.Log(ctx, entry); err != nil {
		t.Fatalf("Log: %v", err)
	}
	if entry.ID == "" {
		t.Fatal("expected audit entry ID to be set")
	}

	entries, err := as.ListByEntity(ctx, "company", company.ID)
	if err != nil {
		t.Fatalf("ListByEntity: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != "create" {
		t.Errorf("Action = %s, want create", entries[0].Action)
	}
}

func TestSAFENoteStore_CreateAndMarkConverted(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "SafeCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	ss := store.NewStakeholderStore(db)
	investor := &domain.Stakeholder{
		CompanyID: company.ID,
		Name:      "Investor",
		Email:     "inv@safeco.com",
		Role:      domain.RoleInvestor,
	}
	if err := ss.Create(ctx, investor); err != nil {
		t.Fatal(err)
	}

	scs := store.NewShareClassStore(db)
	sc := &domain.ShareClass{
		CompanyID:           company.ID,
		Name:                "Series A",
		IsPreferred:         true,
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(5000000),
	}
	if err := scs.Create(ctx, sc); err != nil {
		t.Fatal(err)
	}

	frs := store.NewFundingRoundStore(db)
	round := &domain.FundingRound{
		CompanyID:     company.ID,
		Name:          "Series A",
		PreMoneyVal:   decimal.NewFromInt(10000000),
		AmountRaised:  decimal.NewFromInt(5000000),
		PricePerShare: decimal.NewFromFloat(1.50),
		ShareClassID:  sc.ID,
		RoundDate:     time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := frs.Create(ctx, round); err != nil {
		t.Fatal(err)
	}

	sns := store.NewSAFENoteStore(db)
	cap := decimal.NewFromInt(8000000)
	safe := &domain.SAFENote{
		CompanyID:        company.ID,
		StakeholderID:    investor.ID,
		InvestmentAmount: decimal.NewFromInt(500000),
		ValuationCap:     &cap,
		SAFEType:         domain.SAFEPostMoney,
		IssueDate:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := sns.Create(ctx, safe); err != nil {
		t.Fatal(err)
	}

	if err := sns.MarkConverted(ctx, safe.ID, round.ID); err != nil {
		t.Fatal(err)
	}

	got, err := sns.GetByID(ctx, safe.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsConverted {
		t.Error("expected SAFE to be marked as converted")
	}
	if got.ConvertedInRound == nil || *got.ConvertedInRound != round.ID {
		t.Error("expected ConvertedInRound to be set")
	}
}

func TestStakeholderStore_GetByIDs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "BatchCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	ss := store.NewStakeholderStore(db)
	alice := &domain.Stakeholder{CompanyID: company.ID, Name: "Alice", Email: "alice@batch.co", Role: domain.RoleFounder}
	bob := &domain.Stakeholder{CompanyID: company.ID, Name: "Bob", Email: "bob@batch.co", Role: domain.RoleEmployee}
	if err := ss.Create(ctx, alice); err != nil {
		t.Fatal(err)
	}
	if err := ss.Create(ctx, bob); err != nil {
		t.Fatal(err)
	}

	t.Run("multiple hits", func(t *testing.T) {
		got, err := ss.GetByIDs(ctx, []string{alice.ID, bob.ID})
		if err != nil {
			t.Fatalf("GetByIDs: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 results, got %d", len(got))
		}
		if got[alice.ID].Name != "Alice" || got[bob.ID].Name != "Bob" {
			t.Error("unexpected names in batch result")
		}
	})

	t.Run("partial miss", func(t *testing.T) {
		got, err := ss.GetByIDs(ctx, []string{alice.ID, "nonexistent-id"})
		if err != nil {
			t.Fatalf("GetByIDs: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 result, got %d", len(got))
		}
		if got[alice.ID] == nil {
			t.Error("expected Alice in result")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got, err := ss.GetByIDs(ctx, []string{})
		if err != nil {
			t.Fatalf("GetByIDs: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected 0 results, got %d", len(got))
		}
	})
}

func TestShareClassStore_GetByIDs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "ClassBatchCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	scs := store.NewShareClassStore(db)
	common := &domain.ShareClass{
		CompanyID: company.ID, Name: "Common",
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(10000000),
	}
	preferred := &domain.ShareClass{
		CompanyID: company.ID, Name: "Series A", IsPreferred: true,
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(5000000),
	}
	if err := scs.Create(ctx, common); err != nil {
		t.Fatal(err)
	}
	if err := scs.Create(ctx, preferred); err != nil {
		t.Fatal(err)
	}

	t.Run("multiple hits", func(t *testing.T) {
		got, err := scs.GetByIDs(ctx, []string{common.ID, preferred.ID})
		if err != nil {
			t.Fatalf("GetByIDs: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 results, got %d", len(got))
		}
		if got[common.ID].Name != "Common" || got[preferred.ID].Name != "Series A" {
			t.Error("unexpected names in batch result")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got, err := scs.GetByIDs(ctx, []string{})
		if err != nil {
			t.Fatalf("GetByIDs: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected 0 results, got %d", len(got))
		}
	})
}

func TestShareClassStore_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cs := store.NewCompanyStore(db)
	company := &domain.Company{Name: "UniqueCo"}
	if err := cs.Create(ctx, company); err != nil {
		t.Fatal(err)
	}

	scs := store.NewShareClassStore(db)
	sc1 := &domain.ShareClass{
		CompanyID:           company.ID,
		Name:                "Common",
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(10000000),
	}
	if err := scs.Create(ctx, sc1); err != nil {
		t.Fatal(err)
	}

	sc2 := &domain.ShareClass{
		CompanyID:           company.ID,
		Name:                "Common",
		LiquidationMultiple: decimal.NewFromInt(1),
		AuthorizedShares:    decimal.NewFromInt(5000000),
	}
	err := scs.Create(ctx, sc2)
	if err == nil {
		t.Error("expected unique constraint violation for duplicate share class name")
	}
}
