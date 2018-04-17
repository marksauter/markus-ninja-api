CREATE TABLE account_role
(
  user_id     VARCHAR(45) NOT NULL,
  role_id     SMALLINT    NOT NULL,
  granted_at  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id),
  CONSTRAINT account_role_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT account_role_role_id_fkey FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);
