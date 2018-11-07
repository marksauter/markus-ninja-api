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
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error("failed to get topics")
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
	FROM topic
	WHERE id = $1
`

func GetTopic(
	db Queryer,
	id string,
) (*Topic, error) {
	mylog.Log.WithField("id", id).Info("GetTopic(id)")
	return getTopic(db, "getTopicByID", getTopicByIDSQL, id)
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
) (names []string, err error) {
	mylog.Log.WithField(
		"topicable_id", topicableID,
	).Info("GetNamesByTopicable(topicable_id)")
	topicNames := pgtype.TextArray{}
	err = prepareQueryRow(
		db,
		"getTopicNamesByTopicable",
		getTopicNamesByTopicableSQL,
		topicableID,
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
	topicableID string,
	po *PageOptions,
	filters *TopicFilterOptions,
) ([]*Topic, error) {
	mylog.Log.WithField(
		"topicable_id", topicableID,
	).Info("GetTopicByTopicable(topicable_id)")
	var rows []*Topic
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Topic, 0, limit)
		} else {
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
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, topicDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
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
		mylog.Log.WithError(err).Error("failed to create topic")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
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

	return rows, nil
}

func UpdateTopic(
	db Queryer,
	row *Topic,
) (*Topic, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateTopic(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	if len(sets) == 0 {
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
		return nil, ErrNotFound
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

	return topic, nil
}
