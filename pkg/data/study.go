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
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.applee_id = ` + args.Append(appleeID)
	}
	from := "appled_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByApplee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("studies found"))
	}
	return n, err
}

func CountStudyByEnrollee(
	db Queryer,
	enrolleeID string,
	filters *StudyFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}
	from := "enrolled_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByEnrollee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("studies found"))
	}
	return n, err
}

func CountStudyByTopic(
	db Queryer,
	topicID string,
	filters *StudyFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}
	from := "topiced_study"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByTopic", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("studies found"))
	}
	return n, err
}

func CountStudyByUser(
	db Queryer,
	userID string,
	filters *StudyFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "study_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("studies found"))
	}
	return n, err
}

func CountStudyBySearch(
	db Queryer,
	filters *StudyFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "study_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countStudyBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("studies found"))
	}
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

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
		mylog.Log.WithError(err).Debug(util.Trace(""))
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
	FROM study_search_index
	WHERE id = $1
`

func GetStudy(
	db Queryer,
	id string,
) (*Study, error) {
	study, err := getStudy(db, "getStudyByID", getStudyByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("study found"))
	}
	return study, err
}

func GetStudyByApplee(
	db Queryer,
	appleeID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("studies found"))
	return rows, nil
}

func GetStudyByEnrollee(
	db Queryer,
	enrolleeID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("studies found"))
	return rows, nil
}

func GetStudyByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("studies found"))
	return rows, nil
}

func GetStudyByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
	FROM study_search_index
	WHERE user_id = $1 AND lower(name) = lower($2)
`

func GetStudyByName(
	db Queryer,
	userID,
	name string,
) (*Study, error) {
	study, err := getStudy(db, "getStudyByName", getStudyByNameSQL, userID, name)
	if err != nil {
		mylog.Log.WithField("name", name).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("name", name).Info(util.Trace("study found"))
	}
	return study, err
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
	FROM study_search_index s
	JOIN account a ON lower(a.login) = lower($1)
	WHERE s.user_id = a.id AND lower(s.name) = lower($2)  
`

func GetStudyByUserAndName(
	db Queryer,
	owner,
	name string,
) (*Study, error) {
	study, err := getStudy(db, "getStudyByUserAndName", getStudyByUserAndNameSQL, owner, name)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"owner": owner,
			"name":  name,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"owner": owner,
			"name":  name,
		}).Info(util.Trace("study found"))
	}
	return study, err
}

func CreateStudy(
	db Queryer,
	row *Study,
) (*Study, error) {
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, DuplicateFieldError(pgErr.ConstraintName)
			default:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := NewStudyCreatedPayload(&row.ID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	e, err := NewStudyEvent(eventPayload, &row.ID, &row.UserID, true)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err = CreateEvent(tx, e); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	study, err := GetStudy(tx, row.ID.String)
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

	mylog.Log.Info(util.Trace("study created"))
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
	commandTag, err := prepareExec(db, "deleteStudy", deleteStudySQL, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("study deleted"))
	return nil
}

func SearchStudy(
	db Queryer,
	po *PageOptions,
	filters *StudyFilterOptions,
) ([]*Study, error) {
	var rows []*Study
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Study, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

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

	psName := preparedName("searchStudyIndex", sql)

	if err := getManyStudy(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("studies found"))
	return rows, nil
}

func UpdateStudy(
	db Queryer,
	row *Study,
) (*Study, error) {
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
		mylog.Log.Info(util.Trace("no updates"))
		return GetStudy(db, row.ID.String)
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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, DuplicateFieldError(pgErr.ConstraintName)
			default:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	study, err := GetStudy(tx, row.ID.String)
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("study updated"))
	return study, nil
}
