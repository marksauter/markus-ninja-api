DROP TABLE IF EXISTS email;
CREATE TABLE email(
  id          VARCHAR(40)   PRIMARY KEY,
  value       VARCHAR(40)   NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
);
