CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum types
CREATE TYPE stakeholder_role AS ENUM ('founder', 'employee', 'investor', 'advisor', 'consultant');
CREATE TYPE safe_type AS ENUM ('pre_money', 'post_money');
CREATE TYPE vesting_frequency AS ENUM ('monthly', 'quarterly', 'annually');
CREATE TYPE acceleration_trigger AS ENUM ('none', 'single_trigger', 'double_trigger');

CREATE TABLE companies (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE stakeholders (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID NOT NULL REFERENCES companies(id),
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    role        stakeholder_role NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (company_id, email)
);

CREATE TABLE share_classes (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id              UUID NOT NULL REFERENCES companies(id),
    name                    TEXT NOT NULL,
    is_preferred            BOOLEAN NOT NULL DEFAULT false,
    liquidation_multiple    NUMERIC(10, 4) NOT NULL DEFAULT 1.0,
    is_participating        BOOLEAN NOT NULL DEFAULT false,
    participation_cap       NUMERIC(10, 4),  -- NULL means uncapped
    price_per_share         NUMERIC(20, 10),
    seniority               INTEGER NOT NULL DEFAULT 0, -- higher = more senior in waterfall
    authorized_shares       NUMERIC(20, 4) NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE (company_id, name),
    CONSTRAINT chk_liquidation_multiple CHECK (liquidation_multiple >= 0),
    CONSTRAINT chk_participation_cap CHECK (participation_cap IS NULL OR participation_cap > 0),
    CONSTRAINT chk_seniority CHECK (seniority >= 0)
);

CREATE TABLE vesting_schedules (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cliff_months            INTEGER NOT NULL DEFAULT 12,
    total_months            INTEGER NOT NULL DEFAULT 48,
    frequency               vesting_frequency NOT NULL DEFAULT 'monthly',
    acceleration_trigger    acceleration_trigger NOT NULL DEFAULT 'none',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_cliff CHECK (cliff_months >= 0),
    CONSTRAINT chk_total CHECK (total_months > 0),
    CONSTRAINT chk_cliff_le_total CHECK (cliff_months <= total_months)
);

CREATE TABLE grants (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id              UUID NOT NULL REFERENCES companies(id),
    stakeholder_id          UUID NOT NULL REFERENCES stakeholders(id),
    share_class_id          UUID NOT NULL REFERENCES share_classes(id),
    vesting_schedule_id     UUID REFERENCES vesting_schedules(id),
    quantity                NUMERIC(20, 4) NOT NULL,
    grant_date              DATE NOT NULL,
    exercise_price          NUMERIC(20, 10) NOT NULL DEFAULT 0,
    is_exercised            BOOLEAN NOT NULL DEFAULT false,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    CONSTRAINT chk_quantity CHECK (quantity > 0),
    CONSTRAINT chk_exercise_price CHECK (exercise_price >= 0)
);

CREATE TABLE funding_rounds (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID NOT NULL REFERENCES companies(id),
    name                TEXT NOT NULL,
    pre_money_valuation NUMERIC(20, 4) NOT NULL,
    amount_raised       NUMERIC(20, 4) NOT NULL,
    price_per_share     NUMERIC(20, 10) NOT NULL,
    share_class_id      UUID NOT NULL REFERENCES share_classes(id),
    round_date          DATE NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,
    CONSTRAINT chk_pre_money CHECK (pre_money_valuation > 0),
    CONSTRAINT chk_amount CHECK (amount_raised > 0),
    CONSTRAINT chk_pps CHECK (price_per_share > 0)
);

CREATE TABLE safe_notes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID NOT NULL REFERENCES companies(id),
    stakeholder_id      UUID NOT NULL REFERENCES stakeholders(id),
    investment_amount   NUMERIC(20, 4) NOT NULL,
    valuation_cap       NUMERIC(20, 4),      -- NULL means no cap
    discount_rate       NUMERIC(5, 4),        -- e.g. 0.20 for 20% discount; NULL means no discount
    safe_type           safe_type NOT NULL,
    is_converted        BOOLEAN NOT NULL DEFAULT false,
    converted_in_round  UUID REFERENCES funding_rounds(id),
    issue_date          DATE NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,
    CONSTRAINT chk_investment CHECK (investment_amount > 0),
    CONSTRAINT chk_cap CHECK (valuation_cap IS NULL OR valuation_cap > 0),
    CONSTRAINT chk_discount CHECK (discount_rate IS NULL OR (discount_rate > 0 AND discount_rate < 1))
);

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type     TEXT NOT NULL,
    entity_id       UUID NOT NULL,
    action          TEXT NOT NULL,
    actor_id        UUID,
    before_state    JSONB,
    after_state     JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_stakeholders_company ON stakeholders(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_share_classes_company ON share_classes(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_grants_company ON grants(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_grants_stakeholder ON grants(stakeholder_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_funding_rounds_company ON funding_rounds(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_safe_notes_company ON safe_notes(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at);

-- Trigger to auto-update updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_companies_updated_at BEFORE UPDATE ON companies FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_stakeholders_updated_at BEFORE UPDATE ON stakeholders FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_share_classes_updated_at BEFORE UPDATE ON share_classes FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_grants_updated_at BEFORE UPDATE ON grants FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_funding_rounds_updated_at BEFORE UPDATE ON funding_rounds FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_safe_notes_updated_at BEFORE UPDATE ON safe_notes FOR EACH ROW EXECUTE FUNCTION update_updated_at();
