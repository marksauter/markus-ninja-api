package data

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type LessonCommentDraftBackup struct {
	CreatedAt       pgtype.Timestamptz `db:"created_at" permit:"read"`
	Draft           pgtype.Text        `db:"draft" permit:"read"`
	ID              pgtype.Int4        `db:"id" permit:"read"`
	LessonCommentID mytype.OID         `db:"lesson_comment_id" permit:"read"`
	UpdatedAt       pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func getLessonCommentDraftBackup(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*LessonCommentDraftBackup, error) {
	var row LessonCommentDraftBackup
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Draft,
		&row.ID,
		&row.LessonCommentID,
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

func getManyLessonCommentDraftBackup(
	db Queryer,
	name string,
	sql string,
	rows *[]*LessonCommentDraftBackup,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row LessonCommentDraftBackup
		dbRows.Scan(
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LessonCommentID,
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

const getLessonCommentDraftBackupSQL = `
	SELECT
		created_at,
		draft,
		id,
		lesson_comment_id,
		updated_at
	FROM lesson_comment_draft_backup
	WHERE lesson_comment_id = $1 AND id = $2
`

func GetLessonCommentDraftBackup(
	db Queryer,
	lessonCommentID string,
	id int32,
) (*LessonCommentDraftBackup, error) {
	lessonComment, err := getLessonCommentDraftBackup(
		db,
		"getLessonCommentDraftBackup",
		getLessonCommentDraftBackupSQL,
		lessonCommentID,
		id,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"lesson_comment_id": lessonCommentID,
			"id":                id,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"lesson_comment_id": lessonCommentID,
			"id":                id,
		}).Info(util.Trace("lesson comment draft backup found"))
	}
	return lessonComment, err
}

const getLessonCommentDraftBackupByLessonCommentSQL = `
	SELECT
		created_at,
		draft,
		id,
		lesson_comment_id,
		updated_at
	FROM lesson_comment_draft_backup
	WHERE lesson_comment_id = $1
	ORDER BY updated_at DESC
`

func GetLessonCommentDraftBackupByLessonComment(
	db Queryer,
	lessonCommentID string,
) ([]*LessonCommentDraftBackup, error) {
	var rows []*LessonCommentDraftBackup
	err := getManyLessonCommentDraftBackup(
		db,
		"getLessonCommentDraftBackupByLessonComment",
		getLessonCommentDraftBackupByLessonCommentSQL,
		&rows,
		lessonCommentID,
	)
	if err != nil {
		mylog.Log.WithField("lesson_comment_id", lessonCommentID).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("lesson_comment_id", lessonCommentID).Info(util.Trace("lesson comment draft backups found"))
	}
	return rows, err
}

const restoreLessonCommentDraftFromBackupSQl = `
	UPDATE lesson_comment
	SET draft = backup.draft
	FROM lesson_comment_draft_backup AS backup
	WHERE lesson_comment.id = $1 AND backup.lesson_comment_id = lesson_comment.id AND backup.id = $2
`

func RestoreLessonCommentDraftFromBackup(
	db Queryer,
	lessonCommentID string,
	backupID int32,
) error {
	commandTag, err := prepareExec(
		db,
		"restoreLessonCommentDraftFromBackup",
		restoreLessonCommentDraftFromBackupSQl,
		lessonCommentID,
		backupID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"lesson_comment_id": lessonCommentID,
			"backup_id":         backupID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"lesson_comment_id": lessonCommentID,
		"backup_id":         backupID,
	}).Info(util.Trace("lesson comment draft restored"))
	return nil
}
