# Available permissions

CREATE TYPE operation AS (
  operator  TEXT,
  type      TEXT 
)
CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');

CREATE TABLE permission(
  id          VARCHAR(45) PRIMARY KEY,
  operation   VARCHAR(45) NOT NULL UNIQUE,
  fields      TEXT [],
  audience    audience  NOT NULL,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
