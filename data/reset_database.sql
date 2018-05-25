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
  password      BYTEA       NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  profile       TEXT,
  name          TEXT
);

CREATE UNIQUE INDEX user_unique_login_idx
  ON account (LOWER(login));

CREATE TRIGGER user_updated_at_modtime
  BEFORE UPDATE
  ON account
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

CREATE TYPE email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY');

DROP TABLE IF EXISTS email CASCADE;
CREATE TABLE email(
  id          VARCHAR(40) PRIMARY KEY,
  user_id     VARCHAR(40) NOT NULL,
  value       VARCHAR(40) NOT NULL UNIQUE,
  type        email_type  DEFAULT 'EXTRA',
  public      BOOLEAN     DEFAULT FALSE CHECK (verified_at IS NULL),
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW(),
  verified_at TIMESTAMPTZ,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX email_unique_lower_value_idx
  ON email (LOWER(value));
CREATE INDEX email_user_id_idx ON email (user_id);
CREATE UNIQUE INDEX email_user_id_type_key
  ON email (user_id, type)
  WHERE type = ANY('{"PRIMARY", "BACKUP"}');

CREATE TRIGGER email_updated_at_modtime
  BEFORE UPDATE
  ON email
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

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

DROP TABLE IF EXISTS user_role CASCADE;
CREATE TABLE user_role(
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

DROP TYPE IF EXISTS access_level CASCADE;
CREATE TYPE access_level AS ENUM('Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete');
DROP TYPE IF EXISTS audience CASCADE;
CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');
DROP TYPE IF EXISTS node_type CASCADE;
CREATE TYPE node_type AS ENUM('Email', 'EVT', 'Label', 'Lesson', 'LessonComment', 'PRT', 'Study', 'User', 'UserEmail');

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

DROP TABLE IF EXISTS role_permission CASCADE;
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

DROP TABLE IF EXISTS email_verification_token CASCADE;
CREATE TABLE email_verification_token(
  email_id      VARCHAR(40),
  token         VARCHAR(40),
  user_id       VARCHAR(40)   NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  verified_at   TIMESTAMPTZ,
  PRIMARY KEY (email_id, token),
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX email_verification_token_user_id_idx ON email_verification_token (user_id); 

DROP TABLE IF EXISTS password_reset_token CASCADE;
CREATE TABLE password_reset_token(
  user_id       VARCHAR(40),
  token         VARCHAR(40),
  email_id      VARCHAR(40)   NOT NULL,
  request_ip    INET          NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  PRIMARY KEY (user_id, token),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);

DROP TABLE IF EXISTS study CASCADE;
CREATE TABLE study(
  id            VARCHAR(40)   PRIMARY KEY,
  user_id       VARCHAR(40)   NOT NULL,
  name          TEXT          NOT NULL CHECK (name !~' '),
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  updated_at    TIMESTAMPTZ   DEFAULT NOW(),
  advanced_at   TIMESTAMPTZ,
  description   TEXT,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX study_user_id_name_key ON study (user_id, LOWER(name));
CREATE INDEX study_created_at_idx ON study (created_at);
CREATE INDEX study_updated_at_idx ON study (updated_at);
CREATE INDEX study_advanced_at_idx ON study (advanced_at);

CREATE TRIGGER study_updated_at_modtime
BEFORE UPDATE ON study
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

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
  updated_at      TIMESTAMPTZ DEFAULT NOW(),
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

CREATE INDEX lesson_study_id_number_idx ON lesson (study_id, number);
CREATE INDEX lesson_user_id_idx ON lesson (user_id);
CREATE INDEX lesson_created_at_idx ON lesson (created_at);
CREATE INDEX lesson_updated_at_idx ON lesson (updated_at);
CREATE INDEX lesson_published_at_idx ON lesson (published_at DESC NULLS LAST);

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

CREATE TRIGGER lesson_updated_at_modtime
  BEFORE UPDATE
  ON lesson
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

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

DROP TABLE IF EXISTS label CASCADE;
CREATE TABLE label(
  id          VARCHAR(40) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   DEFAULT NOW()
); 

CREATE UNIQUE INDEX label_unique_name_idx
  ON label (LOWER(name));

CREATE TRIGGER label_updated_at_modtime
  BEFORE UPDATE
  ON label
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();
