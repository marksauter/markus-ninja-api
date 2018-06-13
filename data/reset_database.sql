CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language 'plpgsql';

DROP TABLE IF EXISTS account CASCADE;
CREATE TABLE account(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  login         VARCHAR(40)  NOT NULL,
  name          TEXT         CHECK(name ~ '^[\w|-][\w|-|\s]+[\w|-]$'),
  password      BYTEA        NOT NULL,
  profile       TEXT,
  updated_at    TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX account_unique__login__idx
  ON account (LOWER(login));

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

CREATE UNIQUE INDEX email_unique__value__idx
  ON email (LOWER(value));
CREATE INDEX email_user_id_idx ON email (user_id);
CREATE UNIQUE INDEX email_unique__user_id_type__idx
  ON email (user_id, type)
  WHERE type = ANY('{"PRIMARY", "BACKUP"}');
CREATE UNIQUE INDEX email_unique__user_id_public__idx
  ON email (user_id, public)
  WHERE public = TRUE;

CREATE TRIGGER email_updated_at_modtime
  BEFORE UPDATE ON email
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

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
  'Topic',
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

CREATE UNIQUE INDEX permission__access_level_type_field__key
  ON permission (access_level, type, field)
  WHERE field IS NOT NULL;
CREATE UNIQUE INDEX permission__access_level_type__key
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
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  token         VARCHAR(40),
  user_id       VARCHAR(100)  NOT NULL,
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
  email_id      VARCHAR(100)  NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  request_ip    INET          NOT NULL,
  token         VARCHAR(40),
  user_id       VARCHAR(100),
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
  name          TEXT          NOT NULL CHECK (name ~ '[\w|-]+'),
  name_tokens   TEXT          NOT NULL,
  updated_at    TIMESTAMPTZ   DEFAULT NOW(),
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX study_unique__user_id_name__key
  ON study (user_id, LOWER(name));
CREATE INDEX study_user_id_advanced_at_idx
  ON study (user_id, advanced_at);
CREATE INDEX study_user_id_updated_at_idx
  ON study (user_id, updated_at);

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
  created_at      TIMESTAMPTZ  DEFAULT NOW(),
  body            TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  number          INT          CHECK(number > 0),
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,    
  title           TEXT         NOT NULL,
  title_tokens    TEXT         NOT NULL,
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
  color       TEXT         NOT NULL,
  created_at  TIMESTAMPTZ  DEFAULT NOW(),
  is_default  BOOLEAN      DEFAULT FALSE,
  description TEXT,
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40)  NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT NOW()
); 

CREATE UNIQUE INDEX label_unique__name__idx
  ON label (LOWER(name));

CREATE TRIGGER label_updated_at_modtime
  BEFORE UPDATE
  ON label
  FOR EACH ROW
  EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS lesson_label CASCADE;
