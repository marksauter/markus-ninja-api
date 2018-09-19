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
	ID          mytype.OID         `db:"id" permit:"read"`
	Name        mytype.URLSafeName `db:"name" permit:"create/read"`
	Private     pgtype.Bool        `db:"private" permit:"create/read/update"`
	TopicedAt   pgtype.Timestamptz `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID      mytype.OID         `db:"user_id" permit:"create/read"`
}

const countStudyByAppleeSQL = `
	SELECT COUNT(*)
	FROM study_appled
	WHERE user_id = $1
`

func CountStudyByApplee(
	db Queryer,
	appleeID string,
) (n int32, err error) {
	mylog.Log.WithField("applee_id", appleeID).Info("CountStudyByApplee(applee_id)")
	err = prepareQueryRow(
		db,
		"countStudyByApplee",
		countStudyByAppleeSQL,
		appleeID,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countStudyByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM enrolled
	WHERE user_id = $1 AND status = 'ENROLLED'
`

func CountStudyByEnrollee(
	db Queryer,
	enrolleeID string,
) (n int32, err error) {
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("CountStudyByEnrollee(enrollee_id)")
	err = prepareQueryRow(
		db,
		"countStudyByEnrollee",
		countStudyByEnrolleeSQL,
		enrolleeID,
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
	topicID string,
) (n int32, err error) {
	mylog.Log.WithField(
		"topic_id", topicID,
	).Info("CountStudyByTopic(topic_id)")
	err = prepareQueryRow(
		db,
		"countStudyByTopic",
		countStudyByTopicSQL,
		topicID,
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
	userID string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userID).Info("CountStudyByUser(user_id)")
	err = prepareQueryRow(
		db,
		"countStudyByUser",
		countStudyByUserSQL,
		userID,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountStudyBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountStudyBySearch(query)")
	var n int32
	var args pgx.QueryArgs
	from := "study_search_index"
	in := within
	if in != nil {
		if in.Type != "User" && in.Type != "Topic" {
			return n, fmt.Errorf(
				"cannot search for studies within type `%s`",
				in.Type,
			)
		}
		if in.Type == "Topic" {
			topic, err := GetTopic(db, in.String)
			if err != nil {
				return n, err
			}
			from = ToSubQuery(
				SearchSQL(
					[]string{"*"},
					from,
					nil,
					ToTsQuery(topic.Name.String),
					"topics",
					nil,
					&args,
				),
			)
			// set `in` to nil so that it doesn't affect next call to SearchSQL
			// TODO: fix this ugliness please
			in = nil
		}
	}

	sql := CountSearchSQL(from, in, ToPrefixTsQuery(query), "document", &args)

	psName := preparedName("countStudyBySearch", sql)

	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountStudyByTopicSearch(
	db Queryer,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountStudyBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	if query == "" {
		query = "*"
	}
	sql := `
		SELECT COUNT(*)
		FROM study_search_index
		WHERE topics @@ to_tsquery('simple',` + query + `)
	`
	psName := preparedName("countStudyBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

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

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
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
	userID string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userID).Info("GetStudyByApplee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`applee_id = ` + args.Append(userID)}

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

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByEnrollee(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userID).Info("GetStudyByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userID)}

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

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("topic_id", topicID).Info("GetStudyByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topic_id = ` + args.Append(topicID)}

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

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetStudyByUser(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userID).Info("GetStudyByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userID)}

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
	from := "study"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByUserID", sql)

	return getManyStudy(db, psName, sql, args...)
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
	within *mytype.OID,
	query string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("query", query).Info("SearchStudy(query)")
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

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	in := within
	if in != nil {
		if in.Type != "User" && in.Type != "Topic" {
			return nil, fmt.Errorf(
				"cannot search for studies within type `%s`",
				in.Type,
			)
		}
		if in.Type == "Topic" {
			topic, err := GetTopic(tx, in.String)
			if err != nil {
				return nil, err
			}
			from = ToSubQuery(
				SearchSQL(
					[]string{"*"},
					from,
					nil,
					ToTsQuery(topic.Name.String),
					"topics",
					po,
					&args,
				),
			)
			// set `in` to nil so that it doesn't affect next call to SearchSQL
			// TODO: fix this ugliness please
			in = nil
		}
	}

	sql := SearchSQL(selects, from, in, ToPrefixTsQuery(query), "document", po, &args)

	psName := preparedName("searchStudyIndex", sql)

	studies, err := getManyStudy(tx, psName, sql, args...)
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

	return studies, nil
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

func studyDelimeter(r rune) bool {
	return r == '-' || r == '_'
}
