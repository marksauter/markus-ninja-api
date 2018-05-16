DROP TABLE IF EXISTS user_asset;
CREATE TABLE user_asset(
  id            VARCHAR(40) PRIMARY KEY,
  user_id       VARCHAR(40) NOT NULL,
  name          TEXT        NOT NULL,
  size          BIGINT      NOT NULL,
  content_type  TEXT        NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
