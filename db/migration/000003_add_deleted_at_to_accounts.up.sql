ALTER TABLE accounts ADD COLUMN deleted_at TIMESTAMPTZ;

DROP INDEX IF EXISTS accounts_owner_currency_idx;

CREATE UNIQUE INDEX accounts_owner_currency_idx ON "accounts" ("owner", "currency") 
WHERE deleted_at IS NULL;

CREATE INDEX accounts_deleted_at_idx ON accounts (deleted_at) WHERE deleted_at IS NULL;