package data

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"strings"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

var ErrNotFound = errors.New("not found")

type Queryer interface {
	CopyFrom(pgx.Identifier, []string, pgx.CopyFromSource) (int, error)
	Exec(sql string, arguments ...interface{}) (pgx.CommandTag, error)
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
	QueryRow(sql string, args ...interface{}) *pgx.Row
}

type transactor interface {
	Begin() (*pgx.Tx, error)
	Commit() error
}

type preparer interface {
	Prepare(name, sql string) (*pgx.PreparedStatement, error)
	Deallocate(name string) error
}

func beginTransaction(db Queryer) (Queryer, error) {
	if transactor, ok := db.(transactor); ok {
		return transactor.Begin()
	}
	return db, nil
}

func commitTransaction(db Queryer) error {
	if transactor, ok := db.(transactor); ok {
		return transactor.Commit()
	}
	return nil
}

func rollbackTransaction(db Queryer) error {
	if transactor, ok := db.(transactor); ok {
		return transactor.Commit()
	}
	return nil
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

type OrderDirection bool

const (
	ASC  OrderDirection = false
	DESC OrderDirection = true
)

func ParseOrderDirection(s string) (OrderDirection, error) {
	switch strings.ToLower(s) {
	case "asc":
		return ASC, nil
	case "desc":
		return DESC, nil
	default:
		var o OrderDirection
		return o, fmt.Errorf("invalid OrderDirection: %q", s)
	}
}

func (od OrderDirection) String() string {
	switch od {
	case ASC:
		return "ASC"
	case DESC:
		return "DESC"
	default:
		return "unknown"
	}
}

type Order interface {
	Direction() OrderDirection
	Field() string
}

var ErrEmptyPageOptions = errors.New("`po` (*PageOptions) must not be nil")

type PageOptions struct {
	After  *Cursor
	Before *Cursor
	First  int32
	Last   int32
	Order  Order
}

func NewPageOptions(after, before *string, first, last *int32, o Order) (*PageOptions, error) {
	pageOptions := &PageOptions{
		Order: o,
	}
	if first == nil && last == nil {
		return nil, fmt.Errorf("You must provide a `first` or `last` value to properly paginate.")
	} else if first != nil {
		if last != nil {
			return nil, fmt.Errorf("Passing both `first` and `last` values to paginate the connection is not supported.")
		}
		pageOptions.First = *first
	} else {
		pageOptions.Last = *last
	}
	if after != nil {
		a, err := NewCursor(after)
		if err != nil {
			return nil, err
		}
		pageOptions.After = a
	}
	if before != nil {
		b, err := NewCursor(before)
		if err != nil {
			return nil, err
		}
		pageOptions.Before = b
	}
	return pageOptions, nil
}

// If the query is asking for the last elements in a list, then we need two
// queries to get the items more efficiently and in the right order.
// First, we query the reverse direction of that requested, so that only
// the items needed are returned.
func (p *PageOptions) QueryDirection() string {
	direction := p.Order.Direction()
	if p.Last != 0 {
		direction = !p.Order.Direction()
	}
	return direction.String()
}

func (p *PageOptions) Limit() int32 {
	// Assuming one of these is 0, so the sum will be the non-zero field + 1
	limit := p.First + p.Last + 1
	if (p.After != nil && p.First > 0) ||
		(p.Before != nil && p.Last > 0) {
		limit = limit + int32(1)
	}
	return limit
}

func (p *PageOptions) joins(from string, args *pgx.QueryArgs) []string {
	var joins []string
	if p.After != nil {
		joins = append(joins, fmt.Sprintf(
			"INNER JOIN %s %s2 ON %s2.id = "+args.Append(p.After.Value()),
			from,
		))
	}
	if p.Before != nil {
		joins = append(joins, fmt.Sprintf(
			"INNER JOIN %s %s3 ON %s3.id = "+args.Append(p.Before.Value()),
			from,
		))
	}
	return joins
}

func (p *PageOptions) whereAnds(from string) []string {
	var whereAnds []string
	field := p.Order.Field()
	if p.After != nil {
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %s.%s >= %s2.%s",
			from,
			field,
			from,
			field,
		))
	}
	if p.Before != nil {
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %s.%s >= %s3.%s",
			from,
			field,
			from,
			field,
		))
	}
	return whereAnds
}

func SQL(selects []string, from, where string, args *pgx.QueryArgs, po *PageOptions) string {
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, args)
		whereAnds = po.whereAnds(from)
		limit = "LIMIT " + args.Append(po.Limit())
		orderBy = "ORDER BY " +
			from + "." + po.Order.Field() + " " + po.QueryDirection()
	}

	sql := `
		SELECT 
		` + strings.Join(selects, ",") + `
		FROM ` + from + `
		` + strings.Join(joins, " ") + `
		WHERE ` + where + `
		` + strings.Join(whereAnds, " ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql)
}

func SearchSQL(
	selects []string,
	from string,
	within *mytype.OID,
	query string,
	po *PageOptions,
) (string, pgx.QueryArgs) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, &args)
		whereAnds = po.whereAnds(from)
		limit = "LIMIT " + args.Append(po.Limit())

		field := po.Order.Field()
		orderBy := ""
		if field != "best_match" {
			orderBy = from + "." + field
		} else {
			orderBy = "ts_rank(document, query)"
		}

		orderBy = "ORDER BY " + orderBy + " " + po.QueryDirection()
	}
	if within != nil {
		andIn := fmt.Sprintf(
			"AND %s.%s = %s",
			from,
			within.DBVarName(),
			args.Append(within),
		)
		whereAnds = append(whereAnds, andIn)
	}

	tsquery := ToTsQuery(query)
	sql := `
		SELECT 
		` + strings.Join(selects, ",") + `
		FROM ` + from + `, to_tsquery('simple',` + args.Append(tsquery) + `) query
		` + strings.Join(joins, " ") + `
		WHERE document @@ query
		` + strings.Join(whereAnds, " ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql), args
}

// Then, we can reorder the items to the originally requested direction.
func ReorderQuery(po *PageOptions, query string) string {
	if po != nil && po.Last != 0 {
		return fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			query,
			po.Order.Field(),
			po.Order.Direction(),
		)
	}
	return query
}
