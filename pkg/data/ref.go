package data

import (
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
	WHERE target_id = $2
`

func (s *RefService) CountByTarget(
	userId,
	referentId string,
) (int32, error) {
	mylog.Log.WithField("target_id", referentId).Info("Ref.CountByTarget()")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countRefByTarget",
		countRefByTargetSQL,
		userId,
		referentId,
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
		target_id,
		source_id,
		user_id
	FROM ref
`

func (s *RefService) Get(id string) (*Ref, error) {
	mylog.Log.WithField("id", id).Info("Ref.Get(id)")
	return s.get("getRef", getRefSQL, id)
}

func (s *RefService) GetBySource(
	userId,
	referrerId string,
	po *PageOptions,
) ([]*Ref, error) {
	mylog.Log.WithField("source_id", referrerId).Info("Ref.GetBySource(source_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `ref.study_id = ` + args.Append(userId) + `
		AND ref.source_id = ` + args.Append(referrerId)

	selects := []string{
		"created_at",
		"id",
		"source_id",
		"target_id",
		"study_id",
	}
	from := "ref"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getRefsBySource", sql)

	return s.getMany(psName, sql, args...)
}

func (s *RefService) GetByStudy(
	userId string,
	po *PageOptions,
) ([]*Ref, error) {
	mylog.Log.WithField("study_id", userId).Info("Ref.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `ref.study_id = ` + args.Append(userId)

	selects := []string{
		"created_at",
		"id",
		"source_id",
		"target_id",
		"study_id",
	}
	from := "ref"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getRefsByStudy", sql)

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
		columns = append(columns, "study_id")
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
		INSERT INTO ref(` + strings.Join(columns, ",") + `)
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

func (s *RefService) BatchCreate(row *Ref, referentIds []string) error {
	mylog.Log.Info("Ref.BatchCreate()")

	n := len(referentIds)
	refs := make([][]interface{}, n)
	for i, referentId := range referentIds {
		id, _ := mytype.NewOID("Ref")
		row.Id.Set(id)
		refs[i] = []interface{}{
			row.Id.String,
			referentId,
			row.SourceId.String,
			row.UserId.String,
		}
	}

	copyCount, err := s.db.CopyFrom(
		pgx.Identifier{"ref"},
		[]string{"id", "target_id", "source_id", "study_id"},
		pgx.CopyFromRows(refs),
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

	mylog.Log.WithField("n", copyCount).Info("created refs")

	return nil
}

const deleteRefSQL = `
	DELETE FROM ref
	WHERE id = $1
`

func (s *RefService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Ref.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteRef",
		deleteRefSQL,
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

func (s *RefService) ParseStudyBody(
	userId,
	userId,
	referrerId *mytype.OID,
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
	if err != nil {
		return err
	}
	referentIds := make([]string, 0, len(lessonNumberRefs)+len(userRefs))
	if len(lessonNumberRefs) > 0 {
		lessons, err := lessonSvc.BatchGetByNumber(
			userId.String,
			userId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			referentIds = append(referentIds, l.Id.String)
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
			referentIds = append(referentIds, l.Id.String)
		}
	}

	ref := &Ref{}
	ref.SourceId.Set(referrerId)
	ref.UserId.Set(userId)
	err = refSvc.BatchCreate(ref, referentIds)
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

func (s *RefService) ParseUpdatedStudyBody(
	userId,
	userId,
	referrerId *mytype.OID,
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
	refs, err := refSvc.GetBySource(userId.String, referrerId.String, nil)
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
			userId.String,
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
				ref.SourceId.Set(referrerId)
				ref.UserId.Set(userId)
				_, err = refSvc.Create(ref)
				if err != nil {
					return err
				}
			}
		}
	}
	userRefs := body.AtRefs()
	if err != nil {
		return err
	}
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
				ref.SourceId.Set(referrerId)
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
			err := refSvc.Delete(ref.Id.String)
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
