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
	AppledAt    pgtype.Timestamptz `db:"appled_at" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"read"`
	EnrolledAt  pgtype.Timestamptz `db:"enrolled_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Name        mytype.URLSafeName `db:"name" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
	UserLogin   pgtype.Text        `db:"user_login" permit:"read"`
}

func NewStudyService(db Queryer) *StudyService {
	return &StudyService{db}
}

type StudyService struct {
	db Queryer
}

const countStudyByAppledSQL = `
	SELECT COUNT(*)
	FROM study_apple
	WHERE user_id = $1
`

func (s *StudyService) CountByAppled(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("Study.CountByAppled(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyByAppled",
		countStudyByAppledSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countStudyByUserSQL = `
	SELECT COUNT(*)
	FROM study
	WHERE user_id = $1
`

func (s *StudyService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("Study.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyByUser",
		countStudyByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *StudyService) CountBySearch(within *mytype.OID, query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("Study.CountBySearch(query)")
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

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *StudyService) get(name string, sql string, args ...interface{}) (*Study, error) {
	var row Study
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get study")
		return nil, err
	}

	return &row, nil
}

func (s *StudyService) getMany(name string, sql string, args ...interface{}) ([]*Study, error) {
	var rows []*Study

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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
			&row.UserLogin,
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
		user_id,
		user_login
	FROM study_master
	WHERE id = $1
`

func (s *StudyService) Get(id string) (*Study, error) {
	mylog.Log.WithField("id", id).Info("Study.Get(id)")
	return s.get("getStudyById", getStudyByIdSQL, id)
}

func (s *StudyService) GetByAppled(
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("Study.GetByAppled(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `appled_study.actor_id = ` + args.Append(userId)

	selects := []string{
		"advanced_at",
		"appled_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "appled_study"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudiesByAppled", sql)

	var rows []*Study

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
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
			&row.UserLogin,
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

func (s *StudyService) GetByEnrolled(
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("Study.GetByEnrolled(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `enrolled_study.actor_id = ` + args.Append(userId)

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"enrolled_at",
		"id",
		"name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "enrolled_study"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudiesByEnrolled", sql)

	var rows []*Study

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
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
			&row.UserLogin,
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

func (s *StudyService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("Study.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_master.user_id = ` + args.Append(userId)

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "study_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudiesByUserId", sql)

	return s.getMany(psName, sql, args...)
}

const getStudyByNameSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		updated_at,
		user_id,
		user_login
	FROM study_master
	WHERE user_id = $1 AND lower(name) = lower($2)
`

func (s *StudyService) GetByName(userId, name string) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"name":    name,
	}).Info("Study.GetByName(user_id, name)")
	return s.get("getStudyByName", getStudyByNameSQL, userId, name)
}

const getStudyByUserAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.updated_at,
		s.user_id,
		a.login user_login
	FROM study s
	INNER JOIN account a ON a.login = $1
	WHERE s.user_id = a.id AND lower(s.name) = lower($2)  
`

func (s *StudyService) GetByUserAndName(owner, name string) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"owner": owner,
		"name":  name,
	}).Info("Study.GetByUserAndName(owner, name)")
	return s.get("getStudyByUserAndName", getStudyByUserAndNameSQL, owner, name)
}

func (s *StudyService) Create(row *Study) (*Study, error) {
	mylog.Log.Info("Study.Create()")
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
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

	studySvc := NewStudyService(tx)
	study, err := studySvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
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

func (s *StudyService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Study.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteStudy", deleteStudySQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshStudySearchIndexSQL = `
	REFRESH MATERIALIZED VIEW CONCURRENTLY study_search_index
`

func (s *StudyService) RefreshSearchIndex() error {
	mylog.Log.Info("Study.RefreshSearchIndex()")
	_, err := prepareExec(
		s.db,
		"refreshStudySearchIndex",
		refreshStudySearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudyService) Search(within *mytype.OID, query string, po *PageOptions) ([]*Study, error) {
	mylog.Log.WithField("query", query).Info("Study.Search(query)")
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
		"user_login",
	}
	from := "study_search_index"
	sql, args := SearchSQL(selects, from, within, query, po)

	psName := preparedName("searchStudyIndex", sql)

	return s.getMany(psName, sql, args...)
}

func (s *StudyService) Update(row *Study) (*Study, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Study.Update(id)")
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
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

	studySvc := NewStudyService(tx)
	study, err := studySvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
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
