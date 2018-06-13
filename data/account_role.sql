CREATE TABLE account_role(
  user_id     VARCHAR(40),
  role_id     VARCHAR(40),
  granted_at  TIMESTAMPTZ   DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
