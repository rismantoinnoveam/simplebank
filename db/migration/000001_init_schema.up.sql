CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL,
    currency VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE entries (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL,
    to_account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE entries ADD FOREIGN KEY (account_id) REFERENCES accounts (id);
ALTER TABLE transfers ADD FOREIGN KEY (from_account_id) REFERENCES accounts (id);
ALTER TABLE transfers ADD FOREIGN KEY (to_account_id) REFERENCES accounts (id);

CREATE INDEX accounts_index_0 ON accounts (owner);
CREATE INDEX entries_index_1 ON entries (account_id);
CREATE INDEX transfers_index_2 ON transfers (from_account_id);
CREATE INDEX transfers_index_3 ON transfers (to_account_id);
CREATE INDEX transfers_index_4 ON transfers (from_account_id, to_account_id);


COMMENT ON COLUMN entries.amount IS 'Can be negative or positive';
COMMENT ON COLUMN transfers.amount IS 'Must be positive';

