-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE accounts_type_enum AS ENUM ('user_wallet', 'platform', 'escrow', 'fee');
CREATE TYPE accounts_status_enum AS ENUM ('active', 'frozen', 'closed');
CREATE TYPE transactions_type_enum AS ENUM ('reward', 'purchase', 'trade', 'donate', 'refund', 'fee', 'adjustment');
CREATE TYPE transactions_status_enum AS ENUM('pending', 'processing', 'completed', 'failed', 'cancelled', 'expired');
CREATE TYPE fee_rules_rounding_mode_enum AS ENUM('up', 'down', 'half_up', 'half_even');
CREATE TYPE escrows_status_enum AS ENUM('pending', 'processing', 'completed', 'failed', 'cancelled', 'expired');
CREATE TYPE payments_status_enum AS ENUM('created', 'pending', 'completed', 'failed', 'cancelled', 'refunded');

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
  "owner_id" UUID NULL,
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "status" accounts_status_enum NOT NULL DEFAULT 'active',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "transactions" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "type" transactions_type_enum NOT NULL,
  "reference" TEXT UNIQUE, -- idempotency key
  "status" transactions_status_enum NOT NULL,
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

CREATE TABLE "balance_snapshot" (
  "account_id" UUID NOT NULL REFERENCES accounts(id),
  "version" BIGINT NOT NULL DEFAULT 0,
  "balance" BIGINT NOT NULL,
  "last_entry_id" UUID REFERENCES ledger_entries(id),
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  PRIMARY KEY ("account_id", "version")
);

CREATE TABLE "fee_rules" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "transaction_type" transactions_type_enum NOT NULL,
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "percentage" INTEGER,
  "min_amount" BIGINT NULL,
  "max_amount" BIGINT NULL,
  "effective_from" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  "effective_to" TIMESTAMPTZ NULL,
  "rounding_mode" fee_rules_rounding_mode_enum NOT NULL,
  "is_active" BOOLEAN
);

CREATE TABLE "escrows" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "transaction_id" UUID NULL REFERENCES transactions(id),
  "buyer_account_id" UUID REFERENCES accounts(id),
  "seller_account_id" UUID NULL REFERENCES accounts(id),
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "amount" BIGINT NOT NULL,
  "status" escrows_status_enum NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "payments" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "provider" TEXT NOT NULL,
  "provider_ref" TEXT NOT NULL UNIQUE,
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "amount" BIGINT NOT NULL,
  "status" payments_status_enum NOT NULL,
  "transaction_id" UUID NOT NULL REFERENCES transactions(id),
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX "idx_accounts_owner_id" ON accounts(owner_id);
CREATE INDEX "idx_ledger_entries_account_id" ON "ledger_entries"("account_id");
CREATE INDEX "idx_ledger_entries_transaction_id" ON "ledger_entries"("transaction_id");
CREATE INDEX "idx_escrows_buyer_seller_account_id" ON escrows("buyer_account_id", "seller_account_id");
CREATE INDEX "idx_escrows_transaction_id" ON escrows("transaction_id");

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

-------------------------------------------------
--- transactions: allow only status updates ---
-------------------------------------------------
CREATE OR REPLACE FUNCTION allow_transaction_status_update_only() 
RETURNS TRIGGER AS $$
BEGIN
  IF OLD.id IS DISTINCT FROM NEW.id OR OLD.type IS DISTINCT FROM NEW.type
     OR OLD.reference IS DISTINCT FROM NEW.reference OR OLD.actor_id IS DISTINCT FROM NEW.actor_id
     OR OLD.metadata IS DISTINCT FROM NEW.metadata OR OLD.created_at IS DISTINCT FROM NEW.created_at THEN
    RAISE EXCEPTION 'error: updating transactions is only allowed for status';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transactions_trigger_status_update_only
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE PROCEDURE allow_transaction_status_update_only();

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

