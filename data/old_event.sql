CREATE SCHEMA IF NOT EXISTS event;

CREATE TABLE IF NOT EXISTS event.action (
  created_at  TIMESTAMPTZ DEFAULT now(),
  description TEXT        NOT NULL,
  name        VARCHAR(40) PRIMARY KEY
);

CREATE UNIQUE INDEX IF NOT EXISTS action_unique_lower_name_idx
  ON event.action (lower(name));

INSERT INTO event.action (name, description)
VALUES
  ('appled', 'Source appled target appleable.'),
  ('created', 'Source created target creatable.'),
  ('commented', 'Source commented target commentable.'),
  ('deleted', 'Source deleted target deletable.'),
  ('dismissed', 'Source dimissed from target enrollable.'),
  ('enrolled', 'Source enrolled in target enrollable.'),
  ('mentioned', 'Source @mentioned target mentionable.'),
  ('referenced', 'Source #referenced target referencable.')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS event.event (
  action          VARCHAR(40)  NOT NULL,
  created_at      TIMESTAMPTZ  DEFAULT now(),
  id              VARCHAR(100) PRIMARY KEY,
  source_id       VARCHAR(100) NOT NULL,
  target_id       VARCHAR(100) NOT NULL,
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (action)
    REFERENCES event.action (name)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS event_source_id_action_target_id_created_at_idx
  ON event.event (source_id, action, target_id, created_at);

CREATE INDEX IF NOT EXISTS event_user_id_action_created_at_idx
  ON event.event (user_id, action, created_at);

CREATE INDEX IF NOT EXISTS event_source_id_created_at_idx
  ON event.event (source_id, created_at);

CREATE INDEX IF NOT EXISTS event_target_id_created_at_idx
  ON event.event (target_id, created_at);

CREATE OR REPLACE FUNCTION insert_event()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  INSERT INTO event.event(action, id, source_Id, target_id, user_id)
  VALUES (TG_ARGV[0], NEW.event_id, NEW.source_id, NEW.target_id, NEW.user_id);
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION delete_events()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  DELETE FROM event.event
  WHERE source_id = OLD.id OR target_id = OLD.id;
  RETURN OLD;
END;
$$;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'before_lesson_delete'
) THEN
  CREATE TRIGGER before_lesson_delete
    AFTER DELETE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE delete_events();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'before_lesson_comment_delete'
) THEN
  CREATE TRIGGER before_lesson_comment_delete
    AFTER DELETE ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE delete_events();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'course'
    AND trigger_name = 'before_course_delete'
) THEN
  CREATE TRIGGER before_course_delete
    AFTER DELETE ON course
    FOR EACH ROW EXECUTE PROCEDURE delete_events();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'study'
    AND trigger_name = 'before_study_delete'
) THEN
  CREATE TRIGGER before_study_delete
    AFTER DELETE ON study
    FOR EACH ROW EXECUTE PROCEDURE delete_events();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'account'
    AND trigger_name = 'before_account_delete'
) THEN
  CREATE TRIGGER before_account_delete
    AFTER DELETE ON account
    FOR EACH ROW EXECUTE PROCEDURE delete_events();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION enroll_user_in_lesson(_user_id VARCHAR, _lesson_id VARCHAR, _reason VARCHAR)
  RETURNS VOID
  SECURITY DEFINER
  LANGUAGE sql
AS $$
  INSERT INTO lesson_enrolled(enrollable_id, reason_name, user_id)
  VALUES (_lesson_id, _reason, _user_id);
$$;

