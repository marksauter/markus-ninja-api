CREATE TABLE lesson(
  id         VARCHAR(45) PRIMARY KEY,
  user_id    VARCHAR(45) NOT NULL,
  body_id    VARCHAR(45) NOT NULL,
  created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP,
  title      TEXT,
  CONSTRAINT lesson_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT lesson_body_id_fkey FOREIGN KEY (body_id)
    REFERENCES authorable_body (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
);
