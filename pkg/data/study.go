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

type Study struct {
	AdvancedAt  pgtype.Timestamptz `db:"advanced_at" permit:"read"`
	AppledAt    pgtype.Timestamptz `db:"appled_at"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	EnrolledAt  pgtype.Timestamptz `db:"enrolled_at"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Name        mytype.URLSafeName `db:"name" permit:"create/read"`
	TopicedAt   pgtype.Timestamptz `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"create/read"`
}

const countStudyByAppledSQL = `
	SELECT COUNT(*)
	FROM study_appled
	WHERE applee_id = $1
`

func CountStudyByApplee(
	db Queryer,
	appleeId string,
) (n int32, err error) {
	mylog.Log.WithField("applee_id", appleeId).Info("CountStudyByAppled(applee_id)")
	err = prepareQueryRow(
		db,
		"countStudyByAppled",
		countStudyByAppledSQL,
		appleeId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countStudyByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM study_enrolled
	WHERE enrollee_id = $1
`

func CountStudyByEnrollee(
	db Queryer,
	enrolleeId string,
) (n int32, err error) {
	mylog.Log.WithField("enrollee_id", enrolleeId).Info("CountStudyByEnrollee(enrollee_id)")
	err = prepareQueryRow(
		db,
		"countStudyByEnrollee",
		countStudyByEnrolleeSQL,
		enrolleeId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countStudyByTopicSQL = `
	SELECT COUNT(*)
	FROM topiced_study
	WHERE topic_id = $1
`

func CountStudyByTopic(
	db Queryer,
	topicId string,
) (n int32, err error) {
	mylog.Log.WithField(
		"topic_id", topicId,
	).Info("CountStudyByTopic(topic_id)")
	err = prepareQueryRow(
		db,
		"countStudyByTopic",
		countStudyByTopicSQL,
		topicId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countStudyByUserSQL = `
	SELECT COUNT(*)
	FROM study
	WHERE user_id = $1
`

func CountStudyByUser(
	db Queryer,
	userId string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountStudyByUser(user_id)")
	err = prepareQueryRow(
		db,
		"countStudyByUser",
		countStudyByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountStudyBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountStudyBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM study_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" {
			// Only users 'contain' studies, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND study_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countStudyBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
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
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
		&row.UserId,
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
	args ...interface{},
) ([]*Study, error) {
	var rows []*Study

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getStudyByIdSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
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
	return getStudy(db, "getStudyById", getStudyByIdSQL, id)
}

func GetStudyByApplee(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("GetStudyByAppled(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`applee_id = ` + args.Append(userId)}

	selects := []string{
		"advanced_at",
		"appled_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
	}
	from := "appled_study"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByAppled", sql)

	var rows []*Study

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
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByEnrollee(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("GetStudyByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userId)}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"enrolled_at",
		"id",
		"name",
		"updated_at",
		"user_id",
	}
	from := "enrolled_study"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByEnrollee", sql)

	var rows []*Study

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
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByTopic(
	db Queryer,
	topicId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("topic_id", topicId).Info("GetStudyByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topic_id = ` + args.Append(topicId)}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"topiced_at",
		"updated_at",
		"user_id",
	}
	from := "topiced_study"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByTopic", sql)

	var rows []*Study

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
			&row.Id,
			&row.Name,
			&row.TopicedAt,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("GetStudyByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
	}
	from := "study"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByUserId", sql)

	return getManyStudy(db, psName, sql, args...)
}

const getStudyByNameSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		updated_at,
		user_id
	FROM study
	WHERE user_id = $1 AND lower(name) = lower($2)
`

func GetStudyByName(
	db Queryer,
	userId,
	name string,
) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"name":    name,
	}).Info("GetStudyByName(user_id, name)")
	return getStudy(db, "getStudyByName", getStudyByNameSQL, userId, name)
}

const getStudyByUserAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.updated_at,
		s.user_id
	FROM study s
	INNER JOIN account a ON a.login = $1
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
	mylog.Log.Info("Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Study")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.AdvancedAt.Status != pgtype.Undefined {
		columns = append(columns, "advanced_at")
		values = append(values, args.Append(&row.AdvancedAt))
	}
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
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
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

	study, err := GetStudy(tx, row.Id.String)
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

const refreshStudySearchIndexSQL = `
	SELECT refresh_mv_xxx('study_search_index')
`

func RefreshStudySearchIndex(
	db Queryer,
) error {
	mylog.Log.Info("RefreshStudySearchIndex()")
	_, err := prepareExec(
		db,
		"refreshStudySearchIndex",
		refreshStudySearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func SearchStudy(
	db Queryer,
	within *mytype.OID,
	query string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("query", query).Info("SearchStudy(query)")
	if within != nil {
		if within.Type != "User" {
			return nil, fmt.Errorf(
				"cannot search for studies within type `%s`",
				within.Type,
			)
		}
	}
	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
	}
	from := "study_search_index"
	sql, args := SearchSQL(selects, from, within, query, po)

	psName := preparedName("searchStudyIndex", sql)

	return getManyStudy(db, psName, sql, args...)
}

func UpdateStudy(
	db Queryer,
	row *Study,
) (*Study, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateStudy(id)")
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.AdvancedAt.Status != pgtype.Undefined {
		sets = append(sets, `advanced_at`+"="+args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, studyDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}

	if len(sets) == 0 {
		return GetStudy(db, row.Id.String)
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
		WHERE id = ` + args.Append(row.Id.String) + `
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

	study, err := GetStudy(tx, row.Id.String)
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

func studyDelimeter(r rune) bool {
	return r == '-' || r == '_'
}
