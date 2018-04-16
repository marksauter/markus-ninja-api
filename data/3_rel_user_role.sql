CREATE TABLE rel_user_role (
  user_id VARCHAR(45) REFERENCES "user" (id) ON UPDATE CASCADE,
  role_id INTEGER REFERENCES role (id) ON UPDATE CASCADE,
  CONSTRAINT user_role_pkey PRIMARY KEY (user_id, role_id)
);
