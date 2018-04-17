CREATE TABLE label(
  id          VARCHAR(45) PRIMARY KEY,
  name        VARCHAR(45) NOT NULL UNIQUE,
  created_at  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  color       TEXT        NOT NULL,
  description TEXT,
) 