CREATE TABLE IF NOT EXISTS event.lesson_created_comment (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES lesson (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_created_comment'
    AND trigger_name = 'before_lesson_created_comment_insert'
) THEN
  CREATE TRIGGER before_lesson_created_comment_insert
    BEFORE INSERT ON event.lesson_created_comment
    FOR EACH ROW EXECUTE PROCEDURE insert_event('created');
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_created_comment'
    AND trigger_name = 'after_lesson_created_comment_insert'
) THEN
  CREATE TRIGGER after_lesson_created_comment_insert
    AFTER INSERT ON event.lesson_created_comment
    FOR EACH ROW EXECUTE PROCEDURE lesson_comment_enroll_user('comment');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.lesson_mentioned_user (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES account (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_mentioned_user'
    AND trigger_name = 'before_lesson_mentioned_user_insert'
) THEN
  CREATE TRIGGER before_lesson_mentioned_user_insert
    BEFORE INSERT ON event.lesson_mentioned_user
    FOR EACH ROW EXECUTE PROCEDURE insert_event('mentioned');
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_mentioned_user'
    AND trigger_name = 'after_lesson_mentioned_user_insert'
) THEN
  CREATE TRIGGER after_lesson_mentioned_user_insert
    AFTER INSERT ON event.lesson_mentioned_user
    FOR EACH ROW EXECUTE PROCEDURE lesson_enroll_target('mention');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.lesson_referenced_lesson (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES lesson (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_referenced_lesson'
    AND trigger_name = 'before_lesson_referenced_lesson_insert'
) THEN
  CREATE TRIGGER before_lesson_referenced_lesson_insert
    BEFORE INSERT ON event.lesson_referenced_lesson
    FOR EACH ROW EXECUTE PROCEDURE insert_event('referenced');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.lesson_referenced_user_asset (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES user_asset (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'lesson_referenced_user_asset'
    AND trigger_name = 'before_lesson_referenced_user_asset_insert'
) THEN
  CREATE TRIGGER before_lesson_referenced_user_asset_insert
    BEFORE INSERT ON event.lesson_referenced_user_asset
    FOR EACH ROW EXECUTE PROCEDURE insert_event('referenced');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.study_created_lesson (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES lesson (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'study_created_lesson'
    AND trigger_name = 'before_study_created_lesson_insert'
) THEN
  CREATE TRIGGER before_study_created_lesson_insert
    BEFORE INSERT ON event.study_created_lesson
    FOR EACH ROW EXECUTE PROCEDURE insert_event('created');
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'study_created_lesson'
    AND trigger_name = 'after_study_created_lesson_insert'
) THEN
  CREATE TRIGGER after_study_created_lesson_insert
    AFTER INSERT ON event.study_created_lesson
    FOR EACH ROW EXECUTE PROCEDURE lesson_enroll_user('author');
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION course_enroll_user()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  INSERT INTO course_enrolled(enrollable_id, reason_name, user_id)
  VALUES (NEW.target_id, TG_ARGV[0], NEW.user_id);
  RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS event.user_appled_course (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES course (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_appled_course'
    AND trigger_name = 'before_user_appled_course_insert'
) THEN
  CREATE TRIGGER before_user_appled_course_insert
    BEFORE INSERT ON event.user_appled_course
    FOR EACH ROW EXECUTE PROCEDURE insert_event('appled');
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION study_enroll_user()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  INSERT INTO study_enrolled(enrollable_id, reason_name, user_id)
  VALUES (NEW.target_id, TG_ARGV[0], NEW.user_id);
  RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS event.user_appled_study (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_appled_study'
    AND trigger_name = 'before_user_appled_study_insert'
) THEN
  CREATE TRIGGER before_user_appled_study_insert
    BEFORE INSERT ON event.user_appled_study
    FOR EACH ROW EXECUTE PROCEDURE insert_event('appled');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_created_course (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES course (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_created_course'
    AND trigger_name = 'before_user_created_course_insert'
) THEN
  CREATE TRIGGER before_user_created_course_insert
    BEFORE INSERT ON event.user_created_course
    FOR EACH ROW EXECUTE PROCEDURE insert_event('created');
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_created_course'
    AND trigger_name = 'after_user_created_course_insert'
) THEN
  CREATE TRIGGER after_user_created_course_insert
    AFTER INSERT ON event.user_created_course
    FOR EACH ROW EXECUTE PROCEDURE course_enroll_user('author');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_created_study (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_created_study'
    AND trigger_name = 'before_user_created_study_insert'
) THEN
  CREATE TRIGGER before_user_created_study_insert
    BEFORE INSERT ON event.user_created_study
    FOR EACH ROW EXECUTE PROCEDURE insert_event('created');
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_created_study'
    AND trigger_name = 'after_user_created_study_insert'
) THEN
  CREATE TRIGGER after_user_created_study_insert
    AFTER INSERT ON event.user_created_study
    FOR EACH ROW EXECUTE PROCEDURE study_enroll_user('author');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_dismissed_lesson (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES lesson (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_dismissed_lesson'
    AND trigger_name = 'after_user_dismissed_lesson_insert'
) THEN
  CREATE TRIGGER after_user_dismissed_lesson_insert
    BEFORE INSERT ON event.user_dismissed_lesson
    FOR EACH ROW EXECUTE PROCEDURE insert_event('dismissed');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_enrolled_lesson (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES lesson (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_enrolled_lesson'
    AND trigger_name = 'before_user_enrolled_lesson_insert'
) THEN
  CREATE TRIGGER before_user_enrolled_lesson_insert
    BEFORE INSERT ON event.user_enrolled_lesson
    FOR EACH ROW EXECUTE PROCEDURE insert_event('enrolled');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_dismissed_study (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_dismissed_study'
    AND trigger_name = 'before_user_dismissed_study_insert'
) THEN
  CREATE TRIGGER before_user_dismissed_study_insert
    BEFORE INSERT ON event.user_dismissed_study
    FOR EACH ROW EXECUTE PROCEDURE insert_event('dismissed');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_enrolled_study (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_enrolled_study'
    AND trigger_name = 'before_user_enrolled_study_insert'
) THEN
  CREATE TRIGGER before_user_enrolled_study_insert
    BEFORE INSERT ON event.user_enrolled_study
    FOR EACH ROW EXECUTE PROCEDURE insert_event('enrolled');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_dismissed_user (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_dismissed_user'
    AND trigger_name = 'before_user_dismissed_user_insert'
) THEN
  CREATE TRIGGER before_user_dismissed_user_insert
    BEFORE INSERT ON event.user_dismissed_user
    FOR EACH ROW EXECUTE PROCEDURE insert_event('dismissed');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_enrolled_user (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_enrolled_user'
    AND trigger_name = 'before_user_enrolled_user_insert'
) THEN
  CREATE TRIGGER before_user_enrolled_user_insert
    BEFORE INSERT ON event.user_enrolled_user
    FOR EACH ROW EXECUTE PROCEDURE insert_event('enrolled');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.user_asset_comment_commented_user_asset (
  event_id   VARCHAR(100) PRIMARY KEY,
  source_id  VARCHAR(100) NOT NULL,
  target_id  VARCHAR(100) NOT NULL,
  user_id    VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (source_id)
    REFERENCES user_asset_comment (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (target_id)
    REFERENCES user_asset (id)
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
    WHERE event_object_schema = 'event' 
    AND event_object_table = 'user_asset_comment_commented_user_asset'
    AND trigger_name = 'before_user_asset_comment_commented_user_asset_insert'
) THEN
  CREATE TRIGGER before_user_asset_comment_commented_user_asset_insert
    BEFORE INSERT ON event.user_asset_comment_commented_user_asset
    FOR EACH ROW EXECUTE PROCEDURE insert_event('commented');
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS event.received_event (
  event_id      VARCHAR(100) NOT NULL,
  user_id       VARCHAR(100) NOT NULL,
  PRIMARY KEY (user_id, event_id),
  FOREIGN KEY (event_id)
    REFERENCES event.event (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_subject_type') THEN
    CREATE TYPE notification_subject_type AS ENUM('Lesson', 'UserAsset');
  END IF;
END
$$ language 'plpgsql';
