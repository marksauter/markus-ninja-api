CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language PLPGSQL;

DROP TABLE IF EXISTS account;
CREATE TABLE account(
  id            VARCHAR(40) PRIMARY KEY,
  login         VARCHAR(40) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  bio           TEXT,
  name          TEXT
);

CREATE TRIGGER account_updated_at_modtime
BEFORE UPDATE ON account
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
