package data

import (
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

type LessonModel struct {
	Body         pgtype.Text      `db:"body"`
	CreatedAt    time.Time        `db:"created_at"`
	Id           string           `db:"id"`
	LastEditedAt time.Time        `db:"last_edited_at"`
	Number       int              `db:"number"`
	PublishedAt  pgtype.Timestamp `db:"published_at"`
	Title        string           `db:"title"`
	UserId       string           `db:"user_id"`
}

func NewLessonService(db *mydb.DB) *LessonService {
	return &LessonService{db}
}

type LessonService struct {
	*mydb.DB
}

const getByIdSQL = `
	SELECT
		body,
		created_at,
		id,
		last_edited_at,
		number,
		published_at,
		title,
		user_id,
`
