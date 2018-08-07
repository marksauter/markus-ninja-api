package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Topiced struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          pgtype.Int4        `db:"id" permit:"read"`
	TopicId     mytype.OID         `db:"topic_id" permit:"read"`
	TopicableId mytype.OID         `db:"topicable_id" permit:"read"`
}

const countTopicedByTopicSQL = `
	SELECT COUNT(*)
	FROM topiced
	WHERE topic_id = $1
`

func CountTopicedByTopic(
	db Queryer,
	topicId string,
) (n int32, err error) {
	mylog.Log.WithField("topic_id", topicId).Info("CountTopicedByTopic()")

	err = prepareQueryRow(
		db,
		"countTopicedByTopic",
		countTopicedByTopicSQL,
		topicId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countTopicedByTopicableSQL = `
	SELECT COUNT(*)
	FROM topiced
	WHERE topicable_id = $1
`

func CountTopicedByTopicable(
	db Queryer,
	topicableId string,
) (n int32, err error) {
	mylog.Log.WithField("topicable_id", topicableId).Info("CountTopicedByTopicable()")

	err = prepareQueryRow(
		db,
		"countTopicedByTopicable",
		countTopicedByTopicableSQL,
		topicableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getTopiced(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Topiced, error) {
	var row Topiced
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.TopicId,
		&row.TopicableId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get topiced")
		return nil, err
	}

	return &row, nil
}

func getManyTopiced(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Topiced, error) {
	var rows []*Topiced

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get topiceds")
		return nil, err
	}

	for dbRows.Next() {
		var row Topiced
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.TopicId,
			&row.TopicableId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get topiceds")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getTopicedSQL = `
	SELECT
		created_at,
		id,
		topic_id,
		topicable_id
	FROM topiced
	WHERE id = $1
`

func GetTopiced(
	db Queryer,
	id int32,
) (*Topiced, error) {
	mylog.Log.WithField("id", id).Info("GetTopiced(id)")
	return getTopiced(db, "getTopiced", getTopicedSQL, id)
}

const getTopicedByTopicableAndTopicSQL = `
	SELECT
		created_at,
		id,
		topic_id,
		topicable_id
	FROM topiced
	WHERE topicable_id = $1 AND topic_id = $2
`

func GetTopicedByTopicableAndTopic(
	db Queryer,
	topicableId,
	topicId string,
) (*Topiced, error) {
	mylog.Log.Info("GetTopicedByTopicableAndTopic()")
	return getTopiced(
		db,
		"getTopicedByTopicableAndTopic",
		getTopicedByTopicableAndTopicSQL,
		topicableId,
		topicId,
	)
}

func GetTopicedByTopic(
	db Queryer,
	topicId string,
	po *PageOptions,
) ([]*Topiced, error) {
	mylog.Log.WithField("topic_id", topicId).Info("GetTopicedByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topic_id = ` + args.Append(topicId)}

	selects := []string{
		"created_at",
		"id",
		"topic_id",
		"topicable_id",
	}
	from := "topiced"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getTopicedsByTopic", sql)

	return getManyTopiced(db, psName, sql, args...)
}

func GetTopicedByTopicable(
	db Queryer,
	topicableId string,
	po *PageOptions,
) ([]*Topiced, error) {
	mylog.Log.WithField("topicable_id", topicableId).Info("GetTopicedByTopicable(topicable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topicable_id = ` + args.Append(topicableId)}

	selects := []string{
		"created_at",
		"id",
		"topic_id",
		"topicable_id",
	}
	from := "topiced"
	sql := SQL(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getTopicedsByTopicable", sql)

	return getManyTopiced(db, psName, sql, args...)
}

func CreateTopiced(
	db Queryer,
	row Topiced,
) (*Topiced, error) {
	mylog.Log.Info("CreateTopiced()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.TopicId.Status != pgtype.Undefined {
		columns = append(columns, "topic_id")
		values = append(values, args.Append(&row.TopicId))
	}
	if row.TopicableId.Status != pgtype.Undefined {
		columns = append(columns, "topicable_id")
		values = append(values, args.Append(&row.TopicableId))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	var topicable string
	switch row.TopicableId.Type {
	case "Course":
		topicable = "course"
	case "Study":
		topicable = "study"
	default:
		return nil, fmt.Errorf("invalid type '%s' for topiced topicable id", row.TopicableId.Type)
	}

	table := strings.Join(
		[]string{topicable, "topiced"},
		"_",
	)
	sql := `
		INSERT INTO ` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING topiced_id
	`

	psName := preparedName("createTopiced", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		mylog.Log.WithError(err).Error("failed to create topiced")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		return nil, err
	}

	topiced, err := GetTopicedByTopicableAndTopic(
		tx,
		row.TopicableId.String,
		row.TopicId.String,
	)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return topiced, nil
}

const deleteTopicedSQL = `
	DELETE FROM topiced
	WHERE id = $1
`

func DeleteTopiced(
	db Queryer,
	id int32,
) error {
	mylog.Log.WithField("id", id).Info("DeleteTopiced(id)")
	commandTag, err := prepareExec(
		db,
		"deleteTopiced",
		deleteTopicedSQL,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteTopicedByTopicableAndTopicSQL = `
	DELETE FROM topiced
	WHERE topicable_id = $1 AND topic_id = $2
`

func DeleteTopicedByTopicableAndTopic(
	db Queryer,
	topicable_id,
	topic_id string,
) error {
	mylog.Log.Info("DeleteTopicedByTopicableAndTopic()")
	commandTag, err := prepareExec(
		db,
		"deleteTopicedFromTopicable",
		deleteTopicedByTopicableAndTopicSQL,
		topicable_id,
		topic_id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
