DROP TRIGGER IF EXISTS trg_safe_notes_updated_at ON safe_notes;
DROP TRIGGER IF EXISTS trg_funding_rounds_updated_at ON funding_rounds;
DROP TRIGGER IF EXISTS trg_grants_updated_at ON grants;
DROP TRIGGER IF EXISTS trg_share_classes_updated_at ON share_classes;
DROP TRIGGER IF EXISTS trg_stakeholders_updated_at ON stakeholders;
DROP TRIGGER IF EXISTS trg_companies_updated_at ON companies;
DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS safe_notes;
DROP TABLE IF EXISTS funding_rounds;
DROP TABLE IF EXISTS grants;
DROP TABLE IF EXISTS vesting_schedules;
DROP TABLE IF EXISTS share_classes;
DROP TABLE IF EXISTS stakeholders;
DROP TABLE IF EXISTS companies;

DROP TYPE IF EXISTS acceleration_trigger;
DROP TYPE IF EXISTS vesting_frequency;
DROP TYPE IF EXISTS safe_type;
DROP TYPE IF EXISTS stakeholder_role;

DROP EXTENSION IF EXISTS "uuid-ossp";