CREATE TABLE lesson_label(
  created_at TIMESTAMPTZ  DEFAULT NOW(),
  label_id   VARCHAR(100),
  lesson_id  VARCHAR(100),
  study_id   VARCHAR(100),
  PRIMARY KEY (study_id, lesson_id, label_id),
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DROP TABLE IF EXISTS topic CASCADE;
CREATE TABLE topic(
  created_at  TIMESTAMPTZ  DEFAULT NOW(),
  description TEXT,
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40)  NOT NULL CHECK(name ~ '^[a-zA-Z0-9][a-zA-Z0-9|-]+[a-zA-Z0-9]$'),
  name_tokens TEXT         NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX topic_unique__name__idx
  ON topic (LOWER(name));

CREATE TRIGGER topic_updated_at_modtime
  BEFORE UPDATE ON topic
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

DROP TABLE IF EXISTS study_topic CASCADE;
CREATE TABLE study_topic(
  related_at TIMESTAMPTZ  DEFAULT NOW(),
  study_id   VARCHAR(100),
  topic_id   VARCHAR(100),
  PRIMARY KEY (study_id, topic_id),
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topic_id)
    REFERENCES topic (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DROP TABLE IF EXISTS user_asset CASCADE;
CREATE TABLE user_asset(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  key           TEXT         NOT NULL,
  name          TEXT         NOT NULL CHECK(name ~ '[\w|-]+'),
  name_tokens   TEXT         NOT NULL,
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

CREATE UNIQUE INDEX user_asset_unique__user_id_study_id_name__idx
  ON user_asset (user_id, study_id, LOWER(name));
CREATE INDEX user_asset_user_id_study_id_type_subtype_idx
  ON user_asset (user_id, study_id, type, subtype);
CREATE INDEX user_asset_user_id_study_id_created_at_idx
  ON user_asset (user_id, study_id, created_at);

CREATE TRIGGER user_asset_updated_at_modtime
  BEFORE UPDATE ON user_asset
  FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE VIEW user_master AS
SELECT
  account.created_at,
  account.id,
  account.login,
  account.name,
  account.profile,
  email.value public_email,
  account.updated_at
FROM account
LEFT JOIN email ON email.user_id = account.id
  AND email.public = TRUE;

CREATE VIEW user_credentials AS
SELECT
  backup_email.value backup_email,
  account.id,
  account.login,
  account.password,
  primary_email.value primary_email,
  ARRAY(
    SELECT role.name
    FROM role
    LEFT JOIN user_role ON user_role.user_id = account.id
    WHERE role.id = user_role.role_id
  ) roles
FROM account
JOIN email primary_email ON primary_email.user_id = account.id
  AND primary_email.type = 'PRIMARY'
LEFT JOIN email backup_email ON backup_email.user_id = account.id
  AND backup_email.type = 'BACKUP';

CREATE MATERIALIZED VIEW user_search_index AS
SELECT
  *,
  setweight(to_tsvector('simple', login), 'A') ||
  setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
  setweight(to_tsvector('simple', coalesce(profile, '')), 'B') ||
  setweight(to_tsvector('simple', coalesce(public_email, '')), 'B') as document
FROM user_master;

CREATE UNIQUE INDEX user_search_index_id_unique_idx
  ON user_search_index (id);

CREATE INDEX user_search_index_fts_idx
  ON user_search_index USING gin(document);

CREATE VIEW email_master AS
SELECT
  email.created_at,
  email.id,
  email.public, 
  email.type,
  email.user_id,
  account.login user_login,
  email.updated_at,
  email.value,
  email.verified_at
FROM email
JOIN account ON account.id = email.user_id;

CREATE VIEW study_master AS
SELECT
  study.advanced_at,
  study.created_at,
  study.description,
  study.id,
  study.name,
  study.updated_at,
  study.user_id,
  account.login user_login
FROM study
JOIN account ON account.id = study.user_id;

CREATE MATERIALIZED VIEW study_search_index AS
SELECT
  study.advanced_at,
  study.created_at,
  study.description,
  study.id,
  study.name,
  study.updated_at,
  study.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', study.name_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(study.description, '')), 'B') ||
  setweight(to_tsvector('simple', account.login), 'C') ||
  setweight(to_tsvector('simple', coalesce(string_agg(topic.name, ' '), '')), 'A') as document
FROM study
JOIN account ON account.id = study.user_id
LEFT JOIN study_topic ON study_topic.study_id = study.id
LEFT JOIN topic ON topic.id = study_topic.topic_id
GROUP BY study.id, account.id;

CREATE UNIQUE INDEX study_search_index_id_unique_idx
  ON study_search_index (id);

CREATE INDEX study_search_index_fts_idx
  ON study_search_index USING gin(document);

CREATE VIEW lesson_master AS
SELECT
  lesson.body,
  lesson.created_at,
  lesson.id,
  lesson.number,
  lesson.published_at,
  lesson.study_id,
  study.name study_name,
  lesson.title,
  lesson.updated_at,
  lesson.user_id,
  account.login user_login
FROM lesson
JOIN study ON study.id = lesson.study_id
JOIN account ON account.id = lesson.user_id;

CREATE MATERIALIZED VIEW lesson_search_index AS
SELECT
  lesson.body,
  lesson.created_at,
  lesson.id,
  lesson.number,
  lesson.published_at,
  lesson.study_id,
  study.name study_name,
  lesson.title,
  lesson.updated_at,
  lesson.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', lesson.title_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(lesson.body, '')), 'B') ||
  setweight(to_tsvector('simple', study.name_tokens), 'C') ||
  setweight(to_tsvector('simple', account.login), 'C') ||
  setweight(to_tsvector('simple', coalesce(string_agg(label.name, ' '), '')), 'A') as document
FROM lesson
JOIN study ON study.id = lesson.study_id
JOIN account ON account.id = lesson.user_id
LEFT JOIN lesson_label ON lesson_label.lesson_id = lesson.id
LEFT JOIN label ON label.id = lesson_label.label_id
GROUP BY lesson.id, study.id, account.id;

CREATE UNIQUE INDEX lesson_search_index_id_unique_idx
  ON lesson_search_index (id);

CREATE INDEX lesson_search_index_fts_idx
  ON lesson_search_index USING gin(document);

CREATE VIEW lesson_comment_master AS
SELECT
  lesson_comment.body,
  lesson_comment.created_at,
  lesson_comment.id,
  lesson_comment.lesson_id,
  lesson.number lesson_number,
  lesson_comment.published_at,
  lesson_comment.study_id,
  study.name study_name,
  lesson_comment.updated_at,
  lesson_comment.user_id,
  account.login user_login
FROM lesson_comment
JOIN lesson ON lesson.id = lesson_comment.lesson_id
JOIN study ON study.id = lesson_comment.study_id
JOIN account ON account.id = lesson_comment.user_id;

CREATE VIEW topic_master AS
SELECT
  created_at,
  description,
  id,
  name,
  updated_at
FROM topic;

CREATE VIEW study_topic_master AS
SELECT
  topic.created_at,
  topic.description,
  study_topic.topic_id id,
  topic.name,
  study_topic.related_at,
  study_topic.study_id,
  topic.updated_at
FROM study_topic
JOIN topic ON topic.id = study_topic.topic_id;

CREATE MATERIALIZED VIEW topic_search_index AS
SELECT
  created_at,
  description,
  id,
  name,
  updated_at,
  setweight(to_tsvector('simple', name_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(description, '')), 'B') as document
FROM topic;

CREATE UNIQUE INDEX topic_search_index_id_unique_idx
  ON topic_search_index (id);

CREATE INDEX topic_search_index_fts_idx
  ON topic_search_index USING gin(document);

CREATE VIEW user_asset_master AS
SELECT
  user_asset.created_at,
  user_asset.id,
  user_asset.key,
  user_asset.name,
  user_asset.original_name,
  user_asset.published_at,
  user_asset.size,
  user_asset.study_id,
  study.name study_name,
  user_asset.subtype,
  user_asset.type,
  user_asset.updated_at,
  user_asset.user_id,
  account.login user_login
FROM user_asset
JOIN study ON study.id = user_asset.study_id
JOIN account ON account.id = user_asset.user_id;

CREATE MATERIALIZED VIEW user_asset_search_index AS
SELECT
  user_asset.created_at,
  user_asset.id,
  user_asset.key,
  user_asset.name,
  user_asset.original_name,
  user_asset.published_at,
  user_asset.size,
  user_asset.study_id,
  study.name study_name,
  user_asset.subtype,
  user_asset.type,
  user_asset.updated_at,
  user_asset.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', user_asset.name_tokens), 'A') ||
  setweight(to_tsvector('simple', user_asset.type), 'A') ||
  setweight(to_tsvector('simple', user_asset.subtype), 'C') ||
  setweight(to_tsvector('simple', study.name_tokens), 'C') ||
  setweight(to_tsvector('simple', account.login), 'C') AS document
FROM user_asset
JOIN study ON study.id = user_asset.study_id
JOIN account ON account.id = user_asset.user_id
GROUP BY user_asset.id, study.id, account.id;

CREATE UNIQUE INDEX user_asset_search_index_id_unique_idx
  ON user_asset_search_index (id);

CREATE INDEX user_asset_search_index_fts_idx
  ON user_asset_search_index USING gin(document);
