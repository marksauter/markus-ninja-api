CREATE TYPE access_level AS ENUM('Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete');
CREATE TYPE node_type AS ENUM('Label', 'Lesson', 'LessonComment', 'Study', 'User');

CREATE TABLE IF NOT EXISTS activity (
  id            VARCHAR(45) PRIMARY KEY,
  access_level  access_level NOT NULL,
  type          node_type NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (access_level, type)
)
