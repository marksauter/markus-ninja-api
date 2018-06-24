package data

import (
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Reason struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	Description pgtype.Text        `db:"description"`
	Name        mytype.ReasonName  `db:"name"`
}

type ReasonService struct {
	db Queryer
}

func NewReasonService(db Queryer) *ReasonService {
	return &ReasonService{db}
}
