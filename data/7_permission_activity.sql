CREATE TABLE permission_activity (
  activity_id   VARCHAR(45),
  permission_id VARCHAR(45),
  granted_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (activity_id, permission_id),
  FOREIGN KEY (activity_id)
    REFERENCES activity (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (permission_id)
    REFERENCES permission (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);
