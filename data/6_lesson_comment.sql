CREATE TABLE lesson_comment(
  id         VARCHAR(45) PRIMARY KEY,
  body_id    VARCHAR(45) NOT NULL,
  lesson_id  VARCHAR(45) NOT NULL,
  user_id    VARCHAR(45) NOT NULL,
  created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT lesson_comment_body_id_fkey FOREIGN KEY (body_id)
    REFERENCES authorable_body (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT lesson_comment_lesson_id_fkey FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT lesson_comment_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);
