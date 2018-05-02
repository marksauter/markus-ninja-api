CREATE TABLE lesson(
  id              VARCHAR(40) PRIMARY KEY,
  study_id        VARCHAR(40) NOT NULL,    
  user_id         VARCHAR(40) NOT NULL,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  last_edited_at  TIMESTAMPTZ,
  published_at    TIMESTAMPTZ,
  body            TEXT,
  number          INT,
  title           TEXT,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
