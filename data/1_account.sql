CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE account(
  id            VARCHAR(40) PRIMARY KEY,
  login         VARCHAR(40) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  bio           TEXT,
  name          TEXT
);

CREATE TABLE email(
  id          VARCHAR(40)   PRIMARY KEY,
  value       VARCHAR(40)   NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
)

CREATE TYPE account_email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY', 'PUBLIC')

CREATE TABLE account_email(
  user_id     VARCHAR(40),
  email_id    VARCHAR(40),
  type        account_email_type DEFAULT 'EXTRA',
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  verified_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, email_id),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
)

CREATE UNIQUE INDEX account_email_user_id_type_key
ON account_email (user_id, type)
WHERE type = ANY('PRIMARY', 'BACKUP')
