CREATE TABLE roles
(
  id   VARCHAR(25) PRIMARY KEY,
  name VARCHAR(45) NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles VALUES('0', 'ADMIN');
