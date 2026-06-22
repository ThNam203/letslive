-- +goose Up
-- A single global escrow account holds deposited fiat-equivalent balances
-- before they are credited to the user wallet. ledger_entries.amount on this
-- account is debited as the user wallet is credited, keeping each transaction
-- zero-sum.
INSERT INTO "accounts" ("id", "type", "owner_id", "status")
VALUES ('00000000-0000-0000-0000-000000000001', 'escrow', NULL, 'active')
ON CONFLICT ("id") DO NOTHING;

-- +goose Down
DELETE FROM "accounts" WHERE "id" = '00000000-0000-0000-0000-000000000001';
