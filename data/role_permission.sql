CREATE TABLE role_permission(
  role_id       VARCHAR(40),
  permission_id VARCHAR(40),
  granted_at    TIMESTAMPTZ   DEFAULT NOW(),
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (permission_id)
    REFERENCES permission (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
