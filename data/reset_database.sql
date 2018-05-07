CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language 'plpgsql';

DROP TABLE IF EXISTS account CASCADE;
CREATE TABLE account(
  id            VARCHAR(40) PRIMARY KEY,
  login         VARCHAR(40) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  bio           TEXT,
  name          TEXT
);

CREATE TRIGGER account_updated_at_modtime
BEFORE UPDATE ON account
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS email CASCADE;
CREATE TABLE email(
  id          VARCHAR(40)   PRIMARY KEY,
  value       VARCHAR(40)   NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW()
);

CREATE TYPE account_email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY', 'PUBLIC');

DROP TABLE IF EXISTS account_email;
CREATE TABLE account_email(
  user_id     VARCHAR(40),
  email_id    VARCHAR(40),
  type        account_email_type DEFAULT 'EXTRA',
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  verified_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, email_id),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX account_email_user_id_type_key
ON account_email (user_id, type)
WHERE type = ANY('{"PRIMARY", "BACKUP"}');

DROP TABLE IF EXISTS role CASCADE;
CREATE TABLE role(
  id          VARCHAR(40) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER role_updated_at_modtime
BEFORE UPDATE ON role
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS account_role;
CREATE TABLE account_role(
  user_id     VARCHAR(40),
  role_id     VARCHAR(40),
  granted_at  TIMESTAMPTZ   DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TYPE access_level AS ENUM('Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete');
CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');
CREATE TYPE node_type AS ENUM('Label', 'Lesson', 'LessonComment', 'Study', 'User');

DROP TABLE IF EXISTS permission CASCADE;
CREATE TABLE IF NOT EXISTS permission(
  id            VARCHAR(40)   PRIMARY KEY,
  access_level  access_level  NOT NULL,
  audience      audience      NOT NULL,
  type          node_type     NOT NULL,
  created_at    TIMESTAMPTZ     DEFAULT NOW(),
  updated_at    TIMESTAMPTZ     DEFAULT NOW(),
  field         TEXT
);

CREATE UNIQUE INDEX permission_access_level_type_field_key
ON permission (access_level, type, field)
WHERE field IS NOT NULL;

CREATE UNIQUE INDEX permission_access_level_type_key
ON permission (access_level, type)
WHERE field IS NULL;

CREATE TRIGGER permission_updated_at_modtime
BEFORE UPDATE ON permission
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS role_permission;
CREATE TABLE role_permission(
  role_id       VARCHAR(40),
  permission_id VARCHAR(40),
  granted_at    TIMESTAMPTZ   DEFAULT NOW(),
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (permission_id)
    REFERENCES permission (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DROP TABLE IF EXISTS email_verification_token;
CREATE TABLE email_verification_token(
  user_id       VARCHAR(40),
  token         VARCHAR(40),
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  verified_at   TIMESTAMPTZ,
  PRIMARY KEY (user_id, token),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DROP TABLE IF EXISTS password_reset_token;
CREATE TABLE password_reset_token(
  token         VARCHAR(40)   PRIMARY KEY,
  email         VARCHAR(40)   NOT NULL,
  request_ip    INET          NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  user_id       VARCHAR(40),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);

DROP TABLE IF EXISTS study CASCADE;
CREATE TABLE study(
  id            VARCHAR(40) PRIMARY KEY,
  user_id       VARCHAR(40) NOT NULL,
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  published_at  TIMESTAMPTZ,
  description   TEXT,
  name          TEXT,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX study_user_id_key
ON study (user_id);

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

DROP TABLE IF EXISTS lesson_comment;
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

DROP TABLE IF EXISTS label;
CREATE TABLE label(
  id          VARCHAR(40) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   DEFAULT NOW()
); 

CREATE TRIGGER label_updated_at_modtime
BEFORE UPDATE ON label
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
