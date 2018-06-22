package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type UserFollow struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	PupilId mytype.OID         `db:"pupil_id" permit:"read"`
	LeaderId   mytype.OID         `db:"leader_id" permit:"read"`
}

func NewUserFollowService(db Queryer) *UserFollowService {
	return &UserFollowService{db}
}

type UserFollowService struct {
	db Queryer
}

const countUserFollowByPupilSQL = `
	SELECT COUNT(*)
	FROM user_follow
	WHERE pupil_id = $1
`

func (s *UserFollowService) CountByPupil(pupilId string) (int32, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserFollow.CountByPupil(pupil_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserFollowByPupil",
		countUserFollowByPupilSQL,
		pupilId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserFollowByLeaderSQL = `
	SELECT COUNT(*)
	FROM user_follow
	WHERE leader_id = $1
`

func (s *UserFollowService) CountByLeader(leaderId string) (int32, error) {
	mylog.Log.WithField("leader_id", leaderId).Info("UserFollow.CountByLeader(leader_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserFollowByLeader",
		countUserFollowByLeaderSQL,
		leaderId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *UserFollowService) get(
	name string,
	sql string,
	args ...interface{},
) (*UserFollow, error) {
	var row UserFollow
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.PupilId,
		&row.LeaderId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_follow")
		return nil, err
	}

	return &row, nil
}

func (s *UserFollowService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*UserFollow, error) {
	var rows []*UserFollow

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserFollow
		dbRows.Scan(
			&row.CreatedAt,
			&row.PupilId,
			&row.LeaderId,
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

const getUserFollowSQL = `
	SELECT
		created_at,
		pupil_id,
		leader_id
	FROM user_follow
	WHERE leader_id = $1 AND pupil_id = $2
`

func (s *UserFollowService) Get(leaderId, pupilId string) (*UserFollow, error) {
	mylog.Log.WithFields(logrus.Fields{
		"leader_id":   leaderId,
		"pupil_id": pupilId,
	}).Info("UserFollow.Get()")
	return s.get("getUserFollow", getUserFollowSQL, leaderId, pupilId)
}

func (s *UserFollowService) GetByPupil(
	pupilId string,
	po *PageOptions,
) ([]*UserFollow, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserFollow.GetByPupil(pupil_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_follow.pupil_id = ` + args.Append(pupilId)

	selects := []string{
		"created_at",
		"pupil_id",
		"leader_id",
	}
	from := "user_follow"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserFollowsByPupilId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserFollowService) GetByLeader(
	leaderId string,
	po *PageOptions,
) ([]*UserFollow, error) {
	mylog.Log.WithField("leader_id", leaderId).Info("UserFollow.GetByLeader(leader_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_follow.leader_id = ` + args.Append(leaderId)

	selects := []string{
		"created_at",
		"pupil_id",
		"leader_id",
	}
	from := "user_follow"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserFollowsByLeaderId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserFollowService) Create(row *UserFollow) (*UserFollow, error) {
	mylog.Log.Info("UserFollow.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.PupilId.Status != pgtype.Undefined {
		columns = append(columns, "pupil_id")
		values = append(values, args.Append(&row.PupilId))
	}
	if row.LeaderId.Status != pgtype.Undefined {
		columns = append(columns, "leader_id")
		values = append(values, args.Append(&row.LeaderId))
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
		INSERT INTO user_follow(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserFollow", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_follow")
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

	userFollowSvc := NewUserFollowService(tx)
	userFollow, err := userFollowSvc.Get(row.LeaderId.String, row.PupilId.String)
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

	return userFollow, nil
}

const deleteUserFollowSQL = `
	DELETE FROM user_follow
	WHERE leader_id = $1 AND pupil_id = $2
`

func (s *UserFollowService) Delete(leaderId, pupilId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"leader_id":   leaderId,
		"pupil_id": pupilId,
	}).Info("UserFollow.Delete(leader_id, pupil_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteUserFollow",
		deleteUserFollowSQL,
		leaderId,
		pupilId,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
