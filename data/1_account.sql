CREATE TABLE account(
  id            VARCHAR(45) PRIMARY KEY,
  login         VARCHAR(255) NOT NULL UNIQUE,
  primary_email VARCHAR(355) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  login_at      TIMESTAMP,
  updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  bio           TEXT,
  email         TEXT,
  name          TEXT
);
