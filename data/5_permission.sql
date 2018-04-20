CREATE TYPE access_level AS ENUM('Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete');
CREATE TYPE node_type AS ENUM('Label', 'Lesson', 'LessonComment', 'Study', 'User');

CREATE TABLE IF NOT EXISTS permission(
  id            VARCHAR(45) PRIMARY KEY,
  access_level  access_level NOT NULL,
  type          node_type   NOT NULL,
  field         TEXT,
  created_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX permission_access_level_type_field_key
ON permission (access_level, type, field)
WHERE field IS NOT NULL;

CREATE UNIQUE INDEX permission_access_level_type_key
ON permission (access_level, type)
WHERE field IS NULL;

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER permission_updated_at_modtime
BEFORE UPDATE ON permission
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

