package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/hutfut/vestigo/internal/audit"
	"github.com/hutfut/vestigo/internal/domain"
	"github.com/hutfut/vestigo/internal/graph"
	"github.com/hutfut/vestigo/internal/store"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	cfg := store.Config{
		Host:     envOr("DB_HOST", "localhost"),
		Port:     envIntOr("DB_PORT", 5432),
		User:     envOr("DB_USER", "vestigo"),
		Password: envOr("DB_PASSWORD", "vestigo"),
		DBName:   envOr("DB_NAME", "vestigo"),
		SSLMode:  envOr("DB_SSLMODE", "disable"),
	}

	db, err := store.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	auditStore := store.NewAuditStore(db)
	auditLogger := audit.NewLogger(auditStore)

	resolver := &graph.Resolver{
		DB:               db,
		Companies:        store.NewCompanyStore(db),
		Stakeholders:     store.NewStakeholderStore(db),
		ShareClasses:     store.NewShareClassStore(db),
		VestingSchedules: store.NewVestingScheduleStore(db),
		Grants:           store.NewGrantStore(db),
		FundingRounds:    store.NewFundingRoundStore(db),
		SAFENotes:        store.NewSAFENoteStore(db),
		Audit:            auditLogger,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.SetErrorPresenter(errorPresenter)

	port := envOr("PORT", "8080")

	http.Handle("/", playground.Handler("vestigo", "/query"))
	http.Handle("/query", srv)

	log.Printf("vestigo GraphQL server listening on :%s", port)
	log.Printf("playground available at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func errorPresenter(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)

	var notFound *domain.ErrNotFound
	var conflict *domain.ErrConflict
	var validation *domain.ErrValidation

	switch {
	case errors.As(err, &notFound):
		gqlErr.Extensions = map[string]interface{}{"code": "NOT_FOUND"}
	case errors.As(err, &conflict):
		gqlErr.Extensions = map[string]interface{}{"code": "CONFLICT"}
	case errors.As(err, &validation):
		gqlErr.Extensions = map[string]interface{}{"code": "VALIDATION_ERROR"}
	}

	return gqlErr
}
