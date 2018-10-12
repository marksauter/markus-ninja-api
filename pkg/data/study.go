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

type Study struct {
	AdvancedAt  pgtype.Timestamptz `db:"advanced_at" permit:"read"`
	AppledAt    pgtype.Timestamptz `db:"appled_at"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	EnrolledAt  pgtype.Timestamptz `db:"enrolled_at"`
	ID          mytype.OID         `db:"id" permit:"read"`
	Name        mytype.WordsName   `db:"name" permit:"create/read"`
	Private     pgtype.Bool        `db:"private" permit:"create/read/update"`
	TopicedAt   pgtype.Timestamptz `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID      mytype.OID         `db:"user_id" permit:"create/read"`
}

func studyDelimeter(r rune) bool {
	return r == '-' || r == '_'
}

type StudyFilterOptions struct {
	Topics *[]string
	Search *string
}

func (src *StudyFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 2)
	if src.Topics != nil && len(*src.Topics) > 0 {
		query := ToTsQuery(strings.Join(*src.Topics, " "))
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS topics_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".topics @@ topics_query END",
		)
	}
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

func CountStudyByApplee(
	db Queryer,
	appleeID string,
	filters *StudyFilterOptions,
) (int32, error) {
	mylog.Log.WithField("applee_id", appleeID).Info("CountStudyByApplee(applee_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.applee_id = ` + args.Append(appleeID)
	}
	from := "appled_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByApplee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountStudyByEnrollee(
	db Queryer,
	enrolleeID string,
	filters *StudyFilterOptions,
) (int32, error) {
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("CountStudyByEnrollee(enrollee_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}
	from := "enrolled_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByEnrollee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountStudyByTopic(
	db Queryer,
	topicID string,
	filters *StudyFilterOptions,
) (int32, error) {
	mylog.Log.WithField(
		"topic_id", topicID,
	).Info("CountStudyByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}
	from := "topiced_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByTopic", sql)

	mylog.Log.Debug(sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountStudyByUser(
	db Queryer,
	userID string,
	filters *StudyFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountStudyByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "study_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

const countStudyBySearchSQL = `
	SELECT COUNT(*)
	FROM study_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE document @@ query END)
`

func CountStudyBySearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountStudyBySearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countStudyBySearch",
		countStudyBySearchSQL,
		ToPrefixTsQuery(query),
	).Scan(&n)
	return n, err
}

const countStudyByTopicSearchSQL = `
	SELECT COUNT(*)
	FROM study_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE topics @@ query END)
`

func CountStudyByTopicSearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountStudyBySearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countStudyByTopicSearch",
		countStudyByTopicSearchSQL,
		ToTsQuery(query),
	).Scan(&n)
	return n, err
}

func getStudy(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Study, error) {
	var row Study
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.ID,
		&row.Name,
		&row.Private,
		&row.UpdatedAt,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get study")
		return nil, err
	}

	return &row, nil
}

func getManyStudy(
	db Queryer,
	name string,
	sql string,
	rows *[]*Study,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Private,
			&row.UpdatedAt,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return err
	}

	return nil
}

const getStudyByIDSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		private,
		updated_at,
		user_id
	FROM study
	WHERE id = $1
`

func GetStudy(
	db Queryer,
	id string,
) (*Study, error) {
	mylog.Log.WithField("id", id).Info("GetStudy(id)")
	return getStudy(db, "getStudyByID", getStudyByIDSQL, id)
}

func GetStudyByApplee(
	db Queryer,
	appleeID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	mylog.Log.WithField("applee_id", appleeID).Info("GetStudyByApplee(applee_id)")
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.applee_id = ` + args.Append(appleeID)
	}

	selects := []string{
		"advanced_at",
		"appled_at",
		"created_at",
		"description",
		"id",
		"name",
		"private",
		"updated_at",
		"user_id",
	}
	from := "appled_study"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getStudiesByAppled", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.AppledAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Private,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	return rows, nil
}

