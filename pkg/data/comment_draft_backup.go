package data

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type CommentDraftBackup struct {
	CommentID mytype.OID         `db:"comment_id" permit:"read"`
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Draft     pgtype.Text        `db:"draft" permit:"read"`
	ID        pgtype.Int4        `db:"id" permit:"read"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func getCommentDraftBackup(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*CommentDraftBackup, error) {
	var row CommentDraftBackup
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CommentID,
		&row.CreatedAt,
		&row.Draft,
		&row.ID,
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

func getManyCommentDraftBackup(
	db Queryer,
	name string,
	sql string,
	rows *[]*CommentDraftBackup,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row CommentDraftBackup
		dbRows.Scan(
			&row.CommentID,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
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

const getCommentDraftBackupSQL = `
	SELECT
		comment_id,
		created_at,
		draft,
		id,
		updated_at
	FROM comment_draft_backup
	WHERE comment_id = $1 AND id = $2
`

func GetCommentDraftBackup(
	db Queryer,
	commentID string,
	id int32,
) (*CommentDraftBackup, error) {
	comment, err := getCommentDraftBackup(
		db,
		"getCommentDraftBackup",
		getCommentDraftBackupSQL,
		commentID,
		id,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"comment_id": commentID,
			"id":         id,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"comment_id": commentID,
			"id":         id,
		}).Info(util.Trace("comment draft backup found"))
	}
	return comment, err
}

const getCommentDraftBackupByCommentSQL = `
	SELECT
		comment_id,
		created_at,
		draft,
		id,
		updated_at
	FROM comment_draft_backup
	WHERE comment_id = $1
	ORDER BY updated_at DESC
`

func GetCommentDraftBackupByComment(
	db Queryer,
	commentID string,
) ([]*CommentDraftBackup, error) {
	var rows []*CommentDraftBackup
	err := getManyCommentDraftBackup(
		db,
		"getCommentDraftBackupByComment",
		getCommentDraftBackupByCommentSQL,
		&rows,
		commentID,
	)
	if err != nil {
		mylog.Log.WithField("comment_id", commentID).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("comment_id", commentID).Info(util.Trace("comment draft backups found"))
	}
	return rows, err
}

const restoreCommentDraftFromBackupSQl = `
	UPDATE comment
	SET draft = backup.draft
	FROM comment_draft_backup AS backup
	WHERE comment.id = $1 AND backup.comment_id = comment.id AND backup.id = $2
`

func RestoreCommentDraftFromBackup(
	db Queryer,
	commentID string,
	backupID int32,
) error {
	commandTag, err := prepareExec(
		db,
		"restoreCommentDraftFromBackup",
		restoreCommentDraftFromBackupSQl,
		commentID,
		backupID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"comment_id": commentID,
			"backup_id":  backupID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"comment_id": commentID,
		"backup_id":  backupID,
	}).Info(util.Trace("comment draft restored"))
	return nil
}
