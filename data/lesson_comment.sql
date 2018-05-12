CREATE TABLE lesson_comment(
  id              VARCHAR(40) PRIMARY KEY,
  lesson_id       VARCHAR(40) NOT NULL,
  user_id         VARCHAR(40) NOT NULL,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  last_edited_at  TIMESTAMPTZ,
  published_at    TIMESTAMPTZ,
  body            TEXT,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX lesson_comment_lesson_id_key
ON lesson_comment (lesson_id);

CREATE INDEX lesson_comment_user_id_key
ON lesson_comment (user_id);