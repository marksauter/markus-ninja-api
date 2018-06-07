CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION insert_name_tokens() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.name_array IS NOT NULL THEN
    NEW.name_tokens = array_to_tsvector(New.name_array);
    NEW.name_array = NULL;
  END IF;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_name_tokens() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.name != OLD.name THEN
    IF NEW.name_array IS NULL THEN
      RAISE EXCEPTION '`name_array` must not be null';
    END IF;
    NEW.name_tokens = array_to_tsvector(New.name_array);
    NEW.name_array = NULL;
  END IF;
  RETURN NEW;
END;
$$ language 'plpgsql';

DROP TABLE IF EXISTS account CASCADE;
CREATE TABLE account(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  login         VARCHAR(40)  NOT NULL,
  name          TEXT,
  name_array    TEXT []      CHECK(name_array IS NULL),
  name_tokens   TSVECTOR,
  password      BYTEA        NOT NULL,
  profile       TEXT,
  updated_at    TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX account_unique_login_idx
  ON account (LOWER(login));

CREATE INDEX account_login_text_pattern_ops_idx
  ON account(LOWER(login) text_pattern_ops);

CREATE TRIGGER account_insert_name_tokens
  BEFORE INSERT ON account
  FOR EACH ROW EXECUTE PROCEDURE insert_name_tokens();

CREATE TRIGGER account_update_name_tokens
  BEFORE UPDATE ON account
  FOR EACH ROW EXECUTE PROCEDURE update_name_tokens();

CREATE TRIGGER account_updated_at_modtime
  BEFORE UPDATE ON account
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TYPE email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY');

DROP TABLE IF EXISTS email CASCADE;
CREATE TABLE email(
  id          VARCHAR(100) PRIMARY KEY,
  user_id     VARCHAR(100) NOT NULL,
  value       VARCHAR(40) NOT NULL,
  type        email_type  DEFAULT 'EXTRA',
  public      BOOLEAN     DEFAULT FALSE,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW(),
  verified_at TIMESTAMPTZ,
  CONSTRAINT check_verified_before_public
    CHECK ((public AND verified_at IS NOT NULL) OR NOT public),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX email_unique_value_idx
  ON email (LOWER(value));
CREATE INDEX email_value_text_pattern_ops_idx
  ON email(LOWER(value) text_pattern_ops);
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
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER role_updated_at_modtime
BEFORE UPDATE ON role
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS user_role CASCADE;
CREATE TABLE user_role(
  granted_at  TIMESTAMPTZ   DEFAULT NOW(),
  role_id     VARCHAR(100),
  user_id     VARCHAR(100),
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
CREATE TYPE node_type AS ENUM(
  'Email',
  'EVT',
  'Label',
  'Lesson',
  'LessonComment',
  'PRT',
  'Study',
  'User',
  'UserAsset'
);

DROP TABLE IF EXISTS permission CASCADE;
CREATE TABLE IF NOT EXISTS permission(
  access_level access_level NOT NULL,
  audience     audience     NOT NULL,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  field        TEXT,
  id           VARCHAR(100) PRIMARY KEY,
  type         node_type    NOT NULL,
  updated_at   TIMESTAMPTZ  DEFAULT NOW()
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
  granted_at    TIMESTAMPTZ   DEFAULT NOW(),
  role_id       VARCHAR(100),
  permission_id VARCHAR(100),
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
  email_id      VARCHAR(100),
  token         VARCHAR(40),
  user_id       VARCHAR(100)   NOT NULL,
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

CREATE INDEX email_verification_token_user_id_idx
  ON email_verification_token (user_id); 

DROP TABLE IF EXISTS password_reset_token CASCADE;
CREATE TABLE password_reset_token(
  user_id       VARCHAR(100),
  token         VARCHAR(40),
  email_id      VARCHAR(100)   NOT NULL,
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
  advanced_at   TIMESTAMPTZ,
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  description   TEXT,
  id            VARCHAR(100)  PRIMARY KEY,
  name          TEXT          NOT NULL CHECK (name !~ ' '),
  name_array    TEXT []       CHECK(name_array IS NULL),
  name_tokens   TSVECTOR      NOT NULL,
  updated_at    TIMESTAMPTZ   DEFAULT NOW(),
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX study_user_id_name_key
  ON study (user_id, LOWER(name));
CREATE INDEX study_user_id_advanced_at_idx
  ON study (user_id, advanced_at);
CREATE INDEX study_user_id_updated_at_idx
  ON study (user_id, updated_at);

CREATE TRIGGER study_insert_name_tokens
  BEFORE INSERT ON study
  FOR EACH ROW EXECUTE PROCEDURE insert_name_tokens();

CREATE TRIGGER study_update_name_tokens
  BEFORE UPDATE ON study
  FOR EACH ROW EXECUTE PROCEDURE update_name_tokens();

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

CREATE OR REPLACE FUNCTION insert_title_tokens() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.title_array IS NULL THEN
    RAISE EXCEPTION '`title_array` must not be null';
  END IF;
  NEW.title_tokens = array_to_tsvector(New.title_array);
  NEW.title_array = NULL;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_title_tokens() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.title != OLD.title THEN
    IF NEW.title_array IS NULL THEN
      RAISE EXCEPTION '`title_array` must not be null';
    END IF;
    NEW.title_tokens = array_to_tsvector(New.title_array);
    NEW.title_array = NULL;
  END IF;
  RETURN NEW;
END;
$$ language 'plpgsql';

DROP TABLE IF EXISTS lesson CASCADE;
CREATE TABLE lesson(
  created_at      TIMESTAMPTZ  DEFAULT NOW(),
  body            TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  number          INT          CHECK(number > 0),
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,    
  title           TEXT         NOT NULL,
  title_array     TEXT []      CHECK(title_array IS NULL),
  title_tokens    TSVECTOR     NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT NOW(),
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX lesson_user_id_study_id_number_idx
  ON lesson (user_id, study_id, number);
CREATE INDEX lesson_user_id_study_id_published_at_idx
  ON lesson (user_id, study_id, published_at DESC NULLS LAST);
CREATE INDEX lesson_user_id_study_id_updated_at_idx
  ON lesson (user_id, study_id, updated_at);

CREATE TRIGGER insert_lesson_title_tokens
  BEFORE INSERT ON lesson
  FOR EACH ROW EXECUTE PROCEDURE insert_title_tokens(); 

CREATE TRIGGER update_lesson_title_tokens
  BEFORE UPDATE ON lesson
  FOR EACH ROW EXECUTE PROCEDURE update_title_tokens(); 

CREATE TRIGGER insert_lesson_number
  BEFORE INSERT ON lesson
  FOR EACH ROW EXECUTE PROCEDURE inc_study_lesson_number(); 

CREATE TRIGGER delete_lesson_number
  AFTER DELETE ON lesson
  FOR EACH ROW EXECUTE PROCEDURE dec_study_lesson_number();

CREATE TRIGGER before_update_lesson_number
  BEFORE UPDATE ON lesson
  FOR EACH ROW EXECUTE PROCEDURE check_study_lesson_number();

CREATE TRIGGER after_update_lesson_number
  AFTER UPDATE ON lesson
  FOR EACH ROW WHEN (pg_trigger_depth() = 0)
  EXECUTE PROCEDURE move_study_lesson_number();

CREATE TRIGGER lesson_updated_at_modtime
  BEFORE UPDATE ON lesson
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS lesson_comment CASCADE;
CREATE TABLE lesson_comment(
  body            TEXT,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  id              VARCHAR(100) PRIMARY KEY,
  lesson_id       VARCHAR(100) NOT NULL,
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,
  user_id         VARCHAR(100) NOT NULL,
  updated_at      TIMESTAMPTZ   DEFAULT NOW(),
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

CREATE INDEX lesson_comment_user_id_study_id_lesson_id_published_at_idx
  ON lesson_comment (user_id, study_id, lesson_id, published_at ASC NULLS LAST);

CREATE TRIGGER lesson_comment_updated_at_modtime
  BEFORE UPDATE
  ON lesson_comment
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS label CASCADE;
CREATE TABLE label(
  created_at  TIMESTAMPTZ   DEFAULT NOW(),
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  updated_at  TIMESTAMPTZ   DEFAULT NOW()
); 

CREATE UNIQUE INDEX label_unique_name_idx
  ON label (LOWER(name));

CREATE TRIGGER label_updated_at_modtime
  BEFORE UPDATE
  ON label
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS user_asset CASCADE;
CREATE TABLE user_asset(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  key           TEXT         NOT NULL,
  name          TEXT         NOT NULL,
  name_array    TEXT []      CHECK(name_array IS NULL),
  name_tokens   TSVECTOR     NOT NULL,
  original_name TEXT         NOT NULL, 
  published_at  TIMESTAMPTZ,
  size          BIGINT       NOT NULL,
  study_id      VARCHAR(100) NOT NULL,
  subtype       TEXT         NOT NULL,
  type          TEXT         NOT NULL,
  updated_at    TIMESTAMPTZ  DEFAULT NOW(),
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX user_asset_unique_user_id_study_id_name_idx
  ON user_asset (user_id, study_id, LOWER(name));
CREATE INDEX user_asset_user_id_study_id_type_subtype_idx
  ON user_asset (user_id, study_id, type, subtype);
CREATE INDEX user_asset_user_id_study_id_created_at_idx
  ON user_asset (user_id, study_id, created_at);

CREATE TRIGGER user_asset_insert_name_tokens
  BEFORE INSERT ON user_asset
  FOR EACH ROW EXECUTE PROCEDURE insert_name_tokens();

CREATE TRIGGER user_asset_update_name_tokens
  BEFORE UPDATE ON user_asset
  FOR EACH ROW EXECUTE PROCEDURE update_name_tokens();

CREATE TRIGGER user_asset_updated_at_modtime
  BEFORE UPDATE ON user_asset
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
