CREATE TABLE label(
  id          VARCHAR(40) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   DEFAULT NOW()
) 
