CREATE TABLE study(
  id            VARCHAR(45) PRIMARY KEY,
  user_id       VARCHAR(45) NOT NULL,
  created_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  published_at  TIMESTAMP,
  description   TEXT,
  name          TEXT,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