func GetStudyByEnrollee(
	db Queryer,
	enrolleeID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("GetStudyByEnrollee(enrollee_id)")
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"enrolled_at",
		"id",
		"name",
		"private",
		"updated_at",
		"user_id",
	}
	from := "enrolled_study"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getStudiesByEnrollee", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.EnrolledAt,
			&row.ID,
			&row.Name,
			&row.Private,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	return rows, nil
}

func GetStudyByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	mylog.Log.WithField("topic_id", topicID).Info("GetStudyByTopic(topic_id)")
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"private",
		"topiced_at",
		"updated_at",
		"user_id",
	}
	from := "topiced_study"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getStudiesByTopic", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Private,
			&row.TopicedAt,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	return rows, nil
}

func GetStudyByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userID).Info("GetStudyByUser(user_id)")
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"private",
		"updated_at",
		"user_id",
	}
	from := "study_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getStudiesByUserID", sql)

	if err := getManyStudy(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

const getStudyByNameSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		private,
		updated_at,
		user_id
	FROM study
	WHERE user_id = $1 AND lower(name) = lower($2)
`

func GetStudyByName(
	db Queryer,
	userID,
	name string,
) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"name":    name,
	}).Info("GetStudyByName(user_id, name)")
	return getStudy(db, "getStudyByName", getStudyByNameSQL, userID, name)
}

const getStudyByUserAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.private,
		s.updated_at,
		s.user_id
	FROM study s
	JOIN account a ON lower(a.login) = lower($1)
	WHERE s.user_id = a.id AND lower(s.name) = lower($2)  
`

func GetStudyByUserAndName(
	db Queryer,
	owner,
	name string,
) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"owner": owner,
		"name":  name,
	}).Info("GetStudyByUserAndName(owner, name)")
	return getStudy(db, "getStudyByUserAndName", getStudyByUserAndNameSQL, owner, name)
}

func CreateStudy(
	db Queryer,
	row *Study,
) (*Study, error) {
	mylog.Log.Info("CreateStudy()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Study")
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
		nameTokens.Set(strings.Join(util.Split(row.Name.String, studyDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.Private.Status != pgtype.Undefined {
		columns = append(columns, "private")
		values = append(values, args.Append(&row.Private))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
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
		INSERT INTO study(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createStudy", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create study")
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

	eventPayload, err := NewStudyCreatedPayload(&row.ID)
	if err != nil {
		return nil, err
	}
	e, err := NewStudyEvent(eventPayload, &row.ID, &row.UserID)
	if err != nil {
		return nil, err
	}
	if _, err = CreateEvent(tx, e); err != nil {
		return nil, err
	}

	study, err := GetStudy(tx, row.ID.String)
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

	return study, nil
}

const deleteStudySQL = `
	DELETE FROM study
	WHERE id = $1
`

func DeleteStudy(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteStudy(id)")
	commandTag, err := prepareExec(db, "deleteStudy", deleteStudySQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func SearchStudy(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("query", query).Info("SearchStudy(query)")
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			return rows, nil
		}
	}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"private",
		"updated_at",
		"user_id",
	}
	from := "study_search_index"
	var args pgx.QueryArgs
	sql := SearchSQL2(selects, from, ToPrefixTsQuery(query), &args, po)

	psName := preparedName("searchStudyIndex", sql)

	if err := getManyStudy(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func UpdateStudy(
	db Queryer,
	row *Study,
) (*Study, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateStudy(id)")
	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, studyDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}
	if row.Private.Status != pgtype.Undefined {
		sets = append(sets, `private`+"="+args.Append(&row.Private))
	}

	if len(sets) == 0 {
		return GetStudy(db, row.ID.String)
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
		UPDATE study
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateStudy", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update study")
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
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	study, err := GetStudy(tx, row.ID.String)
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

	return study, nil
}
