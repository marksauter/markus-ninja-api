CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

DROP TABLE IF EXISTS account CASCADE;
CREATE TABLE account(
  id            VARCHAR(40) PRIMARY KEY,
  login         VARCHAR(40) NOT NULL UNIQUE,
  primary_email VARCHAR(40) NOT NULL UNIQUE,
  password      BYTEA NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  bio           TEXT,
  email         TEXT,
  name          TEXT
);

CREATE TRIGGER account_updated_at_modtime
BEFORE UPDATE ON account
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

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

DROP TABLE IF EXISTS account_verification_token;
CREATE TABLE account_verification_token(
  token         VARCHAR(40)   PRIMARY KEY,
  user_id       VARCHAR(40)   NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  ended_at      TIMESTAMPTZ,
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

DROP TABLE IF EXISTS lesson CASCADE;
CREATE TABLE lesson(
  id              VARCHAR(40) PRIMARY KEY,
  study_id        VARCHAR(40) NOT NULL,    
  user_id         VARCHAR(40) NOT NULL,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  last_edited_at  TIMESTAMPTZ,
  published_at    TIMESTAMPTZ,
  body            TEXT,
  number          INT,
  title           TEXT,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

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
