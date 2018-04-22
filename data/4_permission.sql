CREATE TYPE access_level AS ENUM('Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete');
CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');
CREATE TYPE node_type AS ENUM('Label', 'Lesson', 'LessonComment', 'Study', 'User');

CREATE TABLE IF NOT EXISTS permission(
  id            VARCHAR(45)   PRIMARY KEY,
  access_level  access_level  NOT NULL,
  type          node_type     NOT NULL,
  created_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
  audience      audience,
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

