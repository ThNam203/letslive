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

INSERT INTO "currencies"("code", "name", "precision") VALUES ("SPARK", "Spark", 2);
INSERT INTO "currencies"("code", "name", "precision") VALUES ("FLARE", "Flare", 2);

CREATE TABLE "accounts" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "type" account_type_enum NOT NULL,
  "owner_id" UUID NULL,
  "currency_code" UUID NOT NULL REFERENCES currencies(code),
  "status" accounts_status_enum NOT NULL DEFAULT 'active',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "transactions" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "type" transactions_type_enum NOT NULL,
  "reference" TEXT UNIQUE, -- idempotency key
  "status" transactions_status_enum NOT NULL,
  "metadata" JSONB NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "ledger_entries" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "transaction_id" UUID NOT NULL REFERENCES transactions(id),
  "account_id" UUID NOT NULL REFERENCES accounts(id),
  "amount" BIGINT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "balance_snapshot" (
  "account_id" UUID PRIMARY KEY REFERENCES accounts(id),
  "balance" BIGINT NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TABLE "fee_rules" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "currency_code" TEXT NOT NULL REFERENCES currencies(code),
  "percentage" INTEGER,
  "rounding_mode" fee_rules_rounding_mode_enum NOT NULL,
  "is_active" BOOLEAN
);

CREATE TABLE "escrows" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "buyer_account_id" UUID REFERENCES accounts(id),
  "seller_account_id" NULL REFERENCES accounts(id),
  "amount" BIGINT,
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
  "created_at" TIMESTAMPZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX "idx_accounts_owner_id" ON accounts(owner_id);
CREATE INDEX "idx_ledger_entires_account_id" ON "ledger_entries"("account_id");
CREATE INDEX "idx_escrows_buyer_seller_account_id" ON escrows("buyer_account_id", "seller_account_id");

-------------------------------------------------
--- block updates and deletion on some tables ---
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

CREATE TRIGGER ledger_entires_trigger_no_update
BEFORE UPDATE ON ledger_entries
FOR EACH STATEMENT EXECUTE FUNCTION block_update();

CREATE TRIGGER ledger_entires_trigger_no_delete
BEFORE DELETE ON ledger_entries
FOR EACH STATEMENT EXECUTE FUNCTION block_delete();

CREATE TRIGGER transactions_trigger_no_update
BEFORE UPDATE ON transactions
FOR EACH STATEMENT EXECUTE FUNCTION block_update();

CREATE TRIGGER transactions_trigger_no_delete
BEFORE DELETE ON transactions
FOR EACH STATEMENT EXECUTE FUNCTION block_delete();
-------------------------------------------------
-------------------------------------------------
-------------------------------------------------

-- +goose Down
DROP EXTENSION IF EXISTS "uuid-ossp";

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
