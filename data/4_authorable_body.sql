CREATE TABLE authorable_body(
  id             VARCHAR(45) PRIMARY KEY,
  user_id        VARCHAR(45) NOT NULL,
  created_at     TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  last_edited_at TIMESTAMP,
  published_at   TIMESTAMP,
  body           TEXT,
  CONSTRAINT authorable_body_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);
