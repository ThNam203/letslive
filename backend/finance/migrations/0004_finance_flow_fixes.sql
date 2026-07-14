-- +goose Up

-- Deposits were recorded as 'purchase' because no deposit type existed.
-- New value is only used by code running after this migration completes.
ALTER TYPE transactions_type_enum ADD VALUE IF NOT EXISTS 'deposit';

-- shop_items.price is denominated in whole platform-currency units of this currency.
ALTER TABLE shop_items
  ADD COLUMN currency_code TEXT NOT NULL DEFAULT 'SPARK' REFERENCES currencies(code);

-- One wallet per user; concurrent create (deposit initiate vs webhook) must not duplicate.
CREATE UNIQUE INDEX uq_accounts_user_wallet_owner
  ON accounts(owner_id)
  WHERE type = 'user_wallet';

-- Dead schema: fee_rules and check_ledger_zero_sum_for_tid were never referenced
-- by any code or trigger (zero-sum enforcement lives in enforce_zero_sum_on_completion).
DROP TABLE IF EXISTS fee_rules;
DROP TYPE IF EXISTS fee_rules_rounding_mode_enum;
DROP FUNCTION IF EXISTS check_ledger_zero_sum_for_tid(UUID);

-- +goose Down

-- Note: Postgres cannot remove an enum value; 'deposit' stays on down-migration.

DROP INDEX IF EXISTS uq_accounts_user_wallet_owner;

ALTER TABLE shop_items DROP COLUMN IF EXISTS currency_code;

CREATE TYPE fee_rules_rounding_mode_enum AS ENUM('up', 'down', 'half_up', 'half_even');

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

-- +goose StatementBegin
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
-- +goose StatementEnd
