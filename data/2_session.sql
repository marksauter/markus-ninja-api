CREATE TABLE session(
  id            VARCHAR(40) PRIMARY KEY,
  user_id       VARCHAR(40) NOT NULL,
  session_ip    INET        NOT NULL,
  user_agent    TEXT        NOT NULL,
  started_at    TIMESTAMPTZ DEFAULT NOW(),
  
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
);
