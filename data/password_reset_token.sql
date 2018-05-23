CREATE TYPE password_reset_state AS ENUM('FAILURE', 'PENDING', 'SUCCESS')

CREATE TABLE IF NOT EXISTS password_reset_token(
  user_id       VARCHAR(40),
  token         VARCHAR(40),
  email         VARCHAR(40)   NOT NULL,
  request_ip    INET          NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  status        password_reset_state DEFAULT 'PENDING',
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  PRIMARY KEY (user_id, token)
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);
