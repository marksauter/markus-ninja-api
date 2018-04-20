CREATE TABLE lesson_comment(
  id              VARCHAR(45) PRIMARY KEY,
  lesson_id       VARCHAR(45) NOT NULL,
  user_id         VARCHAR(45) NOT NULL,
  created_at      TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  last_edited_at  TIMESTAMP,
  published_at    TIMESTAMP,
  body            TEXT,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
