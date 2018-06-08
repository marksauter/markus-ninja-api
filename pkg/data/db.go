package data

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"html/template"
	"io"
	"strings"

	"github.com/jackc/pgx"
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

// Then, we can reorder the items to the originally requested direction.
func (p *PageOptions) ReorderQuery(query string) string {
	if p.Last != 0 {
		return fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			query,
			p.Order.Field(),
			p.Order.Direction(),
		)
	}
	return query
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

type SQLTemplateArgs struct {
	Select string
	From   string
	As     string
	Joins  string
	Where  string
}

func (p *PageOptions) SQL(tmplArgs *SQLTemplateArgs, queryArgs *pgx.QueryArgs) (string, error) {
	var joins, whereAnds []string
	field := p.Order.Field()
	if p.After != nil {
		joins = append(joins, `INNER JOIN {{.From}} {{.As}}2 ON {{.As}}2.id = `+
			queryArgs.Append(p.After.Value()),
		)
		whereAnds = append(whereAnds, `AND {{.As}}.`+field+` >= {{.As}}2.`+field)
	}
	if p.Before != nil {
		joins = append(joins, `INNER JOIN {{.From}} {{.As}}3 ON {{.As}}3.id = `+
			queryArgs.Append(p.Before.Value()),
		)
		whereAnds = append(whereAnds, `AND {{.As}}.`+field+` <= {{.As}}3.`+field)
	}
	tmpl := template.Must(
		template.New("connection").Parse(`
			SELECT 
			` + tmplArgs.Select + `
			FROM ` + tmplArgs.From +
			strings.Join(joins, " ") +
			tmplArgs.Joins + `
			WHERE (` + tmplArgs.Where + `)
			` + strings.Join(whereAnds, " ") + `
			ORDER BY {{.As}}.` + field + ` ` + p.QueryDirection() + `
			LIMIT ` + queryArgs.Append(p.Limit()),
		),
	)

	var bs bytes.Buffer
	err := tmpl.Execute(&bs, tmplArgs)
	if err != nil {
		return "", err
	}
	return p.ReorderQuery(bs.String()), nil
}
