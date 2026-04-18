DROP INDEX IF EXISTS accounts_owner_currency_idx;
DROP INDEX IF EXISTS accounts_deleted_at_idx;

ALTER TABLE accounts DROP COLUMN IF EXISTS deleted_at;

CREATE UNIQUE INDEX accounts_owner_currency_idx ON "accounts" ("owner", "currency");