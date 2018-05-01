package data

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"

	"github.com/jackc/pgx"
)

var ErrNotFound = errors.New("not found")

type Queryer interface {
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
	QueryRow(sql string, args ...interface{}) *pgx.Row
	Exec(sql string, arguments ...interface{}) (pgx.CommandTag, error)
}

type transactor interface {
	Begin() (*pgx.Tx, error)
}

type preparer interface {
	Prepare(name, sql string) (*pgx.PreparedStatement, error)
	Deallocate(name string) error
}

func beginTransaction(db Queryer) (*pgx.Tx, error) {
	if transactor, ok := db.(transactor); ok {
		return transactor.Begin()
	}
	return nil, errors.New("queryer is not a transactor")
}

func prepareQuery(db Queryer, name, sql string, args ...interface{}) (*pgx.Rows, error) {
	if preparer, ok := db.(preparer); ok {
		if _, err := preparer.Prepare(name, sql); err != nil {
			return nil, err
		}
		sql = name
	}

	return db.Query(sql, args...)
}

func prepareQueryRow(db Queryer, name, sql string, args ...interface{}) *pgx.Row {
	if preparer, ok := db.(preparer); ok {
		// QueryRow doesn't return an error, the error is encoded in the pgx.Row.
		// Since that is private, Ignore the error from Prepare and run the query
		// without the prepared statement. It should fail with the same error.
		if _, err := preparer.Prepare(name, sql); err == nil {
			sql = name
		}
	}

	return db.QueryRow(sql, args...)
}

func prepareExec(db Queryer, name, sql string, args ...interface{}) (pgx.CommandTag, error) {
	if preparer, ok := db.(preparer); ok {
		if _, err := preparer.Prepare(name, sql); err != nil {
			return pgx.CommandTag(""), err
		}
		sql = name
	}

	return db.Exec(sql, args...)
}

func preparedName(baseName, sql string) string {
	h := fnv.New32a()
	if _, err := io.WriteString(h, sql); err != nil {
		// hash.Hash.Write never returns an error so this can't happen
		panic("failed writing to hash")
	}

	return fmt.Sprintf("%s%d", baseName, h.Sum32())
}
