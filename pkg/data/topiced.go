package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Topiced struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID          pgtype.Int4        `db:"id" permit:"read"`
	TopicID     mytype.OID         `db:"topic_id" permit:"read"`
	TopicableID mytype.OID         `db:"topicable_id" permit:"read"`
	Type        pgtype.Text        `db:"type" permit:"read"`
}

const countTopicedByTopicSQL = `
	SELECT COUNT(*)
	FROM topiced
	WHERE topic_id = $1
`

func CountTopicedByTopic(
	db Queryer,
	topicID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countTopicedByTopic",
		countTopicedByTopicSQL,
		topicID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("topiceds found"))
	}
	return n, err
}

const countTopicedByTopicableSQL = `
	SELECT COUNT(*)
	FROM topiced
	WHERE topicable_id = $1
`

func CountTopicedByTopicable(
	db Queryer,
	topicableID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countTopicedByTopicable",
		countTopicedByTopicableSQL,
		topicableID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("topiceds found"))
	}
	return n, err
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
		&row.ID,
		&row.TopicID,
		&row.TopicableID,
		&row.Type,
	)
	if err == pgx.ErrNoRows {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

func getManyTopiced(
	db Queryer,
	name string,
	sql string,
	rows *[]*Topiced,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Topiced
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.TopicID,
			&row.TopicableID,
			&row.Type,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getTopicedSQL = `
	SELECT
		created_at,
		id,
		topic_id,
		topicable_id,
		type
	FROM topiced
	WHERE id = $1
`

func GetTopiced(
	db Queryer,
	id int32,
) (*Topiced, error) {
	topiced, err := getTopiced(db, "getTopiced", getTopicedSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("topiced found"))
	}
	return topiced, err
}

const getTopicedByTopicableAndTopicSQL = `
	SELECT
		created_at,
		id,
		topic_id,
		topicable_id,
		type
	FROM topiced
	WHERE topicable_id = $1 AND topic_id = $2
`

func GetTopicedByTopicableAndTopic(
	db Queryer,
	topicableID,
	topicID string,
) (*Topiced, error) {
	topiced, err := getTopiced(
		db,
		"getTopicedByTopicableAndTopic",
		getTopicedByTopicableAndTopicSQL,
		topicableID,
		topicID,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"topicable_id": topicableID,
			"topic_id":     topicID,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"topicable_id": topicableID,
			"topic_id":     topicID,
		}).Info(util.Trace("topiced found"))
	}
	return topiced, err
}

func GetTopicedByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
) ([]*Topiced, error) {
	var rows []*Topiced
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Topiced, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}

	selects := []string{
		"created_at",
		"id",
		"topic_id",
		"topicable_id",
		"type",
	}
	from := "topiced"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getTopicedsByTopic", sql)

	if err := getManyTopiced(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("topiceds found"))
	return rows, nil
}

func GetTopicedByTopicable(
	db Queryer,
	topicableID string,
	po *PageOptions,
) ([]*Topiced, error) {
	var rows []*Topiced
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Topiced, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topicable_id = ` + args.Append(topicableID)
	}

	selects := []string{
		"created_at",
		"id",
		"topic_id",
		"topicable_id",
		"type",
	}
	from := "topiced"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getTopicedsByTopicable", sql)

	if err := getManyTopiced(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("topiceds found"))
	return rows, nil
}

func CreateTopiced(
	db Queryer,
	row Topiced,
) (*Topiced, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.TopicID.Status != pgtype.Undefined {
		columns = append(columns, "topic_id")
		values = append(values, args.Append(&row.TopicID))
	}
	if row.TopicableID.Status != pgtype.Undefined {
		columns = append(columns, "topicable_id")
		values = append(values, args.Append(&row.TopicableID))
		columns = append(columns, "type")
		values = append(values, args.Append(row.TopicableID.Type))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO topiced(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createTopiced", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	topiced, err := GetTopicedByTopicableAndTopic(
		tx,
		row.TopicableID.String,
		row.TopicID.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.Info(util.Trace("topiced created"))
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
	commandTag, err := prepareExec(
		db,
		"deleteTopiced",
		deleteTopicedSQL,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("topiced deleted"))
	return nil
}

const deleteTopicedByTopicableAndTopicSQL = `
	DELETE FROM topiced
	WHERE topicable_id = $1 AND topic_id = $2
`

func DeleteTopicedByTopicableAndTopic(
	db Queryer,
	topicableID,
	topicID string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteTopicedFromTopicable",
		deleteTopicedByTopicableAndTopicSQL,
		topicableID,
		topicID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"topicable_id": topicableID,
			"topic_id":     topicID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"topicable_id": topicableID,
		"topic_id":     topicID,
	}).Info(util.Trace("topiced deleted"))
	return nil
}
