package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Topic struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"create/read"`
	TopicableId mytype.OID         `db:"topicable_id"`
	TopicedAt   pgtype.Timestamptz `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

const countTopicByTopicableSQL = `
	SELECT COUNT(*)
	FROM topicable_topic
	WHERE topicable_id = $1
`

func CountTopicByTopicable(
	db Queryer,
	topicableId string,
) (int32, error) {
	mylog.Log.WithField("topicable_id", topicableId).Info("CountTopicByTopicable(topicable_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countTopicByTopicable",
		countTopicByTopicableSQL,
		topicableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func CountTopicBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountTopicBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM topic_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToPrefixTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "Topicable" {
			// Only studies 'contain' topics, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND topic_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countTopicBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getTopic(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Topic, error) {
	var row Topic
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get topic")
		return nil, err
	}

	return &row, nil
}

func getManyTopic(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Topic, error) {
	var rows []*Topic

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get topics")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getTopicByIdSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic
	WHERE id = $1
`

func GetTopic(
	db Queryer,
	id string,
) (*Topic, error) {
	mylog.Log.WithField("id", id).Info("GetTopic(id)")
	return getTopic(db, "getTopicById", getTopicByIdSQL, id)
}

const getTopicNamesByTopicableSQL = `
	SELECT
		array_agg(name) topic_names
	FROM topicable_topic
	WHERE topicable_id = $1
	GROUP BY topicable_id
`

func GetNamesByTopicable(
	db Queryer,
	topicableId string,
) (names []string, err error) {
	mylog.Log.WithField(
		"topicable_id", topicableId,
	).Info("GetNamesByTopicable(topicable_id)")
	topicNames := pgtype.TextArray{}
	err = prepareQueryRow(
		db,
		"getTopicNamesByTopicable",
		getTopicNamesByTopicableSQL,
		topicableId,
	).Scan(topicNames)
	if err == pgx.ErrNoRows {
		return
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get topic")
		return
	}

	err = topicNames.AssignTo(names)
	return
}

func GetTopicByTopicable(
	db Queryer,
	topicableId string,
	po *PageOptions,
) ([]*Topic, error) {
	mylog.Log.WithField(
		"topicable_id", topicableId,
	).Info("GetTopicByTopicable(topicable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topicable_id = ` + args.Append(topicableId)}

	selects := []string{
		"created_at",
		"description",
		"id",
		"name",
		"topicable_id",
		"topiced_at",
		"updated_at",
	}
	from := "topicable_topic"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getTopicsByTopicableId", sql)

	var rows []*Topic

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.TopicableId,
			&row.TopicedAt,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get topics")
		return nil, err
	}

	return rows, nil
}

const getTopicByNameSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic
	WHERE LOWER(name) = LOWER($1)
`

func GetTopicByName(
	db Queryer,
	name string,
) (*Topic, error) {
	mylog.Log.WithFields(logrus.Fields{
		"name": name,
	}).Info("GetTopicByName(user_id, name)")
	return getTopic(db, "getTopicByName", getTopicByNameSQL, name)
}

func CreateTopic(
	db Queryer,
	row *Topic,
) (*Topic, error) {
	mylog.Log.Info("CreateTopic()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Topic")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, topicDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO topic(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT(lower("name")) DO UPDATE SET name=EXCLUDED.name RETURNING id
	`

	psName := preparedName("createTopic", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create topic")
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

	if row.TopicableId.Status != pgtype.Undefined {
		topiced := Topiced{}
		topiced.TopicId.Set(&row.Id)
		topiced.TopicableId.Set(&row.TopicableId)
		_, err := CreateTopiced(tx, topiced)
		if err != nil {
			return nil, err
		}
	}

	topic, err := GetTopic(tx, row.Id.String)
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

	return topic, nil
}

const refreshTopicSearchIndexSQL = `
	SELECT refresh_mv_xxx('topic_search_index')
`

func RefreshTopicSearchIndex(
	db Queryer,
) error {
	mylog.Log.Info("RefreshTopicSearchIndex()")
	_, err := prepareExec(
		db,
		"refreshTopicSearchIndex",
		refreshTopicSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func SearchTopic(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*Topic, error) {
	mylog.Log.WithField("query", query).Info("SearchTopic(query)")
	selects := []string{
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
	}
	from := "topic_search_index"
	var args pgx.QueryArgs
	sql := SearchSQL(selects, from, nil, ToPrefixTsQuery(query), "document", po, &args)

	psName := preparedName("searchTopicIndex", sql)

	return getManyTopic(db, psName, sql, args...)
}

func UpdateTopic(
	db Queryer,
	row *Topic,
) (*Topic, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateTopic(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	if len(sets) == 0 {
		return GetTopic(db, row.Id.String)
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		UPDATE topic
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateTopic", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	topic, err := GetTopic(tx, row.Id.String)
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

	return topic, nil
}

func topicDelimeter(r rune) bool {
	return r == '-'
}