CREATE OR REPLACE FUNCTION check_ledger_zero_sum_insert()
RETURNS TRIGGER AS $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN (SELECT DISTINCT transaction_id FROM new_entries)
  LOOP
    PERFORM check_ledger_zero_sum_for_tid(r.transaction_id);
  END LOOP;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_ledger_zero_sum_update()
RETURNS TRIGGER AS $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN (SELECT DISTINCT transaction_id FROM old_entries UNION SELECT DISTINCT transaction_id FROM new_entries)
  LOOP
    PERFORM check_ledger_zero_sum_for_tid(r.transaction_id);
  END LOOP;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_ledger_zero_sum_delete()
RETURNS TRIGGER AS $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN (SELECT DISTINCT transaction_id FROM old_entries)
  LOOP
    PERFORM check_ledger_zero_sum_for_tid(r.transaction_id);
  END LOOP;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ledger_entries_zero_sum_insert
AFTER INSERT ON ledger_entries
REFERENCING NEW TABLE AS new_entries
FOR EACH STATEMENT EXECUTE PROCEDURE check_ledger_zero_sum_insert();

CREATE TRIGGER ledger_entries_zero_sum_update
AFTER UPDATE ON ledger_entries
REFERENCING OLD TABLE AS old_entries NEW TABLE AS new_entries
FOR EACH STATEMENT EXECUTE PROCEDURE check_ledger_zero_sum_update();

CREATE TRIGGER ledger_entries_zero_sum_delete
AFTER DELETE ON ledger_entries
REFERENCING OLD TABLE AS old_entries
FOR EACH STATEMENT EXECUTE PROCEDURE check_ledger_zero_sum_delete();
-------------------------------------------------
-------------------------------------------------
-------------------------------------------------

-- +goose Down
DROP TRIGGER IF EXISTS ledger_entries_zero_sum_delete ON ledger_entries;
DROP TRIGGER IF EXISTS ledger_entries_zero_sum_update ON ledger_entries;
DROP TRIGGER IF EXISTS ledger_entries_zero_sum_insert ON ledger_entries;
DROP TRIGGER IF EXISTS transactions_trigger_status_update_only ON transactions;
DROP TRIGGER IF EXISTS transactions_trigger_no_delete ON transactions;
DROP TRIGGER IF EXISTS ledger_entries_trigger_no_delete ON ledger_entries;
DROP TRIGGER IF EXISTS ledger_entries_trigger_no_update ON ledger_entries;

DROP FUNCTION IF EXISTS check_ledger_zero_sum_delete();
DROP FUNCTION IF EXISTS check_ledger_zero_sum_update();
DROP FUNCTION IF EXISTS check_ledger_zero_sum_insert();
DROP FUNCTION IF EXISTS check_ledger_zero_sum_for_tid(UUID);
DROP FUNCTION IF EXISTS allow_transaction_status_update_only();
DROP FUNCTION IF EXISTS block_update();
DROP FUNCTION IF EXISTS block_delete();

DROP TABLE IF EXISTS "payments";
DROP TABLE IF EXISTS "escrows";
DROP TABLE IF EXISTS "fee_rules";
DROP TABLE IF EXISTS "balance_snapshot";
DROP TABLE IF EXISTS "ledger_entries";
DROP TABLE IF EXISTS "transactions";
DROP TABLE IF EXISTS "accounts";
DROP TABLE IF EXISTS "currencies";

DROP TYPE IF EXISTS payments_status_enum;
DROP TYPE IF EXISTS escrows_status_enum;
DROP TYPE IF EXISTS fee_rules_rounding_mode_enum;
DROP TYPE IF EXISTS transactions_status_enum;
DROP TYPE IF EXISTS transactions_type_enum;
DROP TYPE IF EXISTS accounts_status_enum;
DROP TYPE IF EXISTS accounts_type_enum;

DROP EXTENSION IF EXISTS "uuid-ossp";
