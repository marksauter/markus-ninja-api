CREATE TABLE IF NOT EXISTS password_reset_token(
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
