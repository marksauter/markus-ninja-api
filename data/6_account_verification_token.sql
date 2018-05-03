CREATE TABLE IF NOT EXISTS account_verification_token(
  token         VARCHAR(40)   PRIMARY KEY,
  user_id       VARCHAR(40)   NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  verified_at   TIMESTAMPTZ,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
