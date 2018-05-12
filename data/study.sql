DROP TABLE IF EXISTS study CASCADE;
CREATE TABLE study(
  id            VARCHAR(40)   PRIMARY KEY,
  user_id       VARCHAR(40)   NOT NULL,
  name          TEXT          NOT NULL,
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  updated_at    TIMESTAMPTZ   DEFAULT NOW(),
  published_at  TIMESTAMPTZ,
  description   TEXT,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX study_user_id_name_key
ON study (user_id, name);

CREATE TRIGGER study_updated_at_modtime
BEFORE UPDATE ON study
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();