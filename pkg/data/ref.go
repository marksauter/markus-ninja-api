package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Ref struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id        mytype.OID         `db:"id" permit:"read"`
	SourceId  mytype.OID         `db:"source_id" permit:"read"`
	TargetId  mytype.OID         `db:"target_id" permit:"read"`
	UserId    mytype.OID         `db:"user_id" permit:"read"`
}

func NewRefService(db Queryer) *RefService {
	return &RefService{db}
}

type RefService struct {
	db Queryer
}

const countRefByTargetSQL = `
	SELECT COUNT(*)
	FROM ref
	WHERE target_id = $1
`

func (s *RefService) CountByTarget(
	targetId string,
) (int32, error) {
	mylog.Log.WithField("target_id", targetId).Info("Ref.CountByTarget()")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countRefByTarget",
		countRefByTargetSQL,
		targetId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *RefService) get(
	name string,
	sql string,
	args ...interface{},
) (*Ref, error) {
	var row Ref
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.SourceId,
		&row.TargetId,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get ref")
		return nil, err
	}

	return &row, nil
}

func (s *RefService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Ref, error) {
	var rows []*Ref

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Ref
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.SourceId,
			&row.TargetId,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get refs")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getRefSQL = `
	SELECT
		created_at,
		id,
		source_id,
		target_id,
		user_id
	FROM ref
	WHERE id = $1
`

func (s *RefService) Get(id string) (*Ref, error) {
	mylog.Log.WithField("id", id).Info("Ref.Get(id)")
	return s.get("getRef", getRefSQL, id)
}

func (s *RefService) GetBySource(
	sourceId string,
	po *PageOptions,
) ([]*Ref, error) {
	mylog.Log.WithField("source_id", sourceId).Info("Ref.GetBySource(source_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `ref.source_id = ` + args.Append(sourceId)

	selects := []string{
		"created_at",
		"id",
		"source_id",
		"target_id",
		"user_id",
	}
	from := "ref"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getRefsBySource", sql)

	return s.getMany(psName, sql, args...)
}

func (s *RefService) GetByTarget(
	targetId string,
	po *PageOptions,
) ([]*Ref, error) {
	mylog.Log.WithField("target_id", targetId).Info("Ref.GetByTarget(target_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `ref.target_id = ` + args.Append(targetId)

	selects := []string{
		"created_at",
		"id",
		"source_id",
		"target_id",
		"user_id",
	}
	from := "ref"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getRefsByTarget", sql)

	return s.getMany(psName, sql, args...)
}

func (s *RefService) Create(row *Ref) (*Ref, error) {
	mylog.Log.Info("Ref.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	id, _ := mytype.NewOID("Ref")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.SourceId.Status != pgtype.Undefined {
		columns = append(columns, "source_id")
		values = append(values, args.Append(&row.SourceId))
	}
	if row.TargetId.Status != pgtype.Undefined {
		columns = append(columns, "target_id")
		values = append(values, args.Append(&row.TargetId))
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

	var source string
	switch row.SourceId.Type {
	case "Lesson":
		source = "lesson"
	case "LessonComment":
		source = "lesson_comment"
	default:
		return nil, fmt.Errorf("invalid type '%s' for ref source id", row.SourceId.Type)
	}
	var target string
	switch row.TargetId.Type {
	case "Lesson":
		target = "lesson"
	case "User":
		target = "user"
	default:
		return nil, fmt.Errorf("invalid type '%s' for ref target id", row.TargetId.Type)
	}

	table := strings.Join([]string{source, target, "ref"}, "_")
	sql := `
		INSERT INTO ` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createRef", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create ref")
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

	refSvc := NewRefService(tx)
	ref, err := refSvc.Get(row.Id.String)
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

	return ref, nil
}

func (s *RefService) BatchCreate(src *Ref, targetIds []*mytype.OID) error {
	mylog.Log.Info("Ref.BatchCreate()")

	n := len(targetIds)
	lessonRefs := make([][]interface{}, 0, n)
	userRefs := make([][]interface{}, 0, n)
	for _, targetId := range targetIds {
		id, _ := mytype.NewOID("Ref")
		src.Id.Set(id)
		switch targetId.Type {
		case "Lesson":
			lessonRefs = append(lessonRefs, []interface{}{
				src.Id.String,
				targetId.String,
				src.SourceId.String,
				src.UserId.String,
			})
		case "User":
			userRefs = append(userRefs, []interface{}{
				src.Id.String,
				targetId.String,
				src.SourceId.String,
				src.UserId.String,
			})
		default:
			return fmt.Errorf("invalid type '%s' for ref target id", targetId.Type)
		}
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	var identPrefix string
	switch src.SourceId.Type {
	case "Lesson":
		identPrefix = "lesson_"
	case "LessonComment":
		identPrefix = "lesson_comment_"
	default:
		return fmt.Errorf("invalid type '%s' for ref source id", src.SourceId.Type)
	}
	lessonRefCopyCount, err := tx.CopyFrom(
		pgx.Identifier{identPrefix + "lesson_ref"},
		[]string{"ref_id", "target_id", "source_id", "user_id"},
		pgx.CopyFromRows(lessonRefs),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			default:
				return err
			case UniqueViolation:
				mylog.Log.Warn("refs already created")
				return nil
			}
		}
		return err
	}

	userRefCopyCount, err := tx.CopyFrom(
		pgx.Identifier{identPrefix + "user_ref"},
		[]string{"ref_id", "target_id", "source_id", "user_id"},
		pgx.CopyFromRows(userRefs),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			default:
				return err
			case UniqueViolation:
				mylog.Log.Warn("refs already created")
				return nil
			}
		}
		return err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	mylog.Log.WithField("n", lessonRefCopyCount+userRefCopyCount).Info("created refs")

	return nil
}

const deleteUserRefSQL = `
	DELETE FROM ref
	WHERE id = $1
`

func (s *RefService) Delete(id *mytype.OID) error {
	mylog.Log.WithField("id", id).Info("Ref.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteRef",
		deleteUserRefSQL,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *RefService) ParseBodyForRefs(
	userId,
	studyId,
	sourceId *mytype.OID,
	body *mytype.Markdown,
) error {
	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	lessonSvc := NewLessonService(tx)
	refSvc := NewRefService(tx)

	lessonNumberRefs, err := body.NumberRefs()
	if err != nil {
		return err
	}
	userRefs := body.AtRefs()
	// TODO: add support for cross study references
	// crossStudyRefs, err := body.CrossStudyRefs()
	// if err != nil {
	//   return err
	// }
	targetIds := make([]*mytype.OID, 0, len(lessonNumberRefs)+len(userRefs))
	if len(lessonNumberRefs) > 0 {
		lessons, err := lessonSvc.BatchGetByNumber(
			userId.String,
			studyId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			targetIds = append(targetIds, &l.Id)
		}
	}
	if len(userRefs) > 0 {
		userSvc := NewUserService(tx)
		users, err := userSvc.BatchGetByLogin(
			userRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range users {
			targetIds = append(targetIds, &l.Id)
		}
	}

	ref := &Ref{}
	ref.SourceId.Set(sourceId)
	ref.UserId.Set(userId)
	err = refSvc.BatchCreate(ref, targetIds)
	if err != nil {
		return err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	return nil
}

func (s *RefService) ParseUpdatedBodyForRefs(
	userId,
	studyId,
	sourceId *mytype.OID,
	body *mytype.Markdown,
) error {
	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	lessonSvc := NewLessonService(tx)
	refSvc := NewRefService(tx)

	newRefs := make(map[string]struct{})
	oldRefs := make(map[string]struct{})
	refs, err := refSvc.GetBySource(sourceId.String, nil)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		oldRefs[ref.TargetId.String] = struct{}{}
	}

	lessonNumberRefs, err := body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberRefs) > 0 {
		lessons, err := lessonSvc.BatchGetByNumber(
			userId.String,
			studyId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			newRefs[l.Id.String] = struct{}{}
			if _, prs := oldRefs[l.Id.String]; !prs {
				ref := &Ref{}
				ref.TargetId.Set(l.Id)
				ref.SourceId.Set(sourceId)
				ref.UserId.Set(userId)
				_, err = refSvc.Create(ref)
				if err != nil {
					return err
				}
			}
		}
	}
	userRefs := body.AtRefs()
	// TODO: add support for cross study references
	// crossStudyRefs, err := body.CrossStudyRefs()
	// if err != nil {
	//   return err
	// }
	if len(userRefs) > 0 {
		userSvc := NewUserService(tx)
		users, err := userSvc.BatchGetByLogin(
			userRefs,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			newRefs[u.Id.String] = struct{}{}
			if _, prs := oldRefs[u.Id.String]; !prs {
				ref := &Ref{}
				ref.TargetId.Set(u.Id)
				ref.SourceId.Set(sourceId)
				ref.UserId.Set(userId)
				_, err = refSvc.Create(ref)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, ref := range refs {
		if _, prs := newRefs[ref.TargetId.String]; !prs {
			err := refSvc.Delete(&ref.Id)
			if err != nil {
				return err
			}
		}
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	return nil
}
