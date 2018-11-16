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
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/xid"
)

var ErrNotFound = errors.New("not found")

func Initialize(db Queryer) error {
	mylog.Log.Info("Initializing database...")
	sql, err := ioutil.ReadFile("data/init_database.sql")
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	_, err = db.Exec(string(sql))
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	return nil
}

type Queryer interface {
	BeginBatch() *pgx.Batch
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
}

func BeginTransaction(db Queryer) (Queryer, error, bool) {
	if transactor, ok := db.(transactor); ok {
		tx, err := transactor.Begin()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err, false
		}
		return tx, nil, true
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

func prepare(db Queryer, name, sql string) (*pgx.PreparedStatement, error) {
	if preparer, ok := db.(preparer); ok {
		ps, err := preparer.Prepare(name, sql)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		return ps, nil
	}
	return nil, errors.New("db is not a preparer")
}

func prepareQuery(db Queryer, name, sql string, args ...interface{}) (*pgx.Rows, error) {
	// if preparer, ok := db.(preparer); ok {
	//   if _, err := preparer.Prepare(name, sql); err != nil {
	//     return nil, err
	//   }
	//   sql = name
	// }

	return db.Query(sql, args...)
}

func prepareQueryRow(db Queryer, name, sql string, args ...interface{}) *pgx.Row {
	// if preparer, ok := db.(preparer); ok {
	//   // QueryRow doesn't return an error, the error is encoded in the pgx.Row.
	//   // Since that is private, Ignore the error from Prepare and run the query
	//   // without the prepared statement. It should fail with the same error.
	//   if _, err := preparer.Prepare(name, sql); err == nil {
	//     sql = name
	//   }
	// }

	return db.QueryRow(sql, args...)
}

func prepareExec(db Queryer, name, sql string, args ...interface{}) (pgx.CommandTag, error) {
	// if preparer, ok := db.(preparer); ok {
	//   if _, err := preparer.Prepare(name, sql); err != nil {
	//     return pgx.CommandTag(""), err
	//   }
	//   sql = name
	// }

	return db.Exec(sql, args...)
}

func preparedName(baseName, sql string) string {
	h := fnv.New32a()
	if _, err := io.WriteString(h, sql); err != nil {
		// hash.Hash.Write never returns an error so this shouldn't happen
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
		err := fmt.Errorf("invalid OrderDirection: %q", s)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return o, err
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
		err := fmt.Errorf("You must provide a `first` or `last` value to properly paginate.")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	} else if first != nil {
		if last != nil {
			err := fmt.Errorf("Passing both `first` and `last` values to paginate the connection is not supported.")
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		pageOptions.First = *first
	} else {
		pageOptions.Last = *last
	}
	if after != nil {
		a, err := NewCursor(after)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		pageOptions.After = a
	}
	if before != nil {
		b, err := NewCursor(before)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
	limit := p.First + p.Last
	if limit < 1 {
		return 0
	}
	limit += 1
	if (p.After != nil && p.First > 0) ||
		(p.Before != nil && p.Last > 0) {
		limit = limit + int32(1)
	}
	return limit
}

func (p *PageOptions) joins(from, as string, args *pgx.QueryArgs) string {
	joins := make([]string, 0, 2)
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
	return strings.Join(joins, " ")
}

func (p *PageOptions) where(from string) string {
	where := make([]string, 0, 2)
	field := p.Order.Field()
	if field == "best_match" {
		field = "created_at"
	}
	if p.After != nil {
		relation := ""
		switch p.Order.Direction() {
		case ASC:
			relation = ">="
		case DESC:
			relation = "<="
		}
		where = append(where, fmt.Sprintf(
			"%[1]s.%[2]s %[3]s %[1]s2.%[2]s",
			from,
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
		where = append(where, fmt.Sprintf(
			"%[1]s.%[2]s %[3]s %[1]s3.%[2]s",
			from,
			field,
			relation,
		))
	}

	if len(where) == 0 {
		return ""
	}

	return "(" + strings.Join(where, " AND ") + ")"
}

func (p *PageOptions) orderBy(from string) string {
	field := p.Order.Field()
	if field == "best_match" {
		field = "created_at"
	}
	return fmt.Sprintf("ORDER BY %s.%s %s", from, field, p.QueryDirection())
}

type WhereFrom = func(string) string

func SQL3(
	selects []string,
	from string,
	where WhereFrom,
	filters FilterOptions,
	args *pgx.QueryArgs,
	po *PageOptions,
) string {
	fromAlias := xid.New().String()
	selectSQL := make([]string, len(selects))
	for i, s := range selects {
		selectSQL[i] = fromAlias + "." + s
	}
	fromSQL := []string{from + " AS " + fromAlias}
	joinSQL := []string{}
	whereSQL := []string{where(fromAlias)}

	var limit, orderBy string

	if po != nil {
		joinSQL = append(joinSQL, po.joins(from, fromAlias, args))
		whereSQL = append(whereSQL, po.where(fromAlias))
		limit = "LIMIT " + args.Append(po.Limit())
		orderBy = po.orderBy(fromAlias)
	}

	if filters != nil {
		filterSQL := filters.SQL(fromAlias, args)
		if filterSQL != nil {
			fromSQL = append(fromSQL, filterSQL.From)
			whereSQL = append(whereSQL, filterSQL.Where)
		}
	}

	fromSQL = util.RemoveEmptyStrings(fromSQL)
	whereSQL = util.RemoveEmptyStrings(whereSQL)

	sql := `
		SELECT 
		` + strings.Join(selectSQL, ",") + `
		FROM ` + strings.Join(fromSQL, ", ") + `
		` + strings.Join(joinSQL, " ") + `
		WHERE ` + strings.Join(whereSQL, " AND ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql)
}

func CountSQL(
	from string,
	where WhereFrom,
	filters FilterOptions,
	args *pgx.QueryArgs,
) string {
	fromAlias := xid.New().String()
	fromSQL := []string{from + " AS " + fromAlias}
	whereSQL := []string{where(fromAlias)}

	if filters != nil {
		filterSQL := filters.SQL(fromAlias, args)
		if filterSQL != nil {
			fromSQL = append(fromSQL, filterSQL.From)
			whereSQL = append(whereSQL, filterSQL.Where)
		}
	}

	fromSQL = util.RemoveEmptyStrings(fromSQL)
	whereSQL = util.RemoveEmptyStrings(whereSQL)

	return `
		SELECT count(` + fromAlias + `)
		FROM ` + strings.Join(fromSQL, ", ") + `
		WHERE ` + strings.Join(whereSQL, " AND ")
}

func SearchSQL2(
	selects []string,
	from string,
	query string,
	args *pgx.QueryArgs,
	po *PageOptions,
) string {
	fromAlias := xid.New().String()
	selectSQL := make([]string, len(selects))
	for i, s := range selects {
		selectSQL[i] = fromAlias + "." + s
	}
	fromSQL := []string{
		from + " AS " + fromAlias,
		"to_tsquery('simple', " + args.Append(query) + ") AS document_query",
	}
	joinSQL := []string{}
	whereSQL := []string{
		"CASE " + args.Append(query) + " WHEN '*' THEN TRUE ELSE " + fromAlias + ".document @@ document_query END",
	}

	var limit, orderBy string

	if po != nil {
		joinSQL = append(joinSQL, po.joins(from, fromAlias, args))
		whereSQL = append(whereSQL, po.where(fromAlias))
		limit = "LIMIT " + args.Append(po.Limit())
		orderBy = po.orderBy(fromAlias)
	}

	fromSQL = util.RemoveEmptyStrings(fromSQL)
	whereSQL = util.RemoveEmptyStrings(whereSQL)

	sql := `
		SELECT 
		` + strings.Join(selectSQL, ",") + `
		FROM ` + strings.Join(fromSQL, ", ") + `
		` + strings.Join(joinSQL, " ") + `
		WHERE ` + strings.Join(whereSQL, " AND ") + `
		` + orderBy + `
		` + limit

	return ReorderQuery(po, sql)
}

// Then, we can reorder the items to the originally requested direction.
func ReorderQuery(po *PageOptions, query string) string {
	if po != nil && po.Last != 0 {
		field := po.Order.Field()
		if field == "best_match" {
			field = "created_at"
		}
		return fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			query,
			field,
			po.Order.Direction(),
		)
	}
	return query
}

type SQLParts struct {
	From  string
	Where string
}

type FilterOptions interface {
	SQL(from string, args *pgx.QueryArgs) *SQLParts
}
