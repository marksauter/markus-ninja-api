CREATE OR REPLACE FUNCTION inc_study_lesson_number()
  RETURNS trigger AS
$BODY$
DECLARE
  cnt INT;
BEGIN
  SELECT INTO cnt COUNT(*)::INT
    FROM lesson
    WHERE study_id = NEW.study_id;
  NEW.number = cnt + 1;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dec_study_lesson_number()
  RETURNS trigger AS
$BODY$
BEGIN
  UPDATE lesson
  SET number = number - 1
  WHERE study_id = OLD.study_id AND number > OLD.number;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION check_study_lesson_number()
  RETURNS trigger AS
$BODY$
DECLARE
  cnt INT;
BEGIN
  SELECT INTO cnt COUNT(*)::INT
    FROM lesson
    WHERE study_id = NEW.study_id;
  IF NEW.number > cnt THEN
    NEW.number = cnt;
  END IF;

  RETURN NEW;
END
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION move_study_lesson_number()
  RETURNS trigger AS
$BODY$
BEGIN
  IF NEW.number > OLD.number THEN
    UPDATE lesson
    SET number = number - 1
    WHERE study_id = NEW.study_id 
      AND number <= NEW.number
      AND id != NEW.id;
  ELSIF NEW.number < OLD.number THEN
    UPDATE lesson
    SET number = number + 1
    WHERE study_id = NEW.study_id
      AND number >= NEW.number
      AND id != NEW.id;
  END IF;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

DROP TABLE IF EXISTS lesson CASCADE;
CREATE TABLE lesson(
  id              VARCHAR(40) PRIMARY KEY,
  study_id        VARCHAR(40) NOT NULL,    
  user_id         VARCHAR(40) NOT NULL,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  last_edited_at  TIMESTAMPTZ,
  published_at    TIMESTAMPTZ,
  body            TEXT,
  number          INT         CHECK(number > 0),
  title           TEXT,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX lesson_study_id_key
ON lesson (study_id);

CREATE INDEX lesson_user_id_key
ON lesson (user_id);

CREATE TRIGGER insert_new_lesson
  BEFORE INSERT
  ON lesson
  FOR EACH ROW
  EXECUTE PROCEDURE inc_study_lesson_number(); 

CREATE TRIGGER delete_old_lesson
  AFTER DELETE
  ON lesson
  FOR EACH ROW
  EXECUTE PROCEDURE dec_study_lesson_number();

CREATE TRIGGER before_update_lesson
  BEFORE UPDATE
  ON lesson
  FOR EACH ROW
  EXECUTE PROCEDURE check_study_lesson_number();

CREATE TRIGGER after_update_lesson
  AFTER UPDATE
  ON lesson
  FOR EACH ROW
  WHEN (pg_trigger_depth() = 0)
  EXECUTE PROCEDURE move_study_lesson_number();
