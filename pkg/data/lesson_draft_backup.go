package data

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type LessonDraftBackup struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Draft     pgtype.Text        `db:"draft" permit:"read/update"`
	ID        pgtype.Int4        `db:"id" permit:"read"`
	LessonID  mytype.OID         `db:"lesson_id" permit:"create/read"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func getLessonDraftBackup(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*LessonDraftBackup, error) {
	var row LessonDraftBackup
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Draft,
		&row.ID,
		&row.LessonID,
		&row.UpdatedAt,
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

func getManyLessonDraftBackup(
	db Queryer,
	name string,
	sql string,
	rows *[]*LessonDraftBackup,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row LessonDraftBackup
		dbRows.Scan(
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LessonID,
			&row.UpdatedAt,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getLessonDraftBackupSQL = `
	SELECT
		created_at,
		draft,
		id,
		lesson_id,
		updated_at,
	FROM lesson_draft_backup
	WHERE lesson_id = $1 AND id = $2
`

func GetLessonDraftBackup(
	db Queryer,
	lessonID,
	id string,
) (*LessonDraftBackup, error) {
	lesson, err := getLessonDraftBackup(db, "getLessonDraftBackup", getLessonDraftBackupSQL, lessonID, id)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"lesson_id": lessonID,
			"id":        id,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"lesson_id": lessonID,
			"id":        id,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const getLessonDraftBackupByLessonSQL = `
	SELECT
		created_at,
		draft,
		id,
		lesson_id,
		updated_at,
	FROM lesson_draft_backup
	WHERE lesson_id = $1
`

func GetLessonDraftBackupByLesson(
	db Queryer,
	lessonID string,
) ([]*LessonDraftBackup, error) {
	var rows []*LessonDraftBackup
	err := getManyLessonDraftBackup(
		db,
		"getLessonDraftBackupByLesson",
		getLessonDraftBackupByLessonSQL,
		&rows,
		lessonID,
	)
	if err != nil {
		mylog.Log.WithField("lesson_id", lessonID).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("lesson_id", lessonID).Info(util.Trace("lesson found"))
	}
	return rows, err
}

const restoreLessonDraftFromBackupSQl = `
	UPDATE lesson
	SET draft = backup.draft
	FROM lesson_draft_backup AS backup
	WHERE lesson.id = $1 AND backup.lesson_id = lesson.id AND backup.id = $2
`

func RestoreLessonDraftFromBackup(
	db Queryer,
	lessonID,
	backupID string,
) error {
	commandTag, err := prepareExec(
		db,
		"restoreLessonDraftFromBackup",
		restoreLessonDraftFromBackupSQl,
		lessonID,
		backupID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"lesson_id": lessonID,
			"backup_id": backupID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"lesson_id": lessonID,
		"backup_id": backupID,
	}).Info(util.Trace("lesson draft restored"))
	return nil
}
