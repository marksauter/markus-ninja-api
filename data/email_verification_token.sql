DROP TABLE IF EXISTS email_verification_token;
CREATE TABLE email_verification_token(
  email_id      VARCHAR(40),
  user_id       VARCHAR(40),
  token         VARCHAR(40),
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  verified_at   TIMESTAMPTZ,
  PRIMARY KEY (email_id, user_id, token),
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
