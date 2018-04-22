CREATE TABLE account(
  id            VARCHAR(45) PRIMARY KEY,
  login         VARCHAR(255) NOT NULL UNIQUE,
  primary_email VARCHAR(355) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  bio           TEXT,
  email         TEXT,
  name          TEXT
);

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';
