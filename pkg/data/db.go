package data

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"strings"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

var ErrNotFound = errors.New("not found")

func Initialize(db Queryer) error {
	_, err := db.Exec(initDBSQL)
	return err
}

type Queryer interface {
	CopyFrom(pgx.Identifier, []string, pgx.CopyFromSource) (int, error)
	Exec(sql string, arguments ...interface{}) (pgx.CommandTag, error)
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
	QueryRow(sql string, args ...interface{}) *pgx.Row
}

type transactor interface {
	Begin() (*pgx.Tx, error)
}

type committer interface {
	Commit() error
	Rollback() error
}

type preparer interface {
	Prepare(name, sql string) (*pgx.PreparedStatement, error)
	Deallocate(name string) error
}

func beginTransaction(db Queryer) (Queryer, error, bool) {
	if transactor, ok := db.(transactor); ok {
		tx, err := transactor.Begin()
		return tx, err, true
	}
	return db, nil, false
}

func commitTransaction(db Queryer) error {
	if committer, ok := db.(committer); ok {
		return committer.Commit()
	}
	return nil
}

func rollbackTransaction(db Queryer) error {
	if committer, ok := db.(committer); ok {
		return committer.Rollback()
	}
	return nil
}

func prepareQuery(db Queryer, name, sql string, args ...interface{}) (*pgx.Rows, error) {
	if preparer, ok := db.(preparer); ok {
		if _, err := preparer.Prepare(name, sql); err != nil {
			return nil, err
		}
		sql = name
	}

	return db.Query(sql, args...)
}

func prepareQueryRow(db Queryer, name, sql string, args ...interface{}) *pgx.Row {
	if preparer, ok := db.(preparer); ok {
		// QueryRow doesn't return an error, the error is encoded in the pgx.Row.
		// Since that is private, Ignore the error from Prepare and run the query
		// without the prepared statement. It should fail with the same error.
		if _, err := preparer.Prepare(name, sql); err == nil {
			sql = name
		}
	}

	return db.QueryRow(sql, args...)
}

func prepareExec(db Queryer, name, sql string, args ...interface{}) (pgx.CommandTag, error) {
	if preparer, ok := db.(preparer); ok {
		if _, err := preparer.Prepare(name, sql); err != nil {
			return pgx.CommandTag(""), err
		}
		sql = name
	}

	return db.Exec(sql, args...)
}

func preparedName(baseName, sql string) string {
	h := fnv.New32a()
	if _, err := io.WriteString(h, sql); err != nil {
		// hash.Hash.Write never returns an error so this can't happen
		panic("failed writing to hash")
	}

	return fmt.Sprintf("%s%d", baseName, h.Sum32())
}

type OrderDirection bool

const (
	ASC  OrderDirection = false
	DESC OrderDirection = true
)

func ParseOrderDirection(s string) (OrderDirection, error) {
	switch strings.ToLower(s) {
	case "asc":
		return ASC, nil
	case "desc":
		return DESC, nil
	default:
		var o OrderDirection
		return o, fmt.Errorf("invalid OrderDirection: %q", s)
	}
}

func (od OrderDirection) String() string {
	switch od {
	case ASC:
		return "ASC"
	case DESC:
		return "DESC"
	default:
		return "unknown"
	}
}

type Order interface {
	Direction() OrderDirection
	Field() string
}

var ErrEmptyPageOptions = errors.New("`po` (*PageOptions) must not be nil")

type PageOptions struct {
	After  *Cursor
	Before *Cursor
	First  int32
	Last   int32
	Order  Order
}

func NewPageOptions(after, before *string, first, last *int32, o Order) (*PageOptions, error) {
	pageOptions := &PageOptions{
		Order: o,
	}
	if first == nil && last == nil {
		return nil, fmt.Errorf("You must provide a `first` or `last` value to properly paginate.")
	} else if first != nil {
		if last != nil {
			return nil, fmt.Errorf("Passing both `first` and `last` values to paginate the connection is not supported.")
		}
		pageOptions.First = *first
	} else {
		pageOptions.Last = *last
	}
	if after != nil {
		a, err := NewCursor(after)
		if err != nil {
			return nil, err
		}
		pageOptions.After = a
	}
	if before != nil {
		b, err := NewCursor(before)
		if err != nil {
			return nil, err
		}
		pageOptions.Before = b
	}
	return pageOptions, nil
}

