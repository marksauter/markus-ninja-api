package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Topic struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	ID          mytype.OID         `db:"id" permit:"read"`
	Name        mytype.WordName    `db:"name" permit:"create/read"`
	TopicableID mytype.OID         `db:"topicable_id"`
	TopicedAt   pgtype.Timestamptz `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func topicDelimeter(r rune) bool {
	return r == '-'
}

type TopicFilterOptions struct {
	Search *string
}

func (src *TopicFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 3)
	if src.Search != nil {
		query := ToPrefixTsQuery(*src.Search)
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS document_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".document @@ document_query END",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		From:  strings.Join(fromParts, ", "),
		Where: where,
	}
}

func CountTopicByTopicable(
	db Queryer,
	topicableID string,
	filters *TopicFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topicable_id = ` + args.Append(topicableID)
	}
	from := "topicable_topic"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countTopicByTopicable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("topics found"))
	}
	return n, err
}

func CountTopicBySearch(
	db Queryer,
	filters *TopicFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "topic_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countTopicBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("topics found"))
	}
	return n, err
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
		&row.ID,
		&row.Name,
		&row.UpdatedAt,
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

func getManyTopic(
	db Queryer,
	name string,
	sql string,
	rows *[]*Topic,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.UpdatedAt,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getTopicByIDSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic_search_index
	WHERE id = $1
`

func GetTopic(
	db Queryer,
	id string,
) (*Topic, error) {
	topic, err := getTopic(db, "getTopicByID", getTopicByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("topic found"))
	}
	return topic, err
}

const getTopicByNameSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic_search_index
	WHERE LOWER(name) = LOWER($1)
`

func GetTopicByName(
	db Queryer,
	name string,
) (*Topic, error) {
	topic, err := getTopic(db, "getTopicByName", getTopicByNameSQL, name)
	if err != nil {
		mylog.Log.WithField("name", name).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("name", name).Info(util.Trace("topic found"))
	}
	return topic, err
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
	topicableID string,
) ([]string, error) {
	names := []string{}
	topicNames := pgtype.TextArray{}
	err := prepareQueryRow(
		db,
		"getTopicNamesByTopicable",
		getTopicNamesByTopicableSQL,
		topicableID,
	).Scan(topicNames)
	if err == pgx.ErrNoRows {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return names, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return names, err
	}

	err = topicNames.AssignTo(names)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return names, err
	}

	mylog.Log.WithField("n", len(names)).Info(util.Trace("topic names found"))
	return names, nil
}

func GetTopicByTopicable(
	db Queryer,
	topicableID string,
	po *PageOptions,
	filters *TopicFilterOptions,
) ([]*Topic, error) {
	var rows []*Topic
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Topic, 0, limit)
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
		"description",
		"id",
		"name",
		"topicable_id",
		"topiced_at",
		"updated_at",
	}
	from := "topicable_topic"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getTopicsByTopicableID", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.TopicableID,
			&row.TopicedAt,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("topics found"))
	return rows, nil
}

func CreateTopic(
	db Queryer,
	row *Topic,
) (*Topic, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	var columns, values []string

	id, _ := mytype.NewOID("Topic")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
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
		INSERT INTO topic(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT(lower("name")) DO UPDATE SET name=EXCLUDED.name RETURNING id
	`

	psName := preparedName("createTopic", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.ID,
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if row.TopicableID.Status != pgtype.Undefined {
		topiced := Topiced{}
		topiced.TopicID.Set(&row.ID)
		topiced.TopicableID.Set(&row.TopicableID)
		_, err := CreateTopiced(tx, topiced)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	topic, err := GetTopic(tx, row.ID.String)
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

	mylog.Log.Info(util.Trace("topic created"))
	return topic, nil
}

func SearchTopic(
	db Queryer,
	po *PageOptions,
	filters *TopicFilterOptions,
) ([]*Topic, error) {
	var rows []*Topic
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Topic, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
	}
	from := "topic_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchTopicIndex", sql)

	if err := getManyTopic(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("topics found"))
	return rows, nil
}

func UpdateTopic(
	db Queryer,
	row *Topic,
) (*Topic, error) {
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetTopic(db, row.ID.String)
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
		UPDATE topic
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateTopic", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	topic, err := GetTopic(tx, row.ID.String)
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("topic updated"))
	return topic, nil
}
