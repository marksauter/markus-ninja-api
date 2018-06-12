package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Lesson struct {
	Body        pgtype.Text        `db:"body" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Number      pgtype.Int4        `db:"number" permit:"read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read"`
	StudyId     mytype.OID         `db:"study_id" permit:"read"`
	StudyName   pgtype.Text        `db:"study_name"`
	Title       pgtype.Text        `db:"title" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
	UserLogin   pgtype.Text        `db:"user_login"`
}

func NewLessonService(db Queryer) *LessonService {
	return &LessonService{db}
}

type LessonService struct {
	db Queryer
}

func (s *LessonService) CountBySearch(within *mytype.OID, query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("Lesson.CountBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM lesson_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			// Only users and studies 'contain' lessons, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND lesson_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countLessonBySearch", sql)

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)
	return
}

const countLessonByStudySQL = `
	SELECT COUNT(*)
	FROM lesson
	WHERE user_id = $1 AND study_id = $2
`

func (s *LessonService) CountByStudy(userId, studyId string) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("Lesson.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonByStudy",
		countLessonByStudySQL,
		userId,
		studyId,
	).Scan(&n)
	return n, err
}

const countLessonByUserSQL = `
	SELECT COUNT(*)
	FROM lesson
	WHERE user_id = $1
`

func (s *LessonService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("Lesson.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonByUser",
		countLessonByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *LessonService) get(name string, sql string, args ...interface{}) (*Lesson, error) {
	var row Lesson
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.StudyName,
		&row.Title,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson")
		return nil, err
	}

	return &row, nil
}

func (s *LessonService) getMany(name string, sql string, args ...interface{}) ([]*Lesson, error) {
	var rows []*Lesson

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.Id,
			&row.Number,
			&row.PublishedAt,
			&row.StudyId,
			&row.StudyName,
			&row.Title,
			&row.UpdatedAt,
			&row.UserId,
			&row.UserLogin,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get lessons")
		return nil, err
	}

	return rows, nil
}

const getLessonByIdSQL = `
	SELECT
		body,
		created_at,
		id,
		number,
		published_at,
		study_id,
		study_name,
		title,
		updated_at,
		user_id,
		user_login
	FROM lesson_master
	WHERE id = $1
`

func (s *LessonService) Get(id string) (*Lesson, error) {
	mylog.Log.WithField("id", id).Info("Lesson.Get(id)")
	return s.get("getLessonById", getLessonByIdSQL, id)
}

const numConnArgs = 3

func (s *LessonService) GetByUser(userId string, po *PageOptions) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userId).Info("Lesson.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `lesson_master.user_id = ` + args.Append(userId)

	selects := []string{
		"body",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"study_name",
		"title",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonService) GetByStudy(userId, studyId string, po *PageOptions) ([]*Lesson, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("Lesson.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `
		lesson_master.user_id = ` + args.Append(userId) + ` AND
		lesson_master.study_id = ` + args.Append(studyId)

	selects := []string{
		"body",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"study_name",
		"title",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonsByStudy", sql)

	return s.getMany(psName, sql, args...)
}

const getLessonByNumberSQL = `
	SELECT
		body,
		created_at,
		id,
		number,
		published_at,
		study_id,
		study_name,
		title,
		updated_at,
		user_id,
		user_login
	FROM lesson_master
	WHERE user_id = $1 AND study_id = $2 AND number = $3
`

func (s *LessonService) GetByNumber(userId, studyId string, number int32) (*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"number":   number,
	}).Info("Lesson.GetByNumber(studyId, number)")
	return s.get(
		"getLessonByNumber",
		getLessonByNumberSQL,
		userId,
		studyId,
		number,
	)
}

func (s *LessonService) Create(row *Lesson) (*Lesson, error) {
	mylog.Log.Info("Lesson.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 8))

	var columns, values []string

	id, _ := mytype.NewOID("Lesson")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.Number.Status != pgtype.Undefined {
		columns = append(columns, "number")
		values = append(values, args.Append(&row.Number))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.Title.Status != pgtype.Undefined {
		columns = append(columns, "title")
		values = append(values, args.Append(&row.Title))
		titleTokens := &pgtype.Text{}
		titleTokens.Set(strings.Join(util.Split(row.Title.String, lessonDelimeter), " "))
		columns = append(columns, "title_tokens")
		values = append(values, args.Append(titleTokens))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		INSERT INTO lesson(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLesson", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson")
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

	lessonSvc := NewLessonService(tx)
	lesson, err := lessonSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	study := &Study{}
	study.Id.Set(row.StudyId)
	err = study.AdvancedAt.Set(time.Now())
	if err != nil {
		return nil, err
	}
	studySvc := NewStudyService(s.db)
	_, err = studySvc.Update(study)
	if err != nil {
		return nil, err
	}

	return lesson, nil
}

const deleteLessonSQl = `
	DELETE FROM lesson
	WHERE id = $1
`

func (s *LessonService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Lesson.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteLesson", deleteLessonSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshLessonSearchIndexSQL = `
	REFRESH MATERIALIZED VIEW CONCURRENTLY lesson_search_index
`

func (s *LessonService) RefreshSearchIndex() error {
	mylog.Log.Info("Lesson.RefreshSearchIndex()")
	_, err := prepareExec(
		s.db,
		"refreshLessonSearchIndex",
		refreshLessonSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *LessonService) Search(within *mytype.OID, query string, po *PageOptions) ([]*Lesson, error) {
	mylog.Log.WithField("query", query).Info("Lesson.Search(query)")
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			return nil, fmt.Errorf(
				"cannot search for lessons within type `%s`",
				within.Type,
			)
		}
	}
	selects := []string{
		"body",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"study_name",
		"title",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_search_index"
	sql, args := SearchSQL(selects, from, within, query, po)

	psName := preparedName("searchLessonIndex", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonService) Update(row *Lesson) (*Lesson, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Lesson.Update(id)")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.Number.Status != pgtype.Undefined {
		sets = append(sets, `number`+"="+args.Append(&row.Number))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		sets = append(sets, `study_id`+"="+args.Append(&row.StudyId))
	}
	if row.Title.Status != pgtype.Undefined {
		sets = append(sets, `title`+"="+args.Append(&row.Title))
		titleTokens := &pgtype.Text{}
		titleTokens.Set(strings.Join(util.Split(row.Title.String, lessonDelimeter), " "))
		sets = append(sets, `title_tokens`+"="+args.Append(titleTokens))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		UPDATE lesson
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateLesson", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson")
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

	lessonSvc := NewLessonService(tx)
	lesson, err := lessonSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return lesson, nil
}

func lessonDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}
