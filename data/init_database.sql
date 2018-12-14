CREATE SCHEMA IF NOT EXISTS public;

CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE IF NOT EXISTS account(
  created_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  id            VARCHAR(100) PRIMARY KEY,
  login         VARCHAR(40)  NOT NULL CHECK(login !~ '(^-|--|-$)' AND login ~ '^[a-zA-Z0-9-]{1,39}$'),
  password      BYTEA        NOT NULL,
  updated_at    TIMESTAMPTZ  DEFAULT statement_timestamp()
);

CREATE UNIQUE INDEX IF NOT EXISTS account_unique_login_idx
  ON account (lower(login));

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'email_type') THEN
    CREATE TYPE email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY');
  END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS email(
  created_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  id          VARCHAR(100) PRIMARY KEY,
  public      BOOLEAN      DEFAULT FALSE,
  type        email_type   DEFAULT 'EXTRA',
  updated_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  user_id     VARCHAR(100) NOT NULL,
  value       VARCHAR(40)  NOT NULL,
  verified_at TIMESTAMPTZ,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS email_unique_value_idx
  ON email (lower(value));
CREATE INDEX IF NOT EXISTS email_user_id_idx ON email (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS email_unique_user_id_type_idx
  ON email (user_id, type)
  WHERE type = ANY('{"PRIMARY", "BACKUP"}');

CREATE OR REPLACE FUNCTION email_will_update()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'email'
    AND trigger_name = 'before_email_update'
) THEN
  CREATE TRIGGER before_email_update
    BEFORE INSERT ON email
    FOR EACH ROW EXECUTE PROCEDURE email_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_profile(
  bio           TEXT,
  email_id      VARCHAR(100),
  name          TEXT,
  updated_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  user_id       VARCHAR(100) PRIMARY KEY,
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION check_user_profile_email_id(_user_id VARCHAR, _email_id VARCHAR)
  RETURNS VOID
  LANGUAGE plpgsql
AS $$
  DECLARE e RECORD;
  BEGIN
    IF _email_id IS NOT NULL THEN
      SELECT 
        user_id,
        verified_at
      INTO STRICT e
      FROM email
      WHERE id = _email_id;
      IF NOT FOUND THEN
        RAISE EXCEPTION 'email with id `%` not found', _email_id;
      ELSIF e.user_id != _user_id THEN
        RAISE EXCEPTION 'cannot set user_profile.email_id to an email not owned by the user';
      ELSIF e.verified_at IS NULL THEN
        RAISE EXCEPTION 'cannot set user_profile.email_id to an unverified email';
      END IF;
    END IF;
    RETURN;
  END;
$$;

CREATE OR REPLACE FUNCTION user_profile_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM check_user_profile_email_id(NEW.user_id, NEW.email_id);

  IF NEW.email_id IS DISTINCT FROM OLD.email_id THEN
    IF OLD.email_id IS DISTINCT FROM NULL THEN
      UPDATE email
      SET public = false
      WHERE id = OLD.email_id;
    END IF;

    IF NEW.email_id IS DISTINCT FROM NULL THEN
      UPDATE email
      SET public = true
      WHERE id = NEW.email_id;
    END IF;
  END IF;

  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_profile'
    AND trigger_name = 'before_user_profile_update'
) THEN
  CREATE TRIGGER before_user_profile_update
    BEFORE UPDATE ON user_profile
    FOR EACH ROW EXECUTE PROCEDURE user_profile_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS role(
  created_at  TIMESTAMPTZ DEFAULT statement_timestamp(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS role_unique_upper_name_idx
  ON role (upper(name));

INSERT INTO role(name, description)
VALUES
  ('ADMIN', 'Grants administrative permissions.'),
  ('MEMBER', 'Grants additional permission to users with a membership.'),
  ('OWNER', 'Grants additional permissions to objects owned by the user.'),
  ('USER',  'Grants general user permissions.')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_role(
  created_at  TIMESTAMPTZ   DEFAULT statement_timestamp(),
  role        VARCHAR(40),
  user_id     VARCHAR(100),
  PRIMARY KEY (user_id, role),
  FOREIGN KEY (role)
    REFERENCES role (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE VIEW user_role_master AS
SELECT
  r.created_at,
  r.description,
  ur.created_at AS granted_at,
  r.name,
  ur.user_id
FROM role r
JOIN user_role ur ON ur.role = r.name;


CREATE OR REPLACE VIEW user_master AS
SELECT
  account.updated_at account_updated_at,
  user_profile.bio,
  account.created_at,
  account.id,
  account.login,
  user_profile.name,
  user_profile.updated_at profile_updated_at,
  user_profile.email_id profile_email_id,
  ARRAY(
    SELECT role.name
    FROM role
    LEFT JOIN user_role ON user_role.user_id = account.id
    WHERE role.name = user_role.role
  ) roles,
  CASE 
    WHEN email.verified_at IS NOT NULL THEN
      true
    ELSE false
  END AS verified
FROM account
JOIN email ON email.user_id = account.id
  AND email.type = 'PRIMARY'
JOIN user_profile ON user_profile.user_id = account.id;

CREATE OR REPLACE VIEW user_credentials AS
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
    WHERE role.name = user_role.role
  ) roles,
  CASE 
    WHEN primary_email.verified_at IS NOT NULL THEN
      true
    ELSE false
  END AS verified
FROM account
JOIN email primary_email ON primary_email.user_id = account.id
  AND primary_email.type = 'PRIMARY'
LEFT JOIN email backup_email ON backup_email.user_id = account.id
  AND backup_email.type = 'BACKUP';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'access_level') THEN
    CREATE TYPE access_level AS ENUM(
      'Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete'
    );
  END IF;
END
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audience') THEN
    CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS permission(
  access_level access_level NOT NULL,
  audience     audience     NOT NULL,
  created_at   TIMESTAMPTZ  DEFAULT statement_timestamp(),
  field        TEXT,
  id           VARCHAR(100) PRIMARY KEY,
  type         TEXT         NOT NULL,
  updated_at   TIMESTAMPTZ  DEFAULT statement_timestamp()
);

CREATE UNIQUE INDEX IF NOT EXISTS permission_access_level_type_field_key
  ON permission (access_level, type, field)
  WHERE field IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS permission_access_level_type_key
  ON permission (access_level, type)
  WHERE field IS NULL;

CREATE OR REPLACE FUNCTION permission_will_update()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'permission'
    AND trigger_name = 'before_permission_update'
) THEN
  CREATE TRIGGER before_permission_update
    BEFORE INSERT ON permission
    FOR EACH ROW EXECUTE PROCEDURE permission_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS role_permission(
  created_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  permission_id VARCHAR(100),
  role          VARCHAR(40),
  PRIMARY KEY (role, permission_id),
  FOREIGN KEY (role)
    REFERENCES role (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (permission_id)
    REFERENCES permission (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE VIEW role_permission_master AS
SELECT
  permission.access_level,
  permission.created_at,
  permission.field,
  role_permission.created_at granted_at,
  permission.id,
  role_permission.role,
  permission.type,
  permission.updated_at
FROM permission
JOIN role_permission ON role_permission.permission_id = permission.id;

CREATE TABLE IF NOT EXISTS email_verification_token(
  email_id      VARCHAR(100),
  expires_at    TIMESTAMPTZ   DEFAULT (statement_timestamp() + interval '20 minutes'),
  issued_at     TIMESTAMPTZ   DEFAULT statement_timestamp(),
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

CREATE INDEX IF NOT EXISTS email_verification_token_user_id_idx
  ON email_verification_token (user_id); 

CREATE TABLE IF NOT EXISTS password_reset_token(
  email_id      VARCHAR(100)  NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT statement_timestamp(),
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  expires_at    TIMESTAMPTZ   DEFAULT (statement_timestamp() + interval '20 minutes'),
  request_ip    INET          NOT NULL,
  token         VARCHAR(40),
  user_id       VARCHAR(100),
  PRIMARY KEY (user_id, token),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS study(
  advanced_at   TIMESTAMPTZ,
  created_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  description   TEXT,
  id            VARCHAR(100)  PRIMARY KEY,
  name          VARCHAR(40)   NOT NULL CHECK (name ~ '[\w-]{1,39}'),
  name_tokens   TEXT          NOT NULL,
  private       BOOLEAN       DEFAULT FALSE,
  updated_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS study_unique_user_id_name_key
  ON study (user_id, lower(name));
CREATE INDEX IF NOT EXISTS study_user_id_advanced_at_idx
  ON study (user_id, advanced_at);
CREATE INDEX IF NOT EXISTS study_user_id_updated_at_idx
  ON study (user_id, updated_at);

CREATE OR REPLACE FUNCTION study_will_update()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study'
    AND trigger_name = 'before_study_update'
) THEN
  CREATE TRIGGER before_study_update
    BEFORE INSERT ON study
    FOR EACH ROW EXECUTE PROCEDURE study_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION advance_study(_study_id VARCHAR)
  RETURNS VOID
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE study
  SET advanced_at = statement_timestamp()
  WHERE study.id = _study_id;
$$;

CREATE TABLE IF NOT EXISTS lesson(
  body            TEXT,
  created_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  draft           TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  last_edited_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  number          INT          NOT NULL CHECK(number > 0),
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,    
  title           TEXT         NOT NULL,
  title_tokens    TEXT         NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS lesson_study_id_number_idx
  ON lesson (study_id, number);
CREATE INDEX IF NOT EXISTS lesson_user_id_idx
  ON lesson (user_id);

CREATE OR REPLACE FUNCTION lesson_will_insert()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  SELECT INTO NEW.number count(*)::INT
  FROM lesson
  WHERE study_id = NEW.study_id;
  NEW.number = NEW.number + 1;
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'before_lesson_insert'
) THEN
  CREATE TRIGGER before_lesson_insert
    BEFORE INSERT ON lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_will_insert(); 
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_draft_backup(
  created_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  draft           TEXT,
  id              SERIAL,
  lesson_id       VARCHAR(100) NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  PRIMARY KEY (lesson_id, id),
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_draft_backup_lesson_id_updated_at_idx
  ON lesson_draft_backup (lesson_id, updated_at);

CREATE OR REPLACE FUNCTION lesson_draft_backup_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_draft_backup'
    AND trigger_name = 'before_lesson_draft_backup_update'
) THEN
  CREATE TRIGGER before_lesson_draft_backup_update
    BEFORE UPDATE ON lesson_draft_backup
    FOR EACH ROW EXECUTE PROCEDURE lesson_draft_backup_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION lesson_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
DECLARE 
  backup RECORD;
  new_backup_id INT;
BEGIN
  IF NEW.draft != OLD.draft THEN
    NEW.last_edited_at = statement_timestamp();

    SELECT id, updated_at INTO backup
    FROM lesson_draft_backup
    ORDER BY updated_at DESC
    LIMIT 1;

    IF FOUND THEN
      new_backup_id = backup.id % 5 + 1;

      IF age(statement_timestamp(), backup.updated_at) > INTERVAL '2 minute' THEN
        INSERT INTO lesson_draft_backup(draft, id, lesson_id)
        VALUES(OLD.draft, new_backup_id, NEW.id) 
        ON CONFLICT (lesson_id, id) DO 
          UPDATE SET draft = OLD.draft 
          WHERE lesson_draft_backup.lesson_id = NEW.id AND lesson_draft_backup.id = new_backup_id;
      END IF;
    ELSE
      INSERT INTO lesson_draft_backup(draft, lesson_id)
      VALUES(OLD.draft, NEW.id); 
    END IF;
  END IF;

  NEW.updated_at = statement_timestamp();

  RETURN NEW;
END
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'before_lesson_update'
) THEN
  CREATE TRIGGER before_lesson_update
    BEFORE UPDATE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_will_update();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'course_status') THEN
    CREATE TYPE course_status AS ENUM('ADVANCING', 'COMPLETED');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS course(
  advanced_at   TIMESTAMPTZ,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  description   TEXT,
  id            VARCHAR(100)  PRIMARY KEY,
  name          VARCHAR(40)   NOT NULL,
  name_tokens   TEXT          NOT NULL,
  number        INT           CHECK(number > 0),
  published_at  TIMESTAMPTZ,
  status        course_status DEFAULT 'ADVANCING',
  study_id      VARCHAR(100)  NOT NULL,
  updated_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS course_unique_study_id_name_key
  ON course (study_id, lower(name));
CREATE UNIQUE INDEX IF NOT EXISTS course_study_id_number_idx
  ON course (study_id, number);
CREATE INDEX IF NOT EXISTS course_study_id_created_at_idx
  ON course (study_id, created_at);
CREATE INDEX IF NOT EXISTS course_study_id_advanced_at_idx
  ON course (study_id, advanced_at);
CREATE INDEX IF NOT EXISTS course_user_id_idx
  ON course (user_id);

CREATE OR REPLACE FUNCTION course_will_insert()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  SELECT INTO NEW.number count(*)::INT
  FROM course
  WHERE study_id = NEW.study_id;
  NEW.number = NEW.number + 1;
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course'
    AND trigger_name = 'before_course_insert'
) THEN
  CREATE TRIGGER before_course_insert
    BEFORE INSERT ON course
    FOR EACH ROW EXECUTE PROCEDURE course_will_insert(); 
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course'
    AND trigger_name = 'before_course_update'
) THEN
  CREATE TRIGGER before_course_update
    BEFORE UPDATE ON course
    FOR EACH ROW EXECUTE PROCEDURE course_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION advance_course(_course_id VARCHAR)
  RETURNS VOID
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE course
  SET advanced_at = statement_timestamp()
  WHERE course.id = _course_id;
$$;

CREATE TABLE IF NOT EXISTS course_lesson(
  created_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  course_id       VARCHAR(100) NOT NULL,
  lesson_id       VARCHAR(100) PRIMARY KEY,
  number          INT          NOT NULL CHECK(number > 0),
  UNIQUE (course_id, number) DEFERRABLE INITIALLY DEFERRED,
  FOREIGN KEY (course_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION course_lesson_will_insert()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  SELECT INTO NEW.number count(*)::INT
  FROM course_lesson
  WHERE course_id = NEW.course_id;
  NEW.number = NEW.number + 1;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_lesson'
    AND trigger_name = 'before_course_lesson_insert'
) THEN
  CREATE TRIGGER before_course_lesson_insert
    BEFORE INSERT ON course_lesson
    FOR EACH ROW EXECUTE PROCEDURE course_lesson_will_insert(); 
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE VIEW lesson_master AS
SELECT
  lesson.body,
  lesson.created_at,
  course_lesson.course_id,
  course_lesson.number course_number,
  lesson.draft,
  lesson.id,
  lesson.last_edited_at,
  lesson.number,
  lesson.published_at,
  lesson.study_id,
  lesson.title,
  lesson.updated_at,
  lesson.user_id
FROM lesson
LEFT JOIN course_lesson ON course_lesson.lesson_id = lesson.id;

CREATE TABLE IF NOT EXISTS lesson_comment(
  body            TEXT,
  created_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  draft           TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  last_edited_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  lesson_id       VARCHAR(100) NOT NULL,
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,
  user_id         VARCHAR(100) NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
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

CREATE UNIQUE INDEX IF NOT EXISTS lesson_comment_user_id_lesson_id_null_published_at_unique_idx
  ON lesson_comment (user_id, lesson_id)
  WHERE published_at IS NULL;
CREATE INDEX IF NOT EXISTS lesson_comment_user_id_idx
  ON lesson_comment (user_id);
CREATE INDEX IF NOT EXISTS lesson_comment_study_id_idx
  ON lesson_comment (study_id);
CREATE INDEX IF NOT EXISTS lesson_comment_lesson_id_published_at_idx
  ON lesson_comment (lesson_id, published_at DESC NULLS LAST);

CREATE OR REPLACE FUNCTION lesson_comment_will_insert()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.study_id IS NULL THEN
    SELECT study.id
    INTO NEW.study_id
    FROM study
    JOIN lesson ON lesson.id = NEW.lesson_id
    WHERE study.id = lesson.study_id;
  END IF;
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'before_lesson_comment_insert'
) THEN
  CREATE TRIGGER before_lesson_comment_insert
    BEFORE INSERT ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_will_insert();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_comment_draft_backup(
  created_at        TIMESTAMPTZ  DEFAULT statement_timestamp(),
  draft             TEXT,
  id                SERIAL,
  lesson_comment_id VARCHAR(100) NOT NULL,
  updated_at        TIMESTAMPTZ  DEFAULT statement_timestamp(),
  PRIMARY KEY (lesson_comment_id, id),
  FOREIGN KEY (lesson_comment_id)
    REFERENCES lesson_comment (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION lesson_comment_draft_backup_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment_draft_backup'
    AND trigger_name = 'before_lesson_comment_draft_backup_update'
) THEN
  CREATE TRIGGER before_lesson_comment_draft_backup_update
    BEFORE UPDATE ON lesson_comment_draft_backup
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_draft_backup_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION lesson_comment_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
DECLARE 
  backup RECORD;
  new_backup_id INT;
BEGIN
  IF NEW.draft != OLD.draft THEN
    NEW.last_edited_at = statement_timestamp();

    SELECT id, updated_at INTO backup
    FROM lesson_comment_draft_backup
    ORDER BY updated_at DESC
    LIMIT 1;

    IF FOUND THEN
      new_backup_id = backup.id % 5 + 1;

      IF age(statement_timestamp(), backup.updated_at) > INTERVAL '2 minute' THEN
        INSERT INTO lesson_comment_draft_backup(draft, id, lesson_comment_id)
        VALUES(OLD.draft, new_backup_id, NEW.id) 
        ON CONFLICT (lesson_comment_id, id) DO 
          UPDATE SET draft = OLD.draft 
          WHERE lesson_comment_draft_backup.lesson_comment_id = NEW.id AND lesson_comment_draft_backup.id = new_backup_id;
      END IF;
    ELSE
      INSERT INTO lesson_comment_draft_backup(draft, lesson_comment_id)
      VALUES(OLD.draft, NEW.id); 
    END IF;
  END IF;

  NEW.updated_at = statement_timestamp();

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'before_lesson_comment_update'
) THEN
  CREATE TRIGGER before_lesson_comment_update
    BEFORE UPDATE ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS label(
  color       TEXT         NOT NULL,
  created_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  description TEXT,
  document    TSVECTOR     NOT NULL,
  id          VARCHAR(100) PRIMARY KEY,
  is_default  BOOLEAN      DEFAULT FALSE,
  name        VARCHAR(40)  NOT NULL CHECK (name ~ '[\w-]{1,39}'),
  name_tokens TEXT         NOT NULL,
  study_id    VARCHAR(100) NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT statement_timestamp()
); 

CREATE UNIQUE INDEX IF NOT EXISTS label_unique_study_id_name_idx
  ON label (study_id, lower(name));

CREATE INDEX IF NOT EXISTS label_fts_idx
  ON label USING gin(document);

CREATE OR REPLACE FUNCTION label_will_insert()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.document = setweight(to_tsvector('simple', coalesce(NEW.name_tokens, '')), 'A');

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'label'
    AND trigger_name = 'before_label_insert'
) THEN
  CREATE TRIGGER before_label_insert
    BEFORE INSERT ON label
    FOR EACH ROW EXECUTE PROCEDURE label_will_insert();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION label_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'label'
    AND trigger_name = 'before_label_update'
) THEN
  CREATE TRIGGER before_label_update
    BEFORE UPDATE ON label
    FOR EACH ROW EXECUTE PROCEDURE label_will_update();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'labelable_type') THEN
    CREATE TYPE labelable_type AS ENUM('Lesson', 'LessonComment');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS labeled(
  created_at    TIMESTAMPTZ     DEFAULT statement_timestamp(),
  id            SERIAL          PRIMARY KEY,
  label_id      VARCHAR(100)    NOT NULL,
  labelable_id  VARCHAR(100)    NOT NULL,
  type          labelable_type  NOT NULL,
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS labeled_unique_labelable_id_label_id_idx
  ON labeled (labelable_id, label_id);

CREATE INDEX IF NOT EXISTS labeled_label_id_type_labelable_id_idx
  ON labeled (label_id, type, labelable_id);

CREATE TABLE IF NOT EXISTS lesson_labeled(
  created_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  label_id      VARCHAR(100),
  labelable_id  VARCHAR(100),
  labeled_id    INT          PRIMARY KEY,
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (labeled_id)
    REFERENCES labeled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (labelable_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_labeled_label_id_labelable_id_idx
  ON lesson_labeled (label_id, labelable_id);

CREATE TABLE IF NOT EXISTS lesson_comment_labeled(
  created_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  label_id      VARCHAR(100),
  labelable_id  VARCHAR(100),
  labeled_id    INT          PRIMARY KEY,
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (labeled_id)
    REFERENCES labeled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (labelable_id)
    REFERENCES lesson_comment (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_comment_labeled_label_id_labelable_id_idx
  ON lesson_comment_labeled (label_id, labelable_id);

CREATE OR REPLACE VIEW label_search_index AS
SELECT
  label.*,
  count(labeled) AS labeled_count
FROM label
LEFT JOIN labeled ON labeled.label_id = label.id
GROUP BY label.id;

CREATE OR REPLACE VIEW labelable_label AS
SELECT
  label_search_index.*,
  labeled.labelable_id,
  labeled.created_at labeled_at
FROM label_search_index
JOIN labeled ON labeled.label_id = label_search_index.id;

CREATE OR REPLACE VIEW labeled_lesson_comment AS
SELECT
  lesson_comment.*,
  lesson_comment_labeled.label_id,
  lesson_comment_labeled.created_at labeled_at
FROM lesson_comment
JOIN lesson_comment_labeled ON lesson_comment_labeled.labelable_id = lesson_comment.id;

CREATE TABLE IF NOT EXISTS topic(
  created_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  description TEXT,
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40)  NOT NULL CHECK(name ~ '^[a-zA-Z0-9-]{1,39}$'),
  name_tokens TEXT         NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT statement_timestamp()
);

CREATE UNIQUE INDEX IF NOT EXISTS topic_unique_name_idx
  ON topic (lower(name));

CREATE OR REPLACE FUNCTION topic_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topic'
    AND trigger_name = 'before_topic_update'
) THEN
  CREATE TRIGGER before_topic_update
    BEFORE UPDATE ON topic
    FOR EACH ROW EXECUTE PROCEDURE topic_will_update();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'topicable_type') THEN
    CREATE TYPE topicable_type AS ENUM('Course', 'Study');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS topiced(
  created_at    TIMESTAMPTZ     DEFAULT statement_timestamp(),
  id            SERIAL          PRIMARY KEY,
  topic_id      VARCHAR(100)    NOT NULL,
  topicable_id  VARCHAR(100)    NOT NULL,
  type          topicable_type  NOT NULL,
  FOREIGN KEY (topic_id)
    REFERENCES topic (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS topiced_unique_topicable_id_topic_id_idx
  ON topiced (topicable_id, topic_id);

CREATE INDEX IF NOT EXISTS topiced_topic_id_type_topicable_id_idx
  ON topiced (topic_id, type, topicable_id);

CREATE TABLE IF NOT EXISTS course_topiced(
  created_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  topic_id      VARCHAR(100),
  topicable_id  VARCHAR(100),
  topiced_id    INT          PRIMARY KEY,
  FOREIGN KEY (topic_id)
    REFERENCES topic (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topiced_id)
    REFERENCES topiced (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topicable_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS course_topiced_topic_id_topicable_id_idx
  ON course_topiced (topic_id, topicable_id);

CREATE TABLE IF NOT EXISTS study_topiced(
  created_at    TIMESTAMPTZ  DEFAULT statement_timestamp(),
  topic_id      VARCHAR(100),
  topicable_id  VARCHAR(100),
  topiced_id    INT          PRIMARY KEY,
  FOREIGN KEY (topic_id)
    REFERENCES topic (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topiced_id)
    REFERENCES topiced (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topicable_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS study_topiced_topic_id_topicable_id_idx
  ON study_topiced (topic_id, topicable_id);

CREATE TABLE IF NOT EXISTS asset(
  created_at TIMESTAMPTZ  DEFAULT statement_timestamp(),
  id         BIGSERIAL    PRIMARY KEY,
  key        VARCHAR(40)  NOT NULL,
  name       TEXT         NOT NULL, 
  size       BIGINT       NOT NULL,
  subtype    TEXT         NOT NULL,
  type       TEXT         NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS asset_key_unique_idx
  ON asset (key);

CREATE TABLE IF NOT EXISTS user_asset(
  asset_id     BIGINT       NOT NULL,
  description  TEXT,
  id           VARCHAR(100) PRIMARY KEY,
  name         VARCHAR(40)  NOT NULL CHECK(name ~ '^[\w\-.]{1,39}$'),
  name_tokens  TEXT         NOT NULL,
  study_id     VARCHAR(100) NOT NULL,
  updated_at   TIMESTAMPTZ  DEFAULT statement_timestamp(),
  user_id      VARCHAR(100) NOT NULL,
  FOREIGN KEY (asset_id)
    REFERENCES asset (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_study_id_name_unique_idx
  ON user_asset (study_id, lower(name));
CREATE INDEX IF NOT EXISTS user_asset_user_id_idx
  ON user_asset (user_id);

CREATE OR REPLACE FUNCTION user_asset_will_update()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = statement_timestamp();
  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset'
    AND trigger_name = 'before_user_asset_update'
) THEN
  CREATE TRIGGER before_user_asset_update
    BEFORE UPDATE ON user_asset
    FOR EACH ROW EXECUTE PROCEDURE user_asset_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE VIEW user_asset_master AS
SELECT
  asset.id asset_id,
  asset.created_at,
  user_asset.description,
  user_asset.id,
  asset.key,
  user_asset.name,
  user_asset.name_tokens,
  asset.name original_name,
  asset.size,
  user_asset.study_id,
  asset.subtype,
  asset.type,
  user_asset.updated_at,
  user_asset.user_id
FROM user_asset
JOIN asset ON asset.id = user_asset.asset_id;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'apple_type') THEN
    CREATE TYPE apple_type AS ENUM('Course', 'Study');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS appled(
  appleable_id  VARCHAR(100)  NOT NULL,
  created_at    TIMESTAMPTZ   DEFAULT statement_timestamp(),
  id            SERIAL        PRIMARY KEY,
  type          apple_type    NOT NULL,
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS appled_unique_appleable_id_user_id_idx
  ON appled (appleable_id, user_id);

CREATE INDEX IF NOT EXISTS appled_user_id_appleable_id_idx
  ON appled (user_id, appleable_id);

CREATE TABLE IF NOT EXISTS course_appled(
  appleable_id  VARCHAR(100) NOT NULL,
  appled_id     INT          PRIMARY KEY,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (appled_id)
    REFERENCES appled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (appleable_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS course_appled_user_id_appleable_id_idx
  ON course_appled (user_id, appleable_id);

CREATE TABLE IF NOT EXISTS study_appled(
  appleable_id  VARCHAR(100) NOT NULL,
  appled_id     INT          PRIMARY KEY,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (appled_id)
    REFERENCES appled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (appleable_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS study_appled_user_id_appleable_id_idx
  ON study_appled (user_id, appleable_id);

CREATE TABLE IF NOT EXISTS reason (
  created_at  TIMESTAMPTZ DEFAULT statement_timestamp(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS reason_unique_lower_name_idx
  ON reason (lower(name));

INSERT INTO reason (name, description)
VALUES
  ('author', 'You created the thread.'),
  ('comment', 'You commented on the thread.'),
  ('enrolled', E'You\'re enrolled in the study.'),
  ('manual', 'You enrolled in the thread.'),
  ('mention', 'You were @mentioned in the thread.')
ON CONFLICT (name) DO NOTHING;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enrollment_status') THEN
    CREATE TYPE enrollment_status AS ENUM('ENROLLED', 'IGNORED', 'UNENROLLED');
  END IF;
END
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enrollment_type') THEN
    CREATE TYPE enrollment_type AS ENUM('Lesson', 'Study', 'User');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS enrolled(
  created_at    TIMESTAMPTZ       DEFAULT statement_timestamp(),
  id            SERIAL            PRIMARY KEY,
  enrollable_id VARCHAR(100)      NOT NULL,
  reason_name   VARCHAR(40),
  status        enrollment_status NOT NULL DEFAULT 'ENROLLED',
  type          enrollment_type   NOT NULL,
  user_id       VARCHAR(100)      NOT NULL,
  FOREIGN KEY (reason_name)
    REFERENCES reason (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS enrolled_unique_enrollable_id_user_id_idx
  ON enrolled (enrollable_id, user_id);

CREATE INDEX IF NOT EXISTS enrolled_user_id_type_enrollable_id_idx
  ON enrolled (user_id, type, enrollable_id);

CREATE TABLE IF NOT EXISTS lesson_enrolled(
  enrollable_id VARCHAR(100) NOT NULL,
  enrolled_id   INT          PRIMARY KEY,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (enrolled_id)
    REFERENCES enrolled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (enrollable_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS lesson_enrolled_unique_enrollable_id_user_id_idx
  ON lesson_enrolled (enrollable_id, user_id);

CREATE TABLE IF NOT EXISTS study_enrolled(
  enrollable_id VARCHAR(100)      NOT NULL,
  enrolled_id   INT               PRIMARY KEY,
  user_id       VARCHAR(100)      NOT NULL,
  FOREIGN KEY (enrolled_id)
    REFERENCES enrolled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (enrollable_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS study_enrolled_unique_enrollable_id_user_id_idx
  ON study_enrolled (enrollable_id, user_id);

CREATE TABLE IF NOT EXISTS user_enrolled(
  enrollable_id VARCHAR(100)      NOT NULL,
  enrolled_id   INT               PRIMARY KEY,
  user_id       VARCHAR(100)      NOT NULL,
  FOREIGN KEY (enrolled_id)
    REFERENCES enrolled (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (enrollable_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_enrolled_unique_enrollable_id_user_id_idx
  ON user_enrolled (enrollable_id, user_id);

CREATE TABLE IF NOT EXISTS event_type (
  created_at  TIMESTAMPTZ DEFAULT statement_timestamp(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS event_type_name_key
  ON event_type (lower(name));

INSERT INTO event_type (name, description)
VALUES
  ('CourseEvent', 'Triggered when a course is created, appled, or unappled.'),
  ('LessonEvent', 'Triggered when a lesson is created, added to a course, 
    removed from a course, commented, labeled, unlabeled, referenced, or renamed. 
    Also triggered when a user is mentioned in the lesson body.'),
  ('PublicEvent', 'Triggered when a study is made public.'),
  ('UserAssetEvent', 'Triggered when a user asset is created.'),
  ('StudyEvent', 'Triggered when a study is created, appled, or unappled.')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS event (
  created_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  id          VARCHAR(100) PRIMARY KEY,
  payload     JSONB,
  public      BOOLEAN      DEFAULT true,
  study_id    VARCHAR(100) NOT NULL,
  type        VARCHAR(40)  NOT NULL,
  user_id     VARCHAR(100) NOT NULL,
  FOREIGN KEY (type)
    REFERENCES event_type (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS event_study_id_type_action_created_at_idx
  ON event (study_id, type, (payload->'action'), created_at);

CREATE INDEX IF NOT EXISTS event_user_id_type_action_created_at_idx
  ON event (user_id, type, (payload->'action'), created_at);

CREATE OR REPLACE FUNCTION event_will_insert()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
DECLARE ts TIMESTAMPTZ;
BEGIN
  CASE NEW.type
    WHEN 'LessonEvent' THEN
      CASE NEW.payload->>'action'
        WHEN 'added_to_course' THEN
          SELECT INTO ts created_at
          FROM lesson_event
          WHERE payload->>'course_id' = NEW.payload->>'course_id'
            AND payload->>'action' = 'added_to_course'
            AND payload->>'lesson_id' = NEW.payload->>'lesson_id';

          IF FOUND AND ts >= CURRENT_DATE AND ts < CURRENT_DATE + interval '1 day' THEN
            RETURN NULL;
          END IF;
        WHEN 'labeled' THEN
          SELECT INTO ts created_at
          FROM lesson_event
          WHERE payload->>'label_id' = NEW.payload->>'label_id'
            AND payload->>'action' = 'labeled'
            AND payload->>'lesson_id' = NEW.payload->>'lesson_id';

          IF FOUND AND ts >= CURRENT_DATE AND ts < CURRENT_DATE + interval '1 day' THEN
            RETURN NULL;
          END IF;
        WHEN 'removed_from_course' THEN
          SELECT INTO ts created_at
          FROM lesson_event
          WHERE payload->>'course_id' = NEW.payload->>'course_id'
            AND payload->>'action' = 'removed_from_course'
            AND payload->>'lesson_id' = NEW.payload->>'lesson_id';

          IF FOUND AND ts >= CURRENT_DATE AND ts < CURRENT_DATE + interval '1 day' THEN
            RETURN NULL;
          END IF;
        WHEN 'unlabeled' THEN
          SELECT INTO ts created_at
          FROM lesson_event
          WHERE payload->>'label_id' = NEW.payload->>'label_id'
            AND payload->>'action' = 'unlabeled'
            AND payload->>'lesson_id' = NEW.payload->>'lesson_id';

          IF FOUND AND ts >= CURRENT_DATE AND ts < CURRENT_DATE + interval '1 day' THEN
            RETURN NULL;
          END IF;
        ELSE
          RETURN NEW;
      END CASE;
    ELSE
      RETURN NEW;
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'event'
    AND trigger_name = 'before_event_insert'
) THEN
  CREATE TRIGGER before_event_insert
    BEFORE INSERT ON event
    FOR EACH ROW EXECUTE PROCEDURE event_will_insert();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION event_inserted()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  CASE NEW.type
    WHEN 'CourseEvent' THEN
      INSERT INTO course_event (action, course_id, event_id, study_id, user_id)
      VALUES (
        (NEW.payload->>'action')::course_event_action,
        NEW.payload->>'course_id',
        NEW.id,
        NEW.study_id,
        NEW.user_id
      );
    WHEN 'LessonEvent' THEN
      INSERT INTO lesson_event (action, event_id, lesson_id, payload, study_id, user_id)
      VALUES (
        NEW.payload->>'action',
        NEW.id,
        NEW.payload->>'lesson_id',
        NEW.payload,
        NEW.study_id,
        NEW.user_id
      )
      ON CONFLICT DO NOTHING;
      IF NOT FOUND THEN
        DELETE FROM event WHERE id = NEW.id;
        RETURN NULL;
      END IF;
    WHEN 'StudyEvent' THEN
      INSERT INTO study_event (action, event_id, study_id, user_id)
      VALUES (
        (NEW.payload->>'action')::study_event_action,
        NEW.id,
        NEW.study_id,
        NEW.user_id
      );
    WHEN 'UserAssetEvent' THEN
      INSERT INTO user_asset_event (action, asset_id, event_id, payload, study_id, user_id)
      VALUES (
        NEW.payload->>'action',
        NEW.payload->>'asset_id',
        NEW.id,
        NEW.payload,
        NEW.study_id,
        NEW.user_id
      )
      ON CONFLICT DO NOTHING;
      IF NOT FOUND THEN
        DELETE FROM event WHERE id = NEW.id;
        RETURN NULL;
      END IF;
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'event'
    AND trigger_name = 'after_event_insert'
) THEN
  CREATE TRIGGER after_event_insert
    AFTER INSERT ON event
    FOR EACH ROW EXECUTE PROCEDURE event_inserted();
END IF;
END;
$$ language 'plpgsql';


CREATE TABLE IF NOT EXISTS received_event (
  created_at TIMESTAMPTZ  DEFAULT statement_timestamp(),
  event_id   VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  PRIMARY KEY (event_id, user_id),
  FOREIGN KEY (event_id)
    REFERENCES event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS received_event_user_id_created_at_idx
  ON received_event (user_id, created_at);

CREATE OR REPLACE FUNCTION delete_event()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  DELETE FROM event
  WHERE id = OLD.event_id;
  RETURN OLD;
END;
$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'course_event_action') THEN
    CREATE TYPE course_event_action AS ENUM('created', 'appled', 'unappled', 'published');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS course_event (
  action     course_event_action NOT NULL,
  course_id  VARCHAR(100)        NOT NULL,
  event_id   VARCHAR(100)        PRIMARY KEY,
  study_id    VARCHAR(100)       NOT NULL,
  user_id    VARCHAR(100)        NOT NULL,
  FOREIGN KEY (course_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_event'
    AND trigger_name = 'after_course_event_delete'
) THEN
  CREATE TRIGGER after_course_event_delete
    AFTER DELETE ON course_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_event_inserted()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  CASE NEW.action
    WHEN 'appled' THEN
      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM user_enrolled
      WHERE enrollable_id = NEW.user_id;
    WHEN 'published' THEN
      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM study_enrolled
      WHERE enrollable_id = NEW.study_id AND user_id != NEW.user_id;
    ELSE
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_event'
    AND trigger_name = 'after_course_event_insert'
) THEN
  CREATE TRIGGER after_course_event_insert
    AFTER INSERT ON course_event
    FOR EACH ROW EXECUTE PROCEDURE course_event_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_event_action (
  created_at  TIMESTAMPTZ DEFAULT statement_timestamp(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS lesson_event_action_unique_lower_name_idx
  ON lesson_event_action (lower(name));

INSERT INTO lesson_event_action (name, description)
VALUES
  ('added_to_course', 'The lesson was added to a course.'),
  ('created', 'The lesson was created.'),
  ('commented', 'A comment was added to the lesson.'),
  ('labeled', 'A label was added to the lesson.'),
  ('mentioned', 'The user was @mentioned in a lesson body.'),
  ('published', 'The lesson was published.'),
  ('referenced', 'The lesson was referenced from another lesson.'),
  ('removed_from_course', 'The lesson was removed from a course'),
  ('renamed', 'The lesson title was changed.'),
  ('unlabeled', 'A label was removed from the lesson.')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS lesson_event (
  action     VARCHAR(40)  NOT NULL,
  created_at TIMESTAMPTZ  DEFAULT statement_timestamp(),
  event_id   VARCHAR(100) PRIMARY KEY,
  lesson_id  VARCHAR(100) NOT NULL,
  payload    JSONB        NOT NULL,
  study_id   VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (action)
    REFERENCES lesson_event_action (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
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

CREATE UNIQUE INDEX IF NOT EXISTS lesson_event_source_id_referenced_lesson_id_unique_idx
  ON lesson_event ((payload->'source_id'), action, lesson_id) WHERE action = 'referenced';

CREATE UNIQUE INDEX IF NOT EXISTS lesson_event_lesson_id_mentioned_lesson_id_unique_idx
  ON lesson_event (lesson_id, action, user_id) WHERE action = 'mentioned';

CREATE INDEX IF NOT EXISTS lesson_event_lesson_id_created_at
  ON lesson_event (lesson_id, created_at);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_event'
    AND trigger_name = 'after_lesson_event_delete'
) THEN
  CREATE TRIGGER after_lesson_event_delete
    AFTER DELETE ON lesson_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION lesson_event_inserted()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  CASE NEW.action
    WHEN 'added_to_course' THEN
      INSERT INTO lesson_added_to_course_event (course_id, event_id)
      VALUES (
        NEW.payload->>'course_id',
        NEW.event_id
      );
    WHEN 'created' THEN
      INSERT INTO enrolled(enrollable_id, reason_name, type, user_id)
      VALUES (NEW.lesson_id, 'author', 'Lesson', NEW.user_id)
      ON CONFLICT (enrollable_id, user_id) DO NOTHING;
    WHEN 'commented' THEN
      INSERT INTO lesson_commented_event (comment_id, event_id)
      VALUES (
        NEW.payload->>'comment_id',
        NEW.event_id
      );

      INSERT INTO enrolled(enrollable_id, reason_name, type, user_id)
      VALUES (NEW.lesson_id, 'comment', 'Lesson', NEW.user_id)
      ON CONFLICT (enrollable_id, user_id) DO UPDATE
        SET reason_name = 'comment';
    WHEN 'labeled' THEN
      INSERT INTO lesson_labeled_event (event_id, label_id)
      VALUES (
        NEW.event_id,
        NEW.payload->>'label_id'
      );
    WHEN 'mentioned' THEN
      INSERT INTO enrolled(enrollable_id, reason_name, type, user_id)
      VALUES (NEW.lesson_id, 'mention', 'Lesson', NEW.user_id)
      ON CONFLICT (enrollable_id, user_id) DO UPDATE
        SET reason_name = 'mention';
    WHEN 'published' THEN
      INSERT INTO enrolled(enrollable_id, reason_name, type, user_id)
      SELECT 
        NEW.lesson_id,
        'enrolled',
        'Lesson',
        user_id
      FROM study_enrolled
      WHERE enrollable_id = NEW.study_id
      ON CONFLICT (enrollable_id, user_id) DO NOTHING;

      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM study_enrolled
      WHERE enrollable_id = NEW.study_id AND user_id != NEW.user_id;
    WHEN 'referenced' THEN
      INSERT INTO lesson_referenced_event (event_id, source_id)
      VALUES (
        NEW.event_id,
        NEW.payload->>'source_id'
      );
    WHEN 'removed_from_course' THEN
      INSERT INTO lesson_removed_from_course_event (course_id, event_id)
      VALUES (
        NEW.payload->>'course_id',
        NEW.event_id
      );
    WHEN 'renamed' THEN
      INSERT INTO lesson_renamed_event (event_id, renamed_from, renamed_to)
      VALUES (
        NEW.event_id,
        NEW.payload->'rename'->>'from',
        NEW.payload->'rename'->>'to'
      );
    WHEN 'unlabeled' THEN
      INSERT INTO lesson_unlabeled_event (event_id, label_id)
      VALUES (
        NEW.event_id,
        NEW.payload->>'label_id'
      );
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_event'
    AND trigger_name = 'after_lesson_event_insert'
) THEN
  CREATE TRIGGER after_lesson_event_insert
    AFTER INSERT ON lesson_event
    FOR EACH ROW EXECUTE PROCEDURE lesson_event_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_added_to_course_event (
  course_id       VARCHAR(100) NOT NULL,
  event_id        VARCHAR(100) PRIMARY KEY,
  FOREIGN KEY (course_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_added_to_course_event'
    AND trigger_name = 'after_lesson_added_to_course_event_delete'
) THEN
  CREATE TRIGGER after_lesson_added_to_course_event_delete
    AFTER DELETE ON lesson_added_to_course_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_commented_event (
  comment_id      VARCHAR(100) NOT NULL,
  event_id        VARCHAR(100) PRIMARY KEY,
  FOREIGN KEY (comment_id)
    REFERENCES lesson_comment (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_commented_event'
    AND trigger_name = 'after_lesson_commented_event_delete'
) THEN
  CREATE TRIGGER after_lesson_commented_event_delete
    AFTER DELETE ON lesson_commented_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_labeled_event (
  event_id        VARCHAR(100) PRIMARY KEY,
  label_id        VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_labeled_event'
    AND trigger_name = 'after_lesson_labeled_event_delete'
) THEN
  CREATE TRIGGER after_lesson_labeled_event_delete
    AFTER DELETE ON lesson_labeled_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_referenced_event (
  event_id        VARCHAR(100) PRIMARY KEY,
  source_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_referenced_event'
    AND trigger_name = 'after_lesson_referenced_event_delete'
) THEN
  CREATE TRIGGER after_lesson_referenced_event_delete
    AFTER DELETE ON lesson_referenced_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_removed_from_course_event (
  course_id       VARCHAR(100) NOT NULL,
  event_id        VARCHAR(100) PRIMARY KEY,
  FOREIGN KEY (course_id)
    REFERENCES course (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_removed_from_course_event'
    AND trigger_name = 'after_lesson_removed_from_course_event_delete'
) THEN
  CREATE TRIGGER after_lesson_removed_from_course_event_delete
    AFTER DELETE ON lesson_removed_from_course_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_renamed_event (
  event_id        VARCHAR(100) PRIMARY KEY,
  renamed_from    TEXT         NOT NULL,
  renamed_to      TEXT         NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS lesson_unlabeled_event (
  event_id        VARCHAR(100) PRIMARY KEY,
  label_id        VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES lesson_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_unlabeled_event'
    AND trigger_name = 'after_lesson_unlabeled_event_delete'
) THEN
  CREATE TRIGGER after_lesson_unlabeled_event_delete
    AFTER DELETE ON lesson_unlabeled_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE VIEW lesson_event_master AS
SELECT
  lesson_event.action,
  event.created_at,
  event.id,
  lesson_event.lesson_id,
  lesson_event.payload,
  event.public,
  event.study_id,
  event.type,
  event.user_id
FROM lesson_event
JOIN event ON event.id = lesson_event.event_id;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'study_event_action') THEN
    CREATE TYPE study_event_action AS ENUM('created', 'appled', 'unappled');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS study_event (
  action   study_event_action NOT NULL,
  event_id VARCHAR(100)       PRIMARY KEY,
  study_id VARCHAR(100)       NOT NULL,
  user_id  VARCHAR(100)       NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION study_event_inserted()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  CASE NEW.action
    WHEN 'created' THEN
      INSERT INTO enrolled(enrollable_id, type, user_id)
      VALUES (NEW.study_id, 'Study', NEW.user_id);

      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM user_enrolled
      WHERE enrollable_id = NEW.user_id;
    WHEN 'appled' THEN
      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM user_enrolled
      WHERE enrollable_id = NEW.user_id;
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study_event'
    AND trigger_name = 'after_study_event_insert'
) THEN
  CREATE TRIGGER after_study_event_insert
    AFTER INSERT ON study_event
    FOR EACH ROW EXECUTE PROCEDURE study_event_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_asset_event_action (
  created_at  TIMESTAMPTZ DEFAULT statement_timestamp(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_event_action_unique_lower_name_idx
  ON user_asset_event_action (lower(name));

INSERT INTO user_asset_event_action (name, description)
VALUES
  ('created', 'The user asset was created.'),
  ('referenced', 'The user asset was referenced from a lesson.'),
  ('renamed', 'The user asset name was changed.')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_asset_event (
  action     VARCHAR(40)  NOT NULL,
  asset_id   VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ  DEFAULT statement_timestamp(),
  event_id   VARCHAR(100) PRIMARY KEY,
  payload    JSONB        NOT NULL,
  study_id   VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (action)
    REFERENCES user_asset_event_action (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (asset_id)
    REFERENCES user_asset (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (event_id)
    REFERENCES event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_event_source_id_referenced_asset_id_unique_idx
  ON user_asset_event ((payload->'source_id'), action, asset_id) WHERE action = 'referenced';

CREATE INDEX IF NOT EXISTS user_asset_event_asset_id_created_at
  ON user_asset_event (asset_id, created_at);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset_event'
    AND trigger_name = 'after_user_asset_event_delete'
) THEN
  CREATE TRIGGER after_user_asset_event_delete
    AFTER DELETE ON user_asset_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION user_asset_event_inserted()
  RETURNS TRIGGER
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
BEGIN
  CASE NEW.action
    WHEN 'created' THEN
      INSERT INTO received_event(event_id, user_id)
      SELECT
        NEW.event_id,
        user_id
      FROM study_enrolled
      WHERE enrollable_id = NEW.study_id AND user_id != NEW.user_id;
    WHEN 'referenced' THEN
      INSERT INTO user_asset_referenced_event (event_id, source_id)
      VALUES (
        NEW.event_id,
        NEW.payload->>'source_id'
      );
    WHEN 'renamed' THEN
      INSERT INTO user_asset_renamed_event (event_id, renamed_from, renamed_to)
      VALUES (
        NEW.event_id,
        NEW.payload->'rename'->>'from',
        NEW.payload->'rename'->>'to'
      );
  END CASE;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset_event'
    AND trigger_name = 'after_user_asset_event_insert'
) THEN
  CREATE TRIGGER after_user_asset_event_insert
    AFTER INSERT ON user_asset_event
    FOR EACH ROW EXECUTE PROCEDURE user_asset_event_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_asset_referenced_event (
  event_id  VARCHAR(100) PRIMARY KEY,
  source_id VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES user_asset_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset_referenced_event'
    AND trigger_name = 'after_user_asset_referenced_event_delete'
) THEN
  CREATE TRIGGER after_user_asset_referenced_event_delete
    AFTER DELETE ON user_asset_referenced_event
    FOR EACH ROW EXECUTE PROCEDURE delete_event();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_asset_renamed_event (
  event_id     VARCHAR(100) PRIMARY KEY,
  renamed_from TEXT         NOT NULL,
  renamed_to   TEXT         NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES user_asset_event (event_id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE OR REPLACE VIEW user_asset_event_master AS
SELECT
  user_asset_event.action,
  user_asset_event.asset_id,
  event.created_at,
  event.id,
  user_asset_event.payload,
  event.public,
  event.study_id,
  event.type,
  event.user_id
FROM user_asset_event
JOIN event ON event.id = user_asset_event.event_id;

CREATE OR REPLACE VIEW received_event_master AS
SELECT
  event.*,
  received_event.user_id AS received_user_id
FROM received_event
JOIN event ON event.id = received_event.event_id;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_subject_type') THEN
    CREATE TYPE notification_subject_type AS ENUM('Lesson', 'UserAsset');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS notification (
  created_at    TIMESTAMPTZ               DEFAULT statement_timestamp(),
  id            VARCHAR(100)              PRIMARY KEY,
  last_read_at  TIMESTAMPTZ,
  reason_name   VARCHAR(40)               NOT NULL,
  subject       notification_subject_type NOT NULL,
  subject_id    VARCHAR(100)              NOT NULL,
  study_id      VARCHAR(100)              NOT NULL,
  unread        BOOLEAN                   DEFAULT true,
  updated_at    TIMESTAMPTZ               DEFAULT statement_timestamp(),
  user_id       VARCHAR(100)              NOT NULL,
  FOREIGN KEY (reason_name)
    REFERENCES reason (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS notification_user_id_study_id_created_at_idx
  ON notification (user_id, study_id, created_at ASC);

CREATE OR REPLACE FUNCTION notification_inserted()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.subject = 'Lesson' THEN
    INSERT INTO lesson_notification(notification_id, lesson_id, user_id)
    VALUES (NEW.id, NEW.subject_id, NEW.user_id)
    ON CONFLICT DO NOTHING;

    IF NOT FOUND THEN
      DELETE FROM notification WHERE id = NEW.id;
      RETURN NULL;
    END IF;
  END IF;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'notification'
    AND trigger_name = 'after_notification_insert'
) THEN
  CREATE TRIGGER after_notification_insert
    AFTER INSERT ON notification
    FOR EACH ROW EXECUTE PROCEDURE notification_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION notification_will_update()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'notification'
    AND trigger_name = 'before_notification_update'
) THEN
  CREATE TRIGGER before_notification_update
    BEFORE INSERT ON notification
    FOR EACH ROW EXECUTE PROCEDURE notification_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION delete_notification()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  DELETE FROM notification
  WHERE id = OLD.notification_id;
  RETURN OLD;
END;
$$;

CREATE TABLE IF NOT EXISTS lesson_notification (
  lesson_id       VARCHAR(100) NOT NULL,
  notification_id VARCHAR(100) PRIMARY KEY,
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (notification_id)
    REFERENCES notification (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS lesson_notification_user_id_lesson_id_unique_idx
  ON lesson_notification (user_id, lesson_id);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_notification'
    AND trigger_name = 'after_lesson_notification_delete'
) THEN
  CREATE TRIGGER after_lesson_notification_delete
    AFTER DELETE ON lesson_notification
    FOR EACH ROW EXECUTE PROCEDURE delete_notification();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE VIEW notification_master AS
SELECT
  notification.*,
  reason.description AS reason
FROM notification
JOIN reason ON reason.name = notification.reason_name;

CREATE TABLE IF NOT EXISTS user_search_index (
  account_updated_at  TIMESTAMPTZ  NOT NULL,
  bio                 TEXT,
  created_at          TIMESTAMPTZ  NOT NULL,
  document            TSVECTOR     NOT NULL,
  enrollee_count      BIGINT       NOT NULL DEFAULT 0,
  id                  VARCHAR(100) PRIMARY KEY,
  login               VARCHAR(40)  NOT NULL CHECK(login !~ '(^-|--|-$)' AND login ~ '^[a-zA-Z0-9-]{1,39}$'),
  name                TEXT,
  profile_email_id    VARCHAR(100),
  profile_updated_at  TIMESTAMPTZ  NOT NULL,
  roles               TEXT [],
  study_count         BIGINT       NOT NULL DEFAULT 0,
  verified            BOOLEAN      NOT NULL DEFAULT FALSE,
  FOREIGN KEY (id)
    REFERENCES account (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (profile_email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS user_search_index_fts_idx
  ON user_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS user_search_index_created_at_idx
  ON user_search_index (created_at);

CREATE INDEX IF NOT EXISTS user_search_index_enrollee_count_idx
  ON user_search_index (enrollee_count DESC);

CREATE INDEX IF NOT EXISTS user_search_index_study_count_idx
  ON user_search_index (study_count DESC);

CREATE OR REPLACE FUNCTION account_will_insert()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    profile_updated_at TIMESTAMPTZ;
  BEGIN
    INSERT INTO user_profile(user_id)
    VALUES (NEW.id)
    RETURNING updated_at INTO profile_updated_at;
    INSERT INTO user_search_index(
      account_updated_at,
      created_at,
      document,
      id,
      login,
      profile_updated_at
    ) VALUES (
      NEW.updated_at,
      NEW.created_at,
      setweight(to_tsvector('simple', NEW.login), 'A'),
      NEW.id,
      NEW.login,
      profile_updated_at
    );

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'account'
    AND trigger_name = 'after_account_insert'
) THEN
  CREATE TRIGGER after_account_insert
    AFTER INSERT ON account
    FOR EACH ROW EXECUTE PROCEDURE account_will_insert();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION account_will_update()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'account'
    AND trigger_name = 'before_account_update'
) THEN
  CREATE TRIGGER before_account_update
    BEFORE INSERT ON account
    FOR EACH ROW EXECUTE PROCEDURE account_will_update();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION account_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.login != OLD.login THEN
      doc = (
        SELECT 
          setweight(to_tsvector('simple', NEW.login), 'A') ||
          setweight(to_tsvector('simple', coalesce(user_profile.name, '')), 'A') ||
          setweight(to_tsvector('simple', coalesce(user_profile.bio, '')), 'B') ||
          setweight(to_tsvector('simple', coalesce(email.value, '')), 'B')
        FROM user_profile
        LEFT JOIN email ON email.id = user_profile.email_id
        WHERE user_profile.user_id = NEW.id
      );
    ELSE
      doc = (SELECT document FROM user_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE user_search_index
    SET 
      account_updated_at = NEW.updated_at,
      document = doc,
      login = NEW.login
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'account'
    AND trigger_name = 'after_account_update'
) THEN
  CREATE TRIGGER after_account_update
    AFTER UPDATE ON account
    FOR EACH ROW EXECUTE PROCEDURE account_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION email_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    IF NEW.type = 'PRIMARY' AND NEW.verified_at IS NOT NULL AND OLD.verified_at IS NULL THEN
      UPDATE user_search_index
      SET verified = true 
      WHERE id = NEW.user_id;
    END IF;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'email'
    AND trigger_name = 'after_email_update'
) THEN
  CREATE TRIGGER after_email_update
    AFTER UPDATE ON email
    FOR EACH ROW EXECUTE PROCEDURE email_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION user_profile_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.name != OLD.name OR NEW.bio != OLD.bio OR NEW.email_id != OLD.email_id THEN
      doc = (
        SELECT 
          setweight(to_tsvector('simple', account.login), 'A') ||
          setweight(to_tsvector('simple', coalesce(NEW.name, '')), 'A') ||
          setweight(to_tsvector('simple', coalesce(NEW.bio, '')), 'B') ||
          setweight(to_tsvector('simple', coalesce(email.value, '')), 'B')
        FROM account
        LEFT JOIN email ON email.id = NEW.email_id
        WHERE account.id = NEW.user_id
      );
    ELSE
      doc = (SELECT document FROM user_search_index WHERE id = NEW.user_id); 
    END IF;

    UPDATE user_search_index
    SET 
      bio = NEW.bio,
      document = doc,
      name = NEW.name,
      profile_email_id = NEW.email_id,
      profile_updated_at = NEW.updated_at
    WHERE id = NEW.user_id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_profile'
    AND trigger_name = 'after_user_profile_update'
) THEN
  CREATE TRIGGER after_user_profile_update
    AFTER UPDATE ON user_profile
    FOR EACH ROW EXECUTE PROCEDURE user_profile_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_user_search_index_roles(_user_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE user_search_index
  SET roles = ARRAY(
    SELECT role.name
    FROM role
    LEFT JOIN user_role ON user_role.user_id = _user_id
    WHERE role.name = user_role.role
  )
  WHERE id = _user_id;
$$;

CREATE OR REPLACE FUNCTION user_role_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_user_search_index_roles(NEW.user_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_role'
    AND trigger_name = 'after_user_role_insert'
) THEN
  CREATE TRIGGER after_user_role_insert
    AFTER INSERT ON user_role
    FOR EACH ROW EXECUTE PROCEDURE user_role_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION user_role_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_user_search_index_roles(OLD.user_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_role'
    AND trigger_name = 'after_user_role_delete'
) THEN
  CREATE TRIGGER after_user_role_delete
    AFTER DELETE ON user_role
    FOR EACH ROW EXECUTE PROCEDURE user_role_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_user_search_index_study_count(_user_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE user_search_index
  SET study_count = (
    SELECT count(study) study_count 
    FROM study
    WHERE study.user_id = _user_id
  )
  WHERE id = _user_id;
$$;

CREATE OR REPLACE FUNCTION study_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_user_search_index_study_count(OLD.user_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study'
    AND trigger_name = 'after_study_delete'
) THEN
  CREATE TRIGGER after_study_delete
    AFTER DELETE ON study
    FOR EACH ROW EXECUTE PROCEDURE study_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_user_search_index_enrollee_count(_user_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE user_search_index
  SET enrollee_count = (
    SELECT count(enrolled) enrollee_count 
    FROM enrolled
    WHERE enrolled.enrollable_id = _user_id
  )
  WHERE id = _user_id;
$$;

CREATE OR REPLACE FUNCTION enrolled_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE NEW.type
      WHEN 'Lesson' THEN
        INSERT INTO lesson_enrolled(enrollable_id, enrolled_id, user_id)
        VALUES (NEW.enrollable_id, NEW.id, NEW.user_id);
      WHEN 'Study' THEN
        INSERT INTO study_enrolled(enrollable_id, enrolled_id, user_id)
        VALUES (NEW.enrollable_id, NEW.id, NEW.user_id);
      WHEN 'User' THEN
        INSERT INTO user_enrolled(enrollable_id, enrolled_id, user_id)
        VALUES (NEW.enrollable_id, NEW.id, NEW.user_id);
    END CASE;

    PERFORM refresh_user_search_index_enrollee_count(NEW.user_id);

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'enrolled'
    AND trigger_name = 'after_enrolled_insert'
) THEN
  CREATE TRIGGER after_enrolled_insert
    AFTER INSERT ON enrolled
    FOR EACH ROW EXECUTE PROCEDURE enrolled_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION enrolled_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_user_search_index_enrollee_count(OLD.user_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'enrolled'
    AND trigger_name = 'after_enrolled_delete'
) THEN
  CREATE TRIGGER after_enrolled_delete
    AFTER DELETE ON enrolled
    FOR EACH ROW EXECUTE PROCEDURE enrolled_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS course_search_index (
  advanced_at   TIMESTAMPTZ,
  apple_count   BIGINT       NOT NULL DEFAULT 0,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ  NOT NULL,
  description   TEXT,
  document      TSVECTOR     NOT NULL,
  id            VARCHAR(100) PRIMARY KEY,
  lesson_count  BIGINT       NOT NULL DEFAULT 0,
  name          VARCHAR(40)  NOT NULL,
  name_tokens   TEXT         NOT NULL,
  number        INT          NOT NULL CHECK(number > 0),
  published_at  TIMESTAMPTZ,
  status        course_status NOT NULL,
  study_id      VARCHAR(100) NOT NULL,
  topics        TSVECTOR     NOT NULL,   
  updated_at    TIMESTAMPTZ  NOT NULL,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (id)
    REFERENCES course (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS course_search_index_fts_idx
  ON course_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS course_search_index_topics_idx
  ON course_search_index USING gin(topics);

CREATE INDEX IF NOT EXISTS course_search_index_advanced_at_idx
  ON course_search_index (advanced_at);

CREATE INDEX IF NOT EXISTS course_search_index_created_at_idx
  ON course_search_index (created_at);

CREATE INDEX IF NOT EXISTS course_search_index_apple_count_idx
  ON course_search_index (apple_count);

CREATE INDEX IF NOT EXISTS course_search_index_lesson_count_idx
  ON course_search_index (lesson_count);

CREATE INDEX IF NOT EXISTS course_search_index_user_id_created_at_idx
  ON course_search_index (user_id, created_at);

CREATE INDEX IF NOT EXISTS course_search_index_user_id_apple_count_idx
  ON course_search_index (user_id, apple_count);

CREATE OR REPLACE FUNCTION course_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    INSERT INTO course_search_index(
      created_at,
      description,
      document,
      id,
      name,
      name_tokens,
      number,
      status,
      study_id,
      topics,
      updated_at,
      user_id
    ) VALUES (
      NEW.created_at,
      NEW.description,
      setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
      setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B'),
      NEW.id,
      NEW.name,
      NEW.name_tokens,
      NEW.number,
      NEW.status,
      NEW.study_id,
      setweight(to_tsvector('simple', ''), 'A'),
      NEW.updated_at,
      NEW.user_id
    );
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course'
    AND trigger_name = 'after_course_insert'
) THEN
  CREATE TRIGGER after_course_insert
    AFTER INSERT ON course
    FOR EACH ROW EXECUTE PROCEDURE course_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.name != OLD.name OR NEW.description != OLD.description THEN
      doc = setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B');
    ELSE
      doc = (SELECT document FROM course_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE course_search_index
    SET 
      advanced_at = NEW.advanced_at,
      completed_at = NEW.completed_at,
      description = NEW.description,
      document = doc,
      published_at = NEW.published_at,
      name = NEW.name,
      name_tokens = NEW.name_tokens,
      status = NEW.status,
      updated_at = NEW.updated_at
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course'
    AND trigger_name = 'after_course_update'
) THEN
  CREATE TRIGGER after_course_update
    AFTER UPDATE ON course
    FOR EACH ROW EXECUTE PROCEDURE course_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_course_search_index_lesson_count(_course_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE course_search_index
  SET lesson_count = (
    SELECT count(course_lesson) lesson_count 
    FROM course_lesson
    WHERE course_lesson.course_id = _course_id
  )
  WHERE id = _course_id;
$$;

CREATE OR REPLACE FUNCTION refresh_course_search_index_apple_count(_course_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE course_search_index
  SET apple_count = (
    SELECT count(appled) apple_count 
    FROM appled
    WHERE appled.appleable_id = _course_id
  )
  WHERE id = _course_id;
$$;

CREATE OR REPLACE FUNCTION refresh_course_search_index_topics(_course_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE course_search_index
  SET
    topics = (
      SELECT 
        setweight(to_tsvector('simple', coalesce(string_agg(topic.name_tokens, ' '), '')), 'A')
      FROM course_topiced
      JOIN topic ON topic.id = course_topiced.topic_id
      WHERE course_topiced.topicable_id = _course_id
    )
  WHERE id = _course_id;
$$;

CREATE OR REPLACE FUNCTION course_topiced_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_course_search_index_topics(NEW.topicable_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_topiced'
    AND trigger_name = 'after_course_topiced_insert'
) THEN
  CREATE TRIGGER after_course_topiced_insert
    AFTER INSERT ON course_topiced
    FOR EACH ROW EXECUTE PROCEDURE course_topiced_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_topiced_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_course_search_index_topics(OLD.topicable_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_topiced'
    AND trigger_name = 'after_course_topiced_delete'
) THEN
  CREATE TRIGGER after_course_topiced_delete
    AFTER DELETE ON course_topiced
    FOR EACH ROW EXECUTE PROCEDURE course_topiced_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS study_search_index (
  advanced_at   TIMESTAMPTZ,
  apple_count   BIGINT       NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ  NOT NULL,
  description   TEXT,
  document      TSVECTOR     NOT NULL,
  id            VARCHAR(100) PRIMARY KEY,
  lesson_count  BIGINT       NOT NULL DEFAULT 0,
  name          VARCHAR(40)  NOT NULL CHECK (name ~ '[\w-]{1,39}'),
  name_tokens   TEXT         NOT NULL,
  private       BOOLEAN      DEFAULT FALSE,
  topics        TSVECTOR     NOT NULL,   
  updated_at    TIMESTAMPTZ  NOT NULL,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (id)
    REFERENCES study (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS study_search_index_fts_idx
  ON study_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS study_search_index_topics_idx
  ON study_search_index USING gin(topics);

CREATE INDEX IF NOT EXISTS study_search_index_advanced_at_idx
  ON study_search_index (advanced_at DESC NULLS LAST);

CREATE INDEX IF NOT EXISTS study_search_index_created_at_idx
  ON study_search_index (created_at);

CREATE INDEX IF NOT EXISTS study_search_index_apple_count_idx
  ON study_search_index (apple_count DESC);

CREATE INDEX IF NOT EXISTS study_search_index_lesson_count_idx
  ON study_search_index (lesson_count DESC);

CREATE INDEX IF NOT EXISTS study_search_index_user_id_created_at_idx
  ON study_search_index (user_id, created_at);

CREATE INDEX IF NOT EXISTS study_search_index_user_id_apple_count_idx
  ON study_search_index (user_id, apple_count DESC);

CREATE OR REPLACE FUNCTION study_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    INSERT INTO study_search_index(
      created_at,
      description,
      document,
      id,
      name,
      name_tokens,
      private,
      topics,
      updated_at,
      user_id
    ) VALUES (
      NEW.created_at,
      NEW.description,
      setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
      setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B'),
      NEW.id,
      NEW.name,
      NEW.name_tokens,
      NEW.private,
      setweight(to_tsvector('simple', ''), 'A'),
      NEW.updated_at,
      NEW.user_id
    );
    PERFORM refresh_user_search_index_study_count(NEW.user_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study'
    AND trigger_name = 'after_study_insert'
) THEN
  CREATE TRIGGER after_study_insert
    AFTER INSERT ON study
    FOR EACH ROW EXECUTE PROCEDURE study_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION study_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.name != OLD.name OR NEW.description != OLD.description THEN
      doc = setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B');
    ELSE
      doc = (SELECT document FROM study_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE study_search_index
    SET 
      advanced_at = NEW.advanced_at,
      description = NEW.description,
      document = doc,
      name = NEW.name,
      name_tokens = NEW.name_tokens,
      private = NEW.private,
      updated_at = NEW.updated_at
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study'
    AND trigger_name = 'after_study_update'
) THEN
  CREATE TRIGGER after_study_update
    AFTER UPDATE ON study
    FOR EACH ROW EXECUTE PROCEDURE study_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_study_search_index_lesson_count(_study_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE study_search_index
  SET lesson_count = (
    SELECT count(lesson) lesson_count 
    FROM lesson
    WHERE lesson.study_id = _study_id
  )
  WHERE id = _study_id;
$$;

CREATE OR REPLACE FUNCTION lesson_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_study_search_index_lesson_count(OLD.study_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'after_lesson_delete'
) THEN
  CREATE TRIGGER after_lesson_delete
    AFTER DELETE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_study_search_index_apple_count(_study_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE study_search_index
  SET apple_count = (
    SELECT count(appled) apple_count 
    FROM appled
    WHERE appled.appleable_id = _study_id
  )
  WHERE id = _study_id;
$$;

CREATE OR REPLACE FUNCTION appled_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE NEW.type
      WHEN 'Course' THEN
        INSERT INTO course_appled(appleable_id, appled_id, user_id)
        VALUES (NEW.appleable_id, NEW.id, NEW.user_id);

        PERFORM refresh_course_search_index_apple_count(NEW.appleable_id);
      WHEN 'Study' THEN
        INSERT INTO study_appled(appleable_id, appled_id, user_id)
        VALUES (NEW.appleable_id, NEW.id, NEW.user_id);

        PERFORM refresh_study_search_index_apple_count(NEW.appleable_id);
    END CASE;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'appled'
    AND trigger_name = 'after_appled_insert'
) THEN
  CREATE TRIGGER after_appled_insert
    AFTER INSERT ON appled
    FOR EACH ROW EXECUTE PROCEDURE appled_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION appled_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_course_search_index_apple_count(OLD.appleable_id);
    PERFORM refresh_study_search_index_apple_count(OLD.appleable_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'appled'
    AND trigger_name = 'after_appled_delete'
) THEN
  CREATE TRIGGER after_appled_delete
    AFTER DELETE ON appled
    FOR EACH ROW EXECUTE PROCEDURE appled_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_study_search_index_topics(_study_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE study_search_index
  SET
    topics = (
      SELECT 
        setweight(to_tsvector('simple', coalesce(string_agg(topic.name_tokens, ' '), '')), 'A')
      FROM study_topiced
      JOIN topic ON topic.id = study_topiced.topic_id
      WHERE study_topiced.topicable_id = _study_id
    )
  WHERE id = _study_id;
$$;

CREATE OR REPLACE FUNCTION study_topiced_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_study_search_index_topics(NEW.topicable_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study_topiced'
    AND trigger_name = 'after_study_topiced_insert'
) THEN
  CREATE TRIGGER after_study_topiced_insert
    AFTER INSERT ON study_topiced
    FOR EACH ROW EXECUTE PROCEDURE study_topiced_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION study_topiced_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_study_search_index_topics(OLD.topicable_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study_topiced'
    AND trigger_name = 'after_study_topiced_delete'
) THEN
  CREATE TRIGGER after_study_topiced_delete
    AFTER DELETE ON study_topiced
    FOR EACH ROW EXECUTE PROCEDURE study_topiced_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_search_index (
  created_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  body            TEXT,
  comment_count   BIGINT       NOT NULL DEFAULT 0,
  course_id       VARCHAR(100),
  course_number   INT,
  document        TSVECTOR     NOT NULL,
  draft           TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  labels          TSVECTOR     NOT NULL,   
  last_edited_at  TIMESTAMPTZ  DEFAULT statement_timestamp(),
  number          INT          NOT NULL CHECK(number > 0),
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,    
  title           TEXT         NOT NULL,
  title_tokens    TEXT         NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT statement_timestamp(),
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (course_id)
    REFERENCES course (id)
    ON UPDATE CASCADE ON DELETE NO ACTION,
  FOREIGN KEY (id)
    REFERENCES lesson (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_search_index_fts_idx
  ON lesson_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS lesson_search_index_labels_idx
  ON lesson_search_index USING gin(labels);

CREATE INDEX IF NOT EXISTS lesson_search_index_created_at_idx
  ON lesson_search_index (created_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_updated_at_idx
  ON lesson_search_index (updated_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_comment_count_idx
  ON lesson_search_index (comment_count DESC);

CREATE INDEX IF NOT EXISTS lesson_search_index_course_id_course_number_idx
  ON lesson_search_index (course_id, course_number);

CREATE INDEX IF NOT EXISTS lesson_search_index_study_id_created_at_idx
  ON lesson_search_index (study_id, created_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_study_id_published_at_idx
  ON lesson_search_index (study_id, published_at DESC NULLS LAST);

CREATE INDEX IF NOT EXISTS lesson_search_index_study_id_updated_at_idx
  ON lesson_search_index (study_id, updated_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_study_id_comment_count_idx
  ON lesson_search_index (study_id, comment_count DESC);

CREATE INDEX IF NOT EXISTS lesson_search_index_user_id_created_at_idx
  ON lesson_search_index (user_id, created_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_user_id_published_at_idx
  ON lesson_search_index (user_id, published_at DESC NULLS LAST);

CREATE INDEX IF NOT EXISTS lesson_search_index_user_id_updated_at_idx
  ON lesson_search_index (user_id, updated_at);

CREATE INDEX IF NOT EXISTS lesson_search_index_user_id_comment_count_idx
  ON lesson_search_index (user_id, comment_count DESC);

CREATE OR REPLACE FUNCTION lesson_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM advance_study(NEW.study_id);
    PERFORM refresh_study_search_index_lesson_count(NEW.study_id);

    INSERT INTO lesson_search_index(
      created_at,
      body,
      document,
      draft,
      id,
      labels,
      number,
      study_id,
      title,
      title_tokens,
      updated_at,
      user_id
    ) VALUES (
      NEW.created_at,
      NEW.body,
      setweight(to_tsvector('simple', NEW.title_tokens), 'A') ||
      setweight(to_tsvector('english', coalesce(NEW.body, '')), 'B'),
      NEW.draft,
      NEW.id,
      setweight(to_tsvector('simple', ''), 'A'),
      NEW.number,
      NEW.study_id,
      NEW.title,
      NEW.title_tokens,
      NEW.updated_at,
      NEW.user_id
    );

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'after_lesson_insert'
) THEN
  CREATE TRIGGER after_lesson_insert
    AFTER INSERT ON lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION lesson_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.title != OLD.title OR NEW.body != OLD.body THEN
      doc = setweight(to_tsvector('simple', NEW.title_tokens), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.body, '')), 'B'); 
    ELSE
      doc = (SELECT document FROM lesson_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE lesson_search_index
    SET 
      document = doc,
      body = NEW.body,
      draft = NEW.draft,
      last_edited_at = NEW.last_edited_at,
      number = NEW.number,
      published_at = NEW.published_at,
      title = NEW.title,
      title_tokens = NEW.title_tokens,
      updated_at = NEW.updated_at
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'after_lesson_update'
) THEN
  CREATE TRIGGER after_lesson_update
    AFTER UPDATE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_lesson_search_index_course_info(_lesson_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE lesson_search_index
  SET course_id = course_lesson.course_id,
    course_number = course_lesson.number
  FROM course_lesson
  WHERE lesson_search_index.id = _lesson_id
    AND course_lesson.lesson_id = _lesson_id;
$$;

CREATE OR REPLACE FUNCTION refresh_lesson_search_index_course_number(_lesson_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE lesson_search_index
  SET course_number = course_lesson.number
  FROM course_lesson
  WHERE lesson_search_index.id = _lesson_id
    AND course_lesson.lesson_id = _lesson_id;
$$;

CREATE OR REPLACE FUNCTION course_lesson_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM advance_course(NEW.course_id);
    PERFORM refresh_course_search_index_lesson_count(NEW.course_id);
    PERFORM refresh_lesson_search_index_course_info(NEW.lesson_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_lesson'
    AND trigger_name = 'after_course_lesson_insert'
) THEN
  CREATE TRIGGER after_course_lesson_insert
    AFTER INSERT ON course_lesson
    FOR EACH ROW EXECUTE PROCEDURE course_lesson_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_lesson_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    UPDATE course_lesson
    SET number = number - 1
    WHERE course_id = OLD.course_id AND number > OLD.number;

    UPDATE lesson_search_index
    SET course_id = NULL,
      course_number = NULL
    WHERE id = OLD.lesson_id;

    PERFORM refresh_course_search_index_lesson_count(OLD.course_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_lesson'
    AND trigger_name = 'after_course_lesson_delete'
) THEN
  CREATE TRIGGER after_course_lesson_delete
    AFTER DELETE ON course_lesson
    FOR EACH ROW EXECUTE PROCEDURE course_lesson_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_lesson_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_lesson_search_index_course_number(NEW.lesson_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course_lesson'
    AND trigger_name = 'after_course_lesson_update'
) THEN
  CREATE TRIGGER after_course_lesson_update
    AFTER UPDATE ON course_lesson
    FOR EACH ROW EXECUTE PROCEDURE course_lesson_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_lesson_search_index_comment_count(_lesson_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE lesson_search_index
  SET comment_count = (
    SELECT count(lesson_comment) comment_count 
    FROM lesson_comment
    WHERE lesson_comment.lesson_id = _lesson_id
  )
  WHERE id = _lesson_id;
$$;

CREATE OR REPLACE FUNCTION lesson_comment_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    IF OLD.published_at IS NULL AND NEW.published_at IS NOT NULL THEN
      PERFORM refresh_lesson_search_index_comment_count(NEW.lesson_id);
    END IF;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'after_lesson_comment_update'
) THEN
  CREATE TRIGGER after_lesson_comment_update
    AFTER UPDATE ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION lesson_comment_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM refresh_lesson_search_index_comment_count(OLD.lesson_id);
    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'after_lesson_comment_delete'
) THEN
  CREATE TRIGGER after_lesson_comment_delete
    AFTER DELETE ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_lesson_search_index_labels(_lesson_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE lesson_search_index
  SET
    labels = (
      SELECT 
        setweight(to_tsvector('simple', coalesce(string_agg(label.name_tokens, ' '), '')), 'A')
      FROM lesson_labeled
      JOIN label ON label.id = lesson_labeled.label_id
      WHERE lesson_labeled.labelable_id = _lesson_id
    )
  WHERE id = _lesson_id;
$$;

CREATE OR REPLACE FUNCTION labeled_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE NEW.type
      WHEN 'Lesson' THEN
        INSERT INTO lesson_labeled(labelable_id, labeled_id, label_id)
        VALUES (NEW.labelable_id, NEW.id, NEW.label_id);

        PERFORM refresh_lesson_search_index_labels(NEW.labelable_id);
      WHEN 'LessonComment' THEN
        INSERT INTO lesson_comment_labeled(labelable_id, labeled_id, label_id)
        VALUES (NEW.labelable_id, NEW.id, NEW.label_id);
    END CASE;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'labeled'
    AND trigger_name = 'after_labeled_insert'
) THEN
  CREATE TRIGGER after_labeled_insert
    AFTER INSERT ON labeled
    FOR EACH ROW EXECUTE PROCEDURE labeled_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION labeled_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE OLD.type
      WHEN 'Lesson' THEN
        PERFORM refresh_lesson_search_index_labels(OLD.labelable_id);
    END CASE;

    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'labeled'
    AND trigger_name = 'after_labeled_delete'
) THEN
  CREATE TRIGGER after_labeled_delete
    AFTER DELETE ON labeled
    FOR EACH ROW EXECUTE PROCEDURE labeled_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS topic_search_index (
  created_at    TIMESTAMPTZ  NOT NULL,
  description   TEXT,
  document      TSVECTOR     NOT NULL,
  id            VARCHAR(100) PRIMARY KEY,
  name          VARCHAR(40)  NOT NULL CHECK(name ~ '^[a-zA-Z0-9-]{1,39}$'),
  name_tokens   TEXT         NOT NULL,
  topiced_count BIGINT       NOT NULL DEFAULT 0,  
  updated_at    TIMESTAMPTZ  NOT NULL,
  FOREIGN KEY (id)
    REFERENCES topic (id)
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS topic_search_index_fts_idx
  ON topic_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS topic_search_index_created_at_idx
  ON topic_search_index (created_at);

CREATE INDEX IF NOT EXISTS topic_search_index_topiced_count_idx
  ON topic_search_index (topiced_count DESC);

CREATE OR REPLACE FUNCTION topic_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    INSERT INTO topic_search_index(
      created_at,
      description,
      document,
      id,
      name,
      name_tokens,
      updated_at
    ) VALUES (
      NEW.created_at,
      NEW.description,
      setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
      setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B'),
      NEW.id,
      NEW.name,
      NEW.name_tokens,
      NEW.updated_at
    );

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topic'
    AND trigger_name = 'after_topic_insert'
) THEN
  CREATE TRIGGER after_topic_insert
    AFTER INSERT ON topic
    FOR EACH ROW EXECUTE PROCEDURE topic_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION topic_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.name != OLD.name OR NEW.description != OLD.description THEN
      doc = setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B'); 
    ELSE
      doc = (SELECT document FROM topic_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE topic_search_index
    SET 
      document = doc,
      description = NEW.description,
      name = NEW.name,
      name_tokens = NEW.name_tokens,
      updated_at = NEW.updated_at
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topic'
    AND trigger_name = 'after_topic_update'
) THEN
  CREATE TRIGGER after_topic_update
    AFTER UPDATE ON topic
    FOR EACH ROW EXECUTE PROCEDURE topic_updated();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION refresh_topic_search_index_topiced_count(_topic_id VARCHAR)
  RETURNS VOID 
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  UPDATE topic_search_index
  SET topiced_count = (
    SELECT count(topiced) topiced_count 
    FROM topiced
    WHERE topiced.topic_id = _topic_id
  )
  WHERE id = _topic_id;
$$;

CREATE OR REPLACE FUNCTION topiced_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE NEW.type
      WHEN 'Course' THEN
        INSERT INTO course_topiced(topicable_id, topiced_id, topic_id)
        VALUES (NEW.topicable_id, NEW.id, NEW.topic_id);

        PERFORM refresh_course_search_index_topics(NEW.topicable_id);
      WHEN 'Study' THEN
        INSERT INTO study_topiced(topicable_id, topiced_id, topic_id)
        VALUES (NEW.topicable_id, NEW.id, NEW.topic_id);

        PERFORM refresh_study_search_index_topics(NEW.topicable_id);
    END CASE;

    PERFORM refresh_topic_search_index_topiced_count(NEW.topic_id);
    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topiced'
    AND trigger_name = 'after_topiced_insert'
) THEN
  CREATE TRIGGER after_topiced_insert
    AFTER INSERT ON topiced
    FOR EACH ROW EXECUTE PROCEDURE topiced_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION topiced_deleted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    CASE OLD.type
      WHEN 'Course' THEN
        PERFORM refresh_course_search_index_topics(OLD.topicable_id);
      WHEN 'Study' THEN
        PERFORM refresh_study_search_index_topics(OLD.topicable_id);
    END CASE;

    PERFORM refresh_topic_search_index_topiced_count(OLD.topic_id);

    RETURN OLD;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topiced'
    AND trigger_name = 'after_topiced_delete'
) THEN
  CREATE TRIGGER after_topiced_delete
    AFTER DELETE ON topiced
    FOR EACH ROW EXECUTE PROCEDURE topiced_deleted();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_asset_search_index (
  asset_id      BIGINT       NOT NULL,
  created_at    TIMESTAMPTZ  NOT NULL,
  description   TEXT,
  document      TSVECTOR     NOT NULL,
  id            VARCHAR(100) PRIMARY KEY,
  key           TEXT         NOT NULL,
  name          VARCHAR(40)  NOT NULL CHECK(name ~ '^[\w\-.]{1,39}$'),
  name_tokens   TEXT         NOT NULL,
  original_name TEXT         NOT NULL, 
  size          BIGINT       NOT NULL,
  study_id      VARCHAR(100) NOT NULL,
  subtype       TEXT         NOT NULL,
  type          TEXT         NOT NULL,
  updated_at    TIMESTAMPTZ  NOT NULL,
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (asset_id)
    REFERENCES asset (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (id)
    REFERENCES user_asset (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS user_asset_search_index_fts_idx
  ON user_asset_search_index USING gin(document);

CREATE INDEX IF NOT EXISTS user_asset_search_index_created_at_idx
  ON user_asset_search_index (created_at);

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_search_index_study_id_name_unique_idx
  ON user_asset_search_index (study_id, lower(name));

CREATE INDEX IF NOT EXISTS user_asset_search_index_user_id_idx
  ON user_asset_search_index (user_id);

CREATE OR REPLACE FUNCTION user_asset_inserted()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  BEGIN
    INSERT INTO user_asset_search_index(
      asset_id,
      created_at,
      description,
      document,
      id,
      key,
      name,
      name_tokens,
      original_name,
      size,
      study_id,
      subtype,
      type,
      updated_at,
      user_id)
    SELECT
      asset.id,
      asset.created_at,
      NEW.description,
      setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
      setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B'),
      NEW.id,
      asset.key,
      NEW.name,
      NEW.name_tokens,
      asset.name,
      asset.size,
      NEW.study_id,
      asset.subtype,
      asset.type,
      NEW.updated_at,
      NEW.user_id
    FROM asset
    WHERE asset.id = NEW.asset_id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset'
    AND trigger_name = 'after_user_asset_insert'
) THEN
  CREATE TRIGGER after_user_asset_insert
    AFTER INSERT ON user_asset
    FOR EACH ROW EXECUTE PROCEDURE user_asset_inserted();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION user_asset_updated()
  RETURNS TRIGGER 
  SECURITY DEFINER
  LANGUAGE plpgsql
AS $$
  DECLARE
    doc TSVECTOR;
  BEGIN
    IF NEW.name != OLD.name OR NEW.description != OLD.description THEN
      doc = setweight(to_tsvector('simple', NEW.name_tokens), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B');
    ELSE
      doc = (SELECT document FROM user_asset_search_index WHERE id = NEW.id); 
    END IF;

    UPDATE user_asset_search_index
    SET 
      description = NEW.description,
      document = doc,
      name = NEW.name,
      name_tokens = NEW.name_tokens,
      updated_at = NEW.updated_at
    WHERE id = NEW.id;

    RETURN NEW;
  END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset'
    AND trigger_name = 'after_user_asset_update'
) THEN
  CREATE TRIGGER after_user_asset_update
    AFTER UPDATE ON user_asset
    FOR EACH ROW EXECUTE PROCEDURE user_asset_updated();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT
    FROM pg_catalog.pg_roles
    WHERE rolname = 'client') THEN
    CREATE ROLE client;
  END IF;
END
$$;

CREATE OR REPLACE VIEW apple_giver AS
SELECT
  user_search_index.*,
  appled.appleable_id,
  appled.created_at appled_at
FROM user_search_index
JOIN appled ON appled.user_id = user_search_index.id;

CREATE OR REPLACE VIEW appled_course AS
SELECT
  course_search_index.*,
  appled.user_id applee_id,
  appled.created_at appled_at
FROM course_search_index
JOIN appled ON appled.appleable_id = course_search_index.id;

CREATE OR REPLACE VIEW appled_study AS
SELECT
  study_search_index.*,
  appled.user_id applee_id,
  appled.created_at appled_at
FROM study_search_index
JOIN appled ON appled.appleable_id = study_search_index.id;

CREATE OR REPLACE VIEW enrollee AS
SELECT
  user_search_index.*,
  enrolled.enrollable_id,
  enrolled.created_at enrolled_at
FROM user_search_index
JOIN enrolled ON enrolled.user_id = user_search_index.id;

CREATE OR REPLACE VIEW enrolled_lesson AS
SELECT
  lesson_search_index.*,
  enrolled.created_at enrolled_at,
  enrolled.user_id enrollee_id
FROM lesson_search_index
JOIN enrolled ON enrolled.enrollable_id = lesson_search_index.id AND enrolled.status = 'ENROLLED'; 

CREATE OR REPLACE VIEW enrolled_study AS
SELECT
  study_search_index.*,
  enrolled.created_at enrolled_at,
  enrolled.user_id enrollee_id
FROM study_search_index
JOIN enrolled ON enrolled.enrollable_id = study_search_index.id AND enrolled.status = 'ENROLLED'; 

CREATE OR REPLACE VIEW enrolled_user AS
SELECT
  user_search_index.*,
  enrolled.created_at enrolled_at,
  enrolled.user_id enrollee_id
FROM user_search_index
JOIN enrolled ON enrolled.enrollable_id = user_search_index.id AND enrolled.status = 'ENROLLED';

CREATE OR REPLACE VIEW labeled_lesson AS
SELECT
  lesson_search_index.*,
  lesson_labeled.label_id,
  lesson_labeled.created_at labeled_at
FROM lesson_search_index
JOIN lesson_labeled ON lesson_labeled.labelable_id = lesson_search_index.id;

CREATE OR REPLACE VIEW topicable_topic AS
SELECT
  topic_search_index.*,
  topiced.topicable_id,
  topiced.created_at topiced_at
FROM topic_search_index
JOIN topiced ON topiced.topic_id = topic_search_index.id;

CREATE OR REPLACE VIEW topiced_course AS
SELECT
  course_search_index.*,
  course_topiced.topic_id,
  course_topiced.created_at topiced_at
FROM course_search_index
JOIN course_topiced ON course_topiced.topicable_id = course_search_index.id;

CREATE OR REPLACE VIEW topiced_study AS
SELECT
  study_search_index.*,
  study_topiced.topic_id,
  study_topiced.created_at topiced_at
FROM study_search_index
JOIN study_topiced ON study_topiced.topicable_id = study_search_index.id;

GRANT CONNECT ON DATABASE markusninja TO client;
GRANT USAGE ON SCHEMA public TO client;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON account TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_profile TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON email TO client;
GRANT SELECT ON role TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_role TO client;
GRANT SELECT ON user_master TO client;
GRANT SELECT ON user_credentials TO client;
GRANT SELECT ON permission TO client;
GRANT SELECT ON role_permission TO client;
GRANT SELECT ON role_permission_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON email_verification_token TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON password_reset_token TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON study TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson TO client;
GRANT SELECT, UPDATE ON lesson_draft_backup TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON course TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON course_lesson TO client;
GRANT SELECT ON lesson_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_comment TO client;
GRANT SELECT, UPDATE ON lesson_comment_draft_backup TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON label TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON labeled TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_labeled TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_comment_labeled TO client;
GRANT SELECT ON labelable_label TO client;
GRANT SELECT ON labeled_lesson TO client;
GRANT SELECT ON labeled_lesson_comment TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON topic TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON topiced TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON course_topiced TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON study_topiced TO client;
GRANT SELECT ON topicable_topic TO client;
GRANT SELECT ON topiced_course TO client;
GRANT SELECT ON topiced_study TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON asset TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_asset TO client;
GRANT SELECT ON user_asset_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON appled TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON course_appled TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON study_appled TO client;
GRANT SELECT ON apple_giver TO client;
GRANT SELECT ON appled_course TO client;
GRANT SELECT ON appled_study TO client;
GRANT SELECT ON reason TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON enrolled TO client;
GRANT SELECT ON lesson_enrolled TO client;
GRANT SELECT ON study_enrolled TO client;
GRANT SELECT ON user_enrolled TO client;
GRANT SELECT ON enrollee TO client;
GRANT SELECT ON enrolled_lesson TO client;
GRANT SELECT ON enrolled_study TO client;
GRANT SELECT ON enrolled_user TO client;
GRANT SELECT ON event_type TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON received_event TO client;
GRANT SELECT ON received_event_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON course_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_added_to_course_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_commented_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_labeled_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_referenced_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_removed_from_course_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_renamed_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_unlabeled_event TO client;
GRANT SELECT ON lesson_event_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON study_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_asset_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_asset_referenced_event TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_asset_renamed_event TO client;
GRANT SELECT ON user_asset_event_master TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON notification TO client;
GRANT SELECT, INSERT, UPDATE, DELETE ON lesson_notification TO client;
GRANT SELECT ON notification_master TO client;
GRANT SELECT ON user_search_index TO client;
GRANT SELECT ON course_search_index TO client;
GRANT SELECT ON study_search_index TO client;
GRANT SELECT ON lesson_search_index TO client;
GRANT SELECT ON label_search_index TO client;
GRANT SELECT ON topic_search_index TO client;
GRANT SELECT ON user_asset_search_index TO client;
