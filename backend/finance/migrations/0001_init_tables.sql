-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE accounts_type_enum AS ENUM ('user_wallet', 'platform', 'escrow', 'fee');
CREATE TYPE accounts_status_enum AS ENUM ('active', 'frozen', 'closed');
CREATE TYPE transactions_type_enum AS ENUM ('reward', 'purchase', 'trade', 'donate', 'refund', 'fee', 'adjustment');
CREATE TYPE process_status_enum AS ENUM('created', 'processing', 'completed', 'failed', 'cancelled');
CREATE TYPE fee_rules_rounding_mode_enum AS ENUM('up', 'down', 'half_up', 'half_even');

CREATE TABLE "currencies" (
  "code" TEXT PRIMARY KEY,
  "name" TEXT,
  "precision" INTEGER -- 2 --> 0.01
);

INSERT INTO "currencies"("code", "name", "precision") VALUES ('SPARK', 'Spark', 2);
INSERT INTO "currencies"("code", "name", "precision") VALUES ('FLARE', 'Flare', 2);

CREATE TABLE "accounts" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "type" accounts_type_enum NOT NULL,
  "owner_id" UUID NULL, -- got from user service
  "status" accounts_status_enum NOT NULL DEFAULT 'active',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "transactions" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "type" transactions_type_enum NOT NULL,
  "reference" TEXT UNIQUE, -- idempotency key
  "status" process_status_enum NOT NULL,
  "actor_id" UUID NULL, -- who/service initiated (audit)
  "metadata" JSONB NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "ledger_entries" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "transaction_id" UUID NOT NULL REFERENCES transactions(id),
  "account_id" UUID NOT NULL REFERENCES accounts(id),
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "amount" BIGINT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "account_balances" (
  "account_id" UUID NOT NULL REFERENCES accounts(id),
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "balance" BIGINT NOT NULL,
  "last_entry_id" UUID NULL REFERENCES ledger_entries(id),
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  PRIMARY KEY (account_id, currency_code)
);

CREATE TABLE "fee_rules" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "transaction_type" transactions_type_enum NOT NULL,
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "percentage" INTEGER,
  "effective_from" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  "effective_to" TIMESTAMPTZ NULL,
  "rounding_mode" fee_rules_rounding_mode_enum NOT NULL,
  "is_active" BOOLEAN
);

CREATE TABLE "payments" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "provider" TEXT NOT NULL,
  "provider_ref" TEXT NOT NULL UNIQUE,
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "amount" BIGINT NOT NULL,
  "status" process_status_enum NOT NULL,
  "transaction_id" UUID NOT NULL REFERENCES transactions(id),
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX "idx_accounts_owner_id" ON accounts(owner_id);
CREATE INDEX "idx_ledger_entries_account_id" ON "ledger_entries"("account_id");
CREATE INDEX "idx_ledger_entries_transaction_id" ON "ledger_entries"("transaction_id");

-------------------------------------------------
--- block updates and deletion on ledger_entries ---
-------------------------------------------------
CREATE OR REPLACE FUNCTION block_delete() 
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'error: deleting records from % is forbidden', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION block_update() 
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'error: updating records in % is forbidden', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ledger_entries_trigger_no_update
BEFORE UPDATE ON ledger_entries
FOR EACH STATEMENT EXECUTE PROCEDURE block_update();

CREATE TRIGGER ledger_entries_trigger_no_delete
BEFORE DELETE ON ledger_entries
FOR EACH STATEMENT EXECUTE PROCEDURE block_delete();

CREATE TRIGGER transactions_trigger_no_delete
BEFORE DELETE ON transactions
FOR EACH STATEMENT EXECUTE PROCEDURE block_delete();

-------------------------------------------------
--- double-entry: sum(amount) = 0 per transaction ---
-------------------------------------------------
CREATE OR REPLACE FUNCTION check_ledger_zero_sum_for_tid(tid UUID)
RETURNS void AS $$
DECLARE
  total BIGINT;
BEGIN
  SELECT COALESCE(SUM(amount), 0) INTO total FROM ledger_entries WHERE transaction_id = tid;
  IF total != 0 THEN
    RAISE EXCEPTION 'ledger entries for transaction % sum to % (must be 0)', tid, total;
  END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION allow_transaction_status_update_only()
RETURNS trigger AS $$
BEGIN
  -- completed is terminal
  IF OLD.status = 'completed'
     AND NEW.status IS DISTINCT FROM 'completed'
  THEN
    RAISE EXCEPTION 'completed transactions are immutable';
  END IF;

  -- only status may change
  IF (OLD.* IS DISTINCT FROM NEW.*)
     AND (OLD.status IS NOT DISTINCT FROM NEW.status)
  THEN
    RAISE EXCEPTION
      'only status column may be updated on transactions';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transactions_status_update_only
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION allow_transaction_status_update_only();

CREATE OR REPLACE FUNCTION enforce_zero_sum_on_completion()
RETURNS TRIGGER AS $$
DECLARE
  total BIGINT;
BEGIN
  -- Only care about transition → completed
  IF NEW.status = 'completed' AND OLD.status IS DISTINCT FROM 'completed' THEN
    SELECT COALESCE(SUM(amount), 0)
      INTO total
      FROM ledger_entries
      WHERE transaction_id = NEW.id;

    IF total != 0 THEN
      RAISE EXCEPTION
        'cannot complete transaction %, ledger sum is % (must be 0)',
        NEW.id, total;
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transactions_enforce_zero_sum_on_complete
BEFORE UPDATE OF status ON transactions
FOR EACH ROW
EXECUTE PROCEDURE enforce_zero_sum_on_completion();

-------------------------------------------------
-------------------------------------------------
-------------------------------------------------

-- +goose Down

-------------------------------------------------
-- drop triggers
-------------------------------------------------
DROP TRIGGER IF EXISTS transactions_enforce_zero_sum_on_complete ON transactions;
DROP TRIGGER IF EXISTS transactions_status_update_only ON transactions;
DROP TRIGGER IF EXISTS transactions_trigger_no_delete ON transactions;

DROP TRIGGER IF EXISTS ledger_entries_trigger_no_delete ON ledger_entries;
DROP TRIGGER IF EXISTS ledger_entries_trigger_no_update ON ledger_entries;

-------------------------------------------------
-- drop functions
-------------------------------------------------
DROP FUNCTION IF EXISTS enforce_zero_sum_on_completion();
DROP FUNCTION IF EXISTS allow_transaction_status_update_only();
DROP FUNCTION IF EXISTS check_ledger_zero_sum_for_tid(UUID);
DROP FUNCTION IF EXISTS block_update();
DROP FUNCTION IF EXISTS block_delete();

-------------------------------------------------
-- drop tables (reverse dependency order)
-------------------------------------------------
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS fee_rules;
DROP TABLE IF EXISTS account_balances;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS currencies;

-------------------------------------------------
-- drop enums
-------------------------------------------------
DROP TYPE IF EXISTS fee_rules_rounding_mode_enum;
DROP TYPE IF EXISTS process_status_enum;
DROP TYPE IF EXISTS transactions_type_enum;
DROP TYPE IF EXISTS accounts_status_enum;
DROP TYPE IF EXISTS accounts_type_enum;

-------------------------------------------------
-- drop extensions
-------------------------------------------------
DROP EXTENSION IF EXISTS "uuid-ossp";