// If the query is asking for the last elements in a list, then we need two
// queries to get the items more efficiently and in the right order.
// First, we query the reverse direction of that requested, so that only
// the items needed are returned.
func (p *PageOptions) QueryDirection() string {
	direction := p.Order.Direction()
	if p.Last != 0 {
		direction = !p.Order.Direction()
	}
	return direction.String()
}

func (p *PageOptions) Limit() int32 {
	// Assuming one of these is 0, so the sum will be the non-zero field + 1
	limit := p.First + p.Last + 1
	if (p.After != nil && p.First > 0) ||
		(p.Before != nil && p.Last > 0) {
		limit = limit + int32(1)
	}
	return limit
}

func (p *PageOptions) joins(from string, args *pgx.QueryArgs) []string {
	var joins []string
	if p.After != nil {
		joins = append(joins, fmt.Sprintf(
			"JOIN %[1]s %[1]s2 ON %[1]s2.id = "+args.Append(p.After.Value()),
			from,
		))
	}
	if p.Before != nil {
		joins = append(joins, fmt.Sprintf(
			"JOIN %[1]s %[1]s3 ON %[1]s3.id = "+args.Append(p.Before.Value()),
			from,
		))
	}
	return joins
}

func (p *PageOptions) whereAnds(from string) []string {
	var whereAnds []string
	field := p.Order.Field()
	if p.After != nil {
		relation := ""
		switch p.Order.Direction() {
		case ASC:
			relation = ">="
		case DESC:
			relation = "<="
		}
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %[1]s.%[2]s %s %[1]s2.%[2]s",
			from,
			field,
			relation,
		))
	}
	if p.Before != nil {
		relation := ""
		switch p.Order.Direction() {
		case ASC:
			relation = "<="
		case DESC:
			relation = ">="
		}
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %[1]s.%[2]s %s %[1]s3.%[2]s",
			from,
			field,
			relation,
		))
	}
	return whereAnds
}

func SQL(selects []string, from, where string, args *pgx.QueryArgs, po *PageOptions) string {
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, args)
		whereAnds = po.whereAnds(from)
		limit = "LIMIT " + args.Append(po.Limit())
		orderBy = "ORDER BY " +
			from + "." + po.Order.Field() + " " + po.QueryDirection()
	}
	for i, s := range selects {
		selects[i] = from + "." + s
	}

	sql := `
		SELECT 
		` + strings.Join(selects, ",") + `
		FROM ` + from + `
		` + strings.Join(joins, " ") + `
		WHERE ` + where + `
		` + strings.Join(whereAnds, " ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql)
}

func SearchSQL(
	selects []string,
	from string,
	within *mytype.OID,
	query string,
	po *PageOptions,
) (string, pgx.QueryArgs) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, &args)
		whereAnds = po.whereAnds(from)
		limit = "LIMIT " + args.Append(po.Limit())

		field := po.Order.Field()
		orderBy := ""
		if field != "best_match" {
			orderBy = from + "." + field
		} else {
			orderBy = "ts_rank(document, query)"
		}

		orderBy = "ORDER BY " + orderBy + " " + po.QueryDirection()
	}
	if within != nil {
		andIn := fmt.Sprintf(
			"AND %s.%s = %s",
			from,
			within.DBVarName(),
			args.Append(within),
		)
		whereAnds = append(whereAnds, andIn)
	}

	tsquery := ToTsQuery(query)
	sql := `
		SELECT 
		` + strings.Join(selects, ",") + `
		FROM ` + from + `, to_tsquery('simple',` + args.Append(tsquery) + `) query
		` + strings.Join(joins, " ") + `
		WHERE document @@ query
		` + strings.Join(whereAnds, " ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql), args
}

// Then, we can reorder the items to the originally requested direction.
func ReorderQuery(po *PageOptions, query string) string {
	if po != nil && po.Last != 0 {
		return fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			query,
			po.Order.Field(),
			po.Order.Direction(),
		)
	}
	return query
}

