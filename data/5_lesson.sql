CREATE TABLE lesson(
  id              VARCHAR(45) PRIMARY KEY,
  study_id        VARCHAR(45) NOT NULL,    
  user_id         VARCHAR(45) NOT NULL,
  created_at      TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  last_edited_at  TIMESTAMP,
  published_at    TIMESTAMP,
  body            TEXT,
  number          INT,
  title           TEXT,
  CONSTRAINT lesson_study_id_fkey FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT lesson_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);
