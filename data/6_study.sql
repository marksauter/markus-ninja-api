CREATE TABLE study(
  id            VARCHAR(40) PRIMARY KEY,
  user_id       VARCHAR(40) NOT NULL,
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  published_at  TIMESTAMPTZ,
  description   TEXT,
  name          TEXT,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
