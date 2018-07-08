package data

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"strings"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

var ErrNotFound = errors.New("not found")

func Initialize(db Queryer) error {
	mylog.Log.Info("Initializing database...")
	sql, err := ioutil.ReadFile("data/init_database.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(sql))
	return err
}

type Queryer interface {
	CopyFrom(pgx.Identifier, []string, pgx.CopyFromSource) (int, error)
	Exec(sql string, arguments ...interface{}) (pgx.CommandTag, error)
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
	QueryRow(sql string, args ...interface{}) *pgx.Row
}

type transactor interface {
	Begin() (*pgx.Tx, error)
}

type committer interface {
	Commit() error
	Rollback() error
}

type preparer interface {
	Prepare(name, sql string) (*pgx.PreparedStatement, error)
	Deallocate(name string) error
}

func BeginTransaction(db Queryer) (Queryer, error, bool) {
	if transactor, ok := db.(transactor); ok {
		tx, err := transactor.Begin()
		return tx, err, true
	}
	return db, nil, false
}

func CommitTransaction(db Queryer) error {
	if committer, ok := db.(committer); ok {
		return committer.Commit()
	}
	return nil
}

func RollbackTransaction(db Queryer) error {
	if committer, ok := db.(committer); ok {
		return committer.Rollback()
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

func (p *PageOptions) joins(from, as string, args *pgx.QueryArgs) []string {
	var joins []string
	if p.After != nil {
		joins = append(joins, fmt.Sprintf(
			"JOIN %[1]s AS %[2]s2 ON %[2]s2.id = "+args.Append(p.After.Value()),
			from,
			as,
		))
	}
	if p.Before != nil {
		joins = append(joins, fmt.Sprintf(
			"JOIN %[1]s AS %[2]s3 ON %[2]s3.id = "+args.Append(p.Before.Value()),
			from,
			as,
		))
	}
	return joins
}

func (p *PageOptions) whereAnds(as string) []string {
	var whereAnds []string
	field := p.Order.Field()
	if p.After != nil {
		relation := ""
		switch p.Order.Direction() {
		case ASC:
			relation = ">="
		case DESC:
			relation = "<="
		}
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %[1]s.%[2]s %[3]s %[1]s2.%[2]s",
			as,
			field,
			relation,
		))
	}
	if p.Before != nil {
		relation := ""
		switch p.Order.Direction() {
		case ASC:
			relation = "<="
		case DESC:
			relation = ">="
		}
		whereAnds = append(whereAnds, fmt.Sprintf(
			"AND %[1]s.%[2]s %[3]s %[1]s3.%[2]s",
			as,
			field,
			relation,
		))
	}
	return whereAnds
}

func SQL(
	selects []string,
	from string,
	where []string,
	args *pgx.QueryArgs,
	po *PageOptions,
) string {
	as := string(from[0])
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, as, args)
		whereAnds = po.whereAnds(as)
		limit = "LIMIT " + args.Append(po.Limit())
		orderBy = "ORDER BY " +
			as + "." + po.Order.Field() + " " + po.QueryDirection()
	}
	selectsCopy := make([]string, len(selects))
	for i, s := range selects {
		selectsCopy[i] = as + "." + s
	}
	whereCopy := make([]string, len(where))
	for i, w := range where {
		whereCopy[i] = as + "." + w
	}

	sql := `
		SELECT 
		` + strings.Join(selectsCopy, ",") + `
		FROM ` + from + ` AS ` + as + `
		` + strings.Join(joins, " ") + `
		WHERE ` + strings.Join(whereCopy, " AND ") + `
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
	as := string(from[0])
	var joins, whereAnds []string
	var limit, orderBy string
	if po != nil {
		joins = po.joins(from, as, &args)
		whereAnds = po.whereAnds(as)
		limit = "LIMIT " + args.Append(po.Limit())

		field := po.Order.Field()
		orderBy := ""
		if field != "best_match" {
			orderBy = as + "." + field
		} else {
			orderBy = "ts_rank(document, query)"
		}

		orderBy = "ORDER BY " + orderBy + " " + po.QueryDirection()
	}
	if within != nil {
		andIn := fmt.Sprintf(
			"AND %s.%s = %s",
			as,
			within.DBVarName(),
			args.Append(within),
		)
		whereAnds = append(whereAnds, andIn)
	}

	tsquery := ToTsQuery(query)
	sql := `
		SELECT 
		` + strings.Join(selects, ",") + `
		FROM ` + from + ` AS ` + as + `,
			to_tsquery('simple',` + args.Append(tsquery) + `) AS query
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
