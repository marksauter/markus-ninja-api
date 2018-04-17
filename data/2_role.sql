CREATE TABLE role(
  id          SMALLSERIAL PRIMARY KEY,
  name        VARCHAR(45) NOT NULL UNIQUE,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO role (name)
  VALUES
    ('admin'),
    ('member'),
    ('user');
