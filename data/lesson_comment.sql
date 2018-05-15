DROP TABLE IF EXISTS lesson_comment CASCADE;
CREATE TABLE lesson_comment(
  id              VARCHAR(40) PRIMARY KEY,
  lesson_id       VARCHAR(40) NOT NULL,
  study_id        VARCHAR(40) NOT NULL,
  user_id         VARCHAR(40) NOT NULL,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  updated_at      TIMESTAMPTZ   DEFAULT NOW(),
  published_at    TIMESTAMPTZ,
  body            TEXT,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX lesson_comment_lesson_id_idx ON lesson_comment (lesson_id);
CREATE INDEX lesson_comment_study_id_idx ON lesson_comment (study_id);
CREATE INDEX lesson_comment_user_id_idx ON lesson_comment (user_id);
CREATE INDEX lesson_comment_published_at_idx ON lesson_comment (published_at DESC NULLS LAST);

CREATE TRIGGER lesson_comment_updated_at_modtime
  BEFORE UPDATE
  ON lesson_comment
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();