const initDBSQL = `
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS account(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  login         VARCHAR(40)  NOT NULL,
  name          TEXT         CHECK(name ~ '^[\w|-][\w|-|\s]+[\w|-]$'),
  password      BYTEA        NOT NULL,
  profile       TEXT,
  updated_at    TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS account_unique__login__idx
  ON account (LOWER(login));

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'account'
    AND trigger_name = 'account_updated_at_modtime'
) THEN
  CREATE TRIGGER account_updated_at_modtime
    BEFORE UPDATE ON account
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'email_type') THEN
    CREATE TYPE email_type AS ENUM('BACKUP', 'EXTRA', 'PRIMARY');
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS email(
  id          VARCHAR(100) PRIMARY KEY,
  user_id     VARCHAR(100) NOT NULL,
  value       VARCHAR(40) NOT NULL,
  type        email_type  DEFAULT 'EXTRA',
  public      BOOLEAN     DEFAULT FALSE,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW(),
  verified_at TIMESTAMPTZ,
  CONSTRAINT check_verified_before_public
    CHECK ((public AND verified_at IS NOT NULL) OR NOT public),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS email_unique__value__idx
  ON email (LOWER(value));
CREATE INDEX IF NOT EXISTS email_user_id_idx ON email (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS email_unique__user_id_type__idx
  ON email (user_id, type)
  WHERE type = ANY('{"PRIMARY", "BACKUP"}');
CREATE UNIQUE INDEX IF NOT EXISTS email_unique__user_id_public__idx
  ON email (user_id, public)
  WHERE public = TRUE;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'email'
    AND trigger_name = 'email_updated_at_modtime'
) THEN
  CREATE TRIGGER email_updated_at_modtime
    BEFORE UPDATE ON email
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS role(
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40) NOT NULL UNIQUE,
  updated_at  TIMESTAMPTZ DEFAULT NOW()
);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'role'
    AND trigger_name = 'role_updated_at_modtime'
) THEN
  CREATE TRIGGER role_updated_at_modtime
    BEFORE UPDATE ON role
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS user_role(
  granted_at  TIMESTAMPTZ   DEFAULT NOW(),
  role_id     VARCHAR(100),
  user_id     VARCHAR(100),
  PRIMARY KEY (user_id, role_id),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'access_level') THEN
    CREATE TYPE access_level AS ENUM(
      'Read', 'Create', 'Connect', 'Disconnect', 'Update', 'Delete'
    );
  END IF;
END
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audience') THEN
    CREATE TYPE audience AS ENUM('AUTHENTICATED', 'EVERYONE');
  END IF;
END
$$ language 'plpgsql';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'node_type') THEN
    CREATE TYPE node_type AS ENUM(
      'Email',
      'EVT',
      'Label',
      'Lesson',
      'LessonComment',
      'PRT',
      'Study',
      'Topic',
      'User',
      'UserAsset'
    );
  END IF;
END
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS permission(
  access_level access_level NOT NULL,
  audience     audience     NOT NULL,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  field        TEXT,
  id           VARCHAR(100) PRIMARY KEY,
  type         node_type    NOT NULL,
  updated_at   TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS permission__access_level_type_field__key
  ON permission (access_level, type, field)
  WHERE field IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS permission__access_level_type__key
  ON permission (access_level, type)
  WHERE field IS NULL;

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'permission'
    AND trigger_name = 'permission_updated_at_modtime'
) THEN
  CREATE TRIGGER permission_updated_at_modtime
    BEFORE UPDATE ON permission
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS role_permission(
  granted_at    TIMESTAMPTZ   DEFAULT NOW(),
  role_id       VARCHAR(100),
  permission_id VARCHAR(100),
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id)
    REFERENCES role (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (permission_id)
    REFERENCES permission (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS email_verification_token(
  email_id      VARCHAR(100),
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  token         VARCHAR(40),
  user_id       VARCHAR(100)  NOT NULL,
  verified_at   TIMESTAMPTZ,
  PRIMARY KEY (email_id, token),
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS email_verification_token_user_id_idx
  ON email_verification_token (user_id); 

CREATE TABLE IF NOT EXISTS password_reset_token(
  email_id      VARCHAR(100)  NOT NULL,
  issued_at     TIMESTAMPTZ   DEFAULT NOW(),
  end_ip        INET,
  ended_at      TIMESTAMPTZ,
  expires_at    TIMESTAMPTZ   DEFAULT (NOW() + interval '20 minutes'),
  request_ip    INET          NOT NULL,
  token         VARCHAR(40),
  user_id       VARCHAR(100),
  PRIMARY KEY (user_id, token),
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION,
  FOREIGN KEY (email_id)
    REFERENCES email (id)
    ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS study(
  advanced_at   TIMESTAMPTZ,
  created_at    TIMESTAMPTZ   DEFAULT NOW(),
  description   TEXT,
  id            VARCHAR(100)  PRIMARY KEY,
  name          TEXT          NOT NULL CHECK (name ~ '[\w|-]+'),
  name_tokens   TEXT          NOT NULL,
  updated_at    TIMESTAMPTZ   DEFAULT NOW(),
  user_id       VARCHAR(100)  NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS study_unique__user_id_name__key
  ON study (user_id, LOWER(name));
CREATE INDEX IF NOT EXISTS study_user_id_advanced_at_idx
  ON study (user_id, advanced_at);
CREATE INDEX IF NOT EXISTS study_user_id_updated_at_idx
  ON study (user_id, updated_at);

CREATE OR REPLACE FUNCTION inc_study_lesson_number()
  RETURNS trigger AS
$BODY$
DECLARE
  cnt INT;
BEGIN
  SELECT INTO cnt COUNT(*)::INT
    FROM lesson
    WHERE study_id = NEW.study_id;
  NEW.number = cnt + 1;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dec_study_lesson_number()
  RETURNS trigger AS
$BODY$
BEGIN
  UPDATE lesson
  SET number = number - 1
  WHERE study_id = OLD.study_id AND number > OLD.number;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION check_study_lesson_number()
  RETURNS trigger AS
$BODY$
DECLARE
  cnt INT;
BEGIN
  SELECT INTO cnt COUNT(*)::INT
    FROM lesson
    WHERE study_id = NEW.study_id;
  IF NEW.number > cnt THEN
    NEW.number = cnt;
  END IF;

  RETURN NEW;
END
$BODY$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION move_study_lesson_number()
  RETURNS trigger AS
$BODY$
BEGIN
  IF NEW.number > OLD.number THEN
    UPDATE lesson
    SET number = number - 1
    WHERE study_id = NEW.study_id 
      AND number <= NEW.number
      AND id != NEW.id;
  ELSIF NEW.number < OLD.number THEN
    UPDATE lesson
    SET number = number + 1
    WHERE study_id = NEW.study_id
      AND number >= NEW.number
      AND id != NEW.id;
  END IF;

  RETURN NEW;
END;
$BODY$ LANGUAGE PLPGSQL;

CREATE TABLE IF NOT EXISTS lesson(
  created_at      TIMESTAMPTZ  DEFAULT NOW(),
  body            TEXT,
  id              VARCHAR(100) PRIMARY KEY,
  number          INT          CHECK(number > 0),
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,    
  title           TEXT         NOT NULL,
  title_tokens    TEXT         NOT NULL,
  updated_at      TIMESTAMPTZ  DEFAULT NOW(),
  user_id         VARCHAR(100) NOT NULL,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_user_id_study_id_number_idx
  ON lesson (user_id, study_id, number);
CREATE INDEX IF NOT EXISTS lesson_user_id_study_id_published_at_idx
  ON lesson (user_id, study_id, published_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS lesson_user_id_study_id_updated_at_idx
  ON lesson (user_id, study_id, updated_at);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'insert_lesson_number'
) THEN
  CREATE TRIGGER insert_lesson_number
    BEFORE INSERT ON lesson
    FOR EACH ROW EXECUTE PROCEDURE inc_study_lesson_number(); 
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'delete_lesson_number'
) THEN
  CREATE TRIGGER delete_lesson_number
    AFTER DELETE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE dec_study_lesson_number();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'before_update_lesson_number'
) THEN
  CREATE TRIGGER before_update_lesson_number
    BEFORE UPDATE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE check_study_lesson_number();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'after_update_lesson_number'
) THEN
  CREATE TRIGGER after_update_lesson_number
    AFTER UPDATE ON lesson
    FOR EACH ROW WHEN (pg_trigger_depth() = 0)
    EXECUTE PROCEDURE move_study_lesson_number();
END IF;
END;
$$ language 'plpgsql';

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson'
    AND trigger_name = 'lesson_updated_at_modtime'
) THEN
  CREATE TRIGGER lesson_updated_at_modtime
    BEFORE UPDATE ON lesson
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_comment(
  body            TEXT,
  created_at      TIMESTAMPTZ   DEFAULT NOW(),
  id              VARCHAR(100) PRIMARY KEY,
  lesson_id       VARCHAR(100) NOT NULL,
  published_at    TIMESTAMPTZ,
  study_id        VARCHAR(100) NOT NULL,
  user_id         VARCHAR(100) NOT NULL,
  updated_at      TIMESTAMPTZ   DEFAULT NOW(),
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS lesson_comment_user_id_study_id_lesson_id_published_at_idx
  ON lesson_comment (user_id, study_id, lesson_id, published_at ASC NULLS LAST);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'lesson_comment'
    AND trigger_name = 'lesson_comment_updated_at_modtime'
) THEN
  CREATE TRIGGER lesson_comment_updated_at_modtime
    BEFORE UPDATE ON lesson_comment
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS label(
  color       TEXT         NOT NULL,
  created_at  TIMESTAMPTZ  DEFAULT NOW(),
  is_default  BOOLEAN      DEFAULT FALSE,
  description TEXT,
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40)  NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT NOW()
); 

CREATE UNIQUE INDEX IF NOT EXISTS label_unique__name__idx
  ON label (LOWER(name));

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'label'
    AND trigger_name = 'label_updated_at_modtime'
) THEN
  CREATE TRIGGER label_updated_at_modtime
    BEFORE UPDATE ON label
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS lesson_label(
  created_at TIMESTAMPTZ  DEFAULT NOW(),
  label_id   VARCHAR(100),
  lesson_id  VARCHAR(100),
  study_id   VARCHAR(100),
  PRIMARY KEY (study_id, lesson_id, label_id),
  FOREIGN KEY (label_id)
    REFERENCES label (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (lesson_id)
    REFERENCES lesson (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS topic(
  created_at  TIMESTAMPTZ  DEFAULT NOW(),
  description TEXT,
  id          VARCHAR(100) PRIMARY KEY,
  name        VARCHAR(40)  NOT NULL CHECK(name ~ '^[a-zA-Z0-9][a-zA-Z0-9|-]+[a-zA-Z0-9]$'),
  name_tokens TEXT         NOT NULL,
  updated_at  TIMESTAMPTZ  DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS topic_unique__name__idx
  ON topic (LOWER(name));

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'topic'
    AND trigger_name = 'topic_updated_at_modtime'
) THEN
  CREATE TRIGGER topic_updated_at_modtime
    BEFORE UPDATE ON topic
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS study_topic(
  related_at TIMESTAMPTZ  DEFAULT NOW(),
  study_id   VARCHAR(100),
  topic_id   VARCHAR(100),
  PRIMARY KEY (study_id, topic_id),
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (topic_id)
    REFERENCES topic (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_asset(
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  id            VARCHAR(100) PRIMARY KEY,
  key           TEXT         NOT NULL,
  name          TEXT         NOT NULL CHECK(name ~ '[\w|-]+'),
  name_tokens   TEXT         NOT NULL,
  original_name TEXT         NOT NULL, 
  published_at  TIMESTAMPTZ,
  size          BIGINT       NOT NULL,
  study_id      VARCHAR(100) NOT NULL,
  subtype       TEXT         NOT NULL,
  type          TEXT         NOT NULL,
  updated_at    TIMESTAMPTZ  DEFAULT NOW(),
  user_id       VARCHAR(100) NOT NULL,
  FOREIGN KEY (user_id)
    REFERENCES account (id)
    ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (study_id)
    REFERENCES study (id)
    ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_unique__user_id_study_id_name__idx
  ON user_asset (user_id, study_id, LOWER(name));
CREATE INDEX IF NOT EXISTS user_asset_user_id_study_id_type_subtype_idx
  ON user_asset (user_id, study_id, type, subtype);
CREATE INDEX IF NOT EXISTS user_asset_user_id_study_id_created_at_idx
  ON user_asset (user_id, study_id, created_at);

DO $$
BEGIN
IF NOT EXISTS(
  SELECT *
    FROM information_schema.triggers
    WHERE event_object_table = 'user_asset'
    AND trigger_name = 'user_asset_updated_at_modtime'
) THEN
  CREATE TRIGGER user_asset_updated_at_modtime
    BEFORE UPDATE ON user_asset
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
END IF;
END;
$$ language 'plpgsql';

CREATE OR REPLACE VIEW user_master AS
SELECT
  account.created_at,
  account.id,
  account.login,
  account.name,
  account.profile,
  email.value public_email,
  account.updated_at
FROM account
LEFT JOIN email ON email.user_id = account.id
  AND email.public = TRUE;

CREATE OR REPLACE VIEW user_credentials AS
SELECT
  backup_email.value backup_email,
  account.id,
  account.login,
  account.password,
  primary_email.value primary_email,
  ARRAY(
    SELECT role.name
    FROM role
    LEFT JOIN user_role ON user_role.user_id = account.id
    WHERE role.id = user_role.role_id
  ) roles
FROM account
JOIN email primary_email ON primary_email.user_id = account.id
  AND primary_email.type = 'PRIMARY'
LEFT JOIN email backup_email ON backup_email.user_id = account.id
  AND backup_email.type = 'BACKUP';

CREATE MATERIALIZED VIEW IF NOT EXISTS user_search_index AS
SELECT
  *,
  setweight(to_tsvector('simple', login), 'A') ||
  setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
  setweight(to_tsvector('simple', coalesce(profile, '')), 'B') ||
  setweight(to_tsvector('simple', coalesce(public_email, '')), 'B') as document
FROM user_master;

CREATE UNIQUE INDEX IF NOT EXISTS user_search_index_id_unique_idx
  ON user_search_index (id);

CREATE INDEX IF NOT EXISTS user_search_index_fts_idx
  ON user_search_index USING gin(document);

CREATE OR REPLACE VIEW email_master AS
SELECT
  email.created_at,
  email.id,
  email.public, 
  email.type,
  email.user_id,
  account.login user_login,
  email.updated_at,
  email.value,
  email.verified_at
FROM email
JOIN account ON account.id = email.user_id;

CREATE OR REPLACE VIEW study_master AS
SELECT
  study.advanced_at,
  study.created_at,
  study.description,
  study.id,
  study.name,
  study.updated_at,
  study.user_id,
  account.login user_login
FROM study
JOIN account ON account.id = study.user_id;

CREATE MATERIALIZED VIEW IF NOT EXISTS study_search_index AS
SELECT
  study.advanced_at,
  study.created_at,
  study.description,
  study.id,
  study.name,
  study.updated_at,
  study.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', study.name_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(study.description, '')), 'B') ||
  setweight(to_tsvector('simple', account.login), 'C') ||
  setweight(to_tsvector('simple', coalesce(string_agg(topic.name, ' '), '')), 'A') as document
FROM study
JOIN account ON account.id = study.user_id
LEFT JOIN study_topic ON study_topic.study_id = study.id
LEFT JOIN topic ON topic.id = study_topic.topic_id
GROUP BY study.id, account.id;

CREATE UNIQUE INDEX IF NOT EXISTS study_search_index_id_unique_idx
  ON study_search_index (id);

CREATE INDEX IF NOT EXISTS study_search_index_fts_idx
  ON study_search_index USING gin(document);

CREATE OR REPLACE VIEW lesson_master AS
SELECT
  lesson.body,
  lesson.created_at,
  lesson.id,
  lesson.number,
  lesson.published_at,
  lesson.study_id,
  study.name study_name,
  lesson.title,
  lesson.updated_at,
  lesson.user_id,
  account.login user_login
FROM lesson
JOIN study ON study.id = lesson.study_id
JOIN account ON account.id = lesson.user_id;

CREATE MATERIALIZED VIEW IF NOT EXISTS lesson_search_index AS
SELECT
  lesson.body,
  lesson.created_at,
  lesson.id,
  lesson.number,
  lesson.published_at,
  lesson.study_id,
  study.name study_name,
  lesson.title,
  lesson.updated_at,
  lesson.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', lesson.title_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(lesson.body, '')), 'B') ||
  setweight(to_tsvector('simple', study.name_tokens), 'C') ||
  setweight(to_tsvector('simple', account.login), 'C') ||
  setweight(to_tsvector('simple', coalesce(string_agg(label.name, ' '), '')), 'A') as document
FROM lesson
JOIN study ON study.id = lesson.study_id
JOIN account ON account.id = lesson.user_id
LEFT JOIN lesson_label ON lesson_label.lesson_id = lesson.id
LEFT JOIN label ON label.id = lesson_label.label_id
GROUP BY lesson.id, study.id, account.id;

CREATE UNIQUE INDEX IF NOT EXISTS lesson_search_index_id_unique_idx
  ON lesson_search_index (id);

CREATE INDEX IF NOT EXISTS lesson_search_index_fts_idx
  ON lesson_search_index USING gin(document);

CREATE OR REPLACE VIEW lesson_comment_master AS
SELECT
  lesson_comment.body,
  lesson_comment.created_at,
  lesson_comment.id,
  lesson_comment.lesson_id,
  lesson.number lesson_number,
  lesson_comment.published_at,
  lesson_comment.study_id,
  study.name study_name,
  lesson_comment.updated_at,
  lesson_comment.user_id,
  account.login user_login
FROM lesson_comment
JOIN lesson ON lesson.id = lesson_comment.lesson_id
JOIN study ON study.id = lesson_comment.study_id
JOIN account ON account.id = lesson_comment.user_id;

CREATE OR REPLACE VIEW topic_master AS
SELECT
  created_at,
  description,
  id,
  name,
  updated_at
FROM topic;

CREATE OR REPLACE VIEW study_topic_master AS
SELECT
  topic.created_at,
  topic.description,
  study_topic.topic_id id,
  topic.name,
  study_topic.related_at,
  study_topic.study_id,
  topic.updated_at
FROM study_topic
JOIN topic ON topic.id = study_topic.topic_id;

CREATE MATERIALIZED VIEW IF NOT EXISTS topic_search_index AS
SELECT
  created_at,
  description,
  id,
  name,
  updated_at,
  setweight(to_tsvector('simple', name_tokens), 'A') ||
  setweight(to_tsvector('english', coalesce(description, '')), 'B') as document
FROM topic;

CREATE UNIQUE INDEX IF NOT EXISTS topic_search_index_id_unique_idx
  ON topic_search_index (id);

CREATE INDEX IF NOT EXISTS topic_search_index_fts_idx
  ON topic_search_index USING gin(document);

CREATE OR REPLACE VIEW user_asset_master AS
SELECT
  user_asset.created_at,
  user_asset.id,
  user_asset.key,
  user_asset.name,
  user_asset.original_name,
  user_asset.published_at,
  user_asset.size,
  user_asset.study_id,
  study.name study_name,
  user_asset.subtype,
  user_asset.type,
  user_asset.updated_at,
  user_asset.user_id,
  account.login user_login
FROM user_asset
JOIN study ON study.id = user_asset.study_id
JOIN account ON account.id = user_asset.user_id;

CREATE MATERIALIZED VIEW IF NOT EXISTS user_asset_search_index AS
SELECT
  user_asset.created_at,
  user_asset.id,
  user_asset.key,
  user_asset.name,
  user_asset.original_name,
  user_asset.published_at,
  user_asset.size,
  user_asset.study_id,
  study.name study_name,
  user_asset.subtype,
  user_asset.type,
  user_asset.updated_at,
  user_asset.user_id,
  account.login user_login,
  setweight(to_tsvector('simple', user_asset.name_tokens), 'A') ||
  setweight(to_tsvector('simple', user_asset.type), 'A') ||
  setweight(to_tsvector('simple', user_asset.subtype), 'C') ||
  setweight(to_tsvector('simple', study.name_tokens), 'C') ||
  setweight(to_tsvector('simple', account.login), 'C') AS document
FROM user_asset
JOIN study ON study.id = user_asset.study_id
JOIN account ON account.id = user_asset.user_id
GROUP BY user_asset.id, study.id, account.id;

CREATE UNIQUE INDEX IF NOT EXISTS user_asset_search_index_id_unique_idx
  ON user_asset_search_index (id);

CREATE INDEX IF NOT EXISTS user_asset_search_index_fts_idx
  ON user_asset_search_index USING gin(document);
`
