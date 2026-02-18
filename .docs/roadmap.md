# vestigo roadmap

## 1. Fix N+1 queries in resolvers

`CapTable`, `ModelDilution`, and `Waterfall` resolvers call `GetByID` per grant inside loops. Add batch-fetch methods (`GetByIDs` or `ListByIDs`) to the stakeholder and share class stores, then refactor resolvers to pre-fetch in a single query.

## 2. Implement non-participating preferred conversion in waterfall

The waterfall engine currently pays non-participating preferred their liquidation preference without comparing it to their as-converted common payout. Add the conversion comparison: for each non-participating preferred class, calculate the as-converted pro-rata payout and take the higher of the two. This requires iterating to find the breakpoint where conversion becomes optimal.

## 3. Add typed errors

Replace `fmt.Errorf` string errors with a small set of domain error types: `ErrNotFound`, `ErrConflict`, `ErrValidation`. Map these to appropriate GraphQL error extensions in the resolver layer so clients get structured error responses instead of opaque strings.

## 4. Add a `Known Limitations` section to the README

Call out the simplifications explicitly: N+1 queries, non-participating preferred conversion not fully modeled, no auth, single-tenant. Frame each as a design choice for scope, not an oversight. Acknowledging limits before a reviewer finds them changes the read entirely.

## 5. Clean up the domain/model duplication

Evaluate whether the full two-type-system approach (domain types + gqlgen model types + converter layer) is worth the boilerplate for a project this size. Options: configure gqlgen to autobind domain types directly, or keep the separation but extract the converters into a dedicated `internal/graph/convert` package to reduce resolver noise.

## 6. Structure git history

Before pushing to GitHub, rewrite the history into logical, incremental commits that reflect how the project was built:
- `init: go module, project structure, docker-compose`
- `schema: postgresql migrations and domain types`
- `engine: vesting calculator with tests`
- `engine: SAFE conversion with tests`
- `engine: dilution modeling with tests`
- `engine: waterfall analysis with tests`
- `store: postgresql repository layer`
- `store: integration tests with testcontainers`
- `graphql: schema, resolvers, server entrypoint`
- `audit: append-only audit logging`
- `docs: README`

Each commit should compile and pass its relevant tests.
