package data

import (
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type Role int

const (
	UnknownRole Role = iota
	AdminRole
	MemberRole
	SelfRole
	UserRole
)

func (r Role) String() string {
	switch r {
	case AdminRole:
		return "ADMIN"
	case MemberRole:
		return "MEMBER"
	case SelfRole:
		return "SELF"
	case UserRole:
		return "USER"
	default:
		return "UNKNOWN"
	}
}

type RoleModel struct {
	Id        oid.MaybeOID   `db:"id"`
	Name      pgtype.Varchar `db:"name"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type RoleService struct {
	db Queryer
}

func NewRoleService(db Queryer) *RoleService {
	return &RoleService{db}
}

var createRoleSQL = `
	INSERT INTO role(id, name)
	VALUES ($1, $2)
	ON CONFLICT ON CONSTRAINT role_name_key DO NOTHING
	RETURNING
		id,
		name,
		created_at,
		updated_at
`

func (s *RoleService) Create(row *RoleModel) error {
	mylog.Log.WithField("name", row.Name.String).Info("Create(name) Role")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	id, _ := oid.New("Role")
	row.Id.Just(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, `name`)
		values = append(values, args.Append(&row.Name))
	}

	createRoleSQL := `
		INSERT INTO role(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT ON CONSTRAINT role_name_key DO NOTHING
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createRole", createRoleSQL)

	err := prepareQueryRow(s.db, psName, createRoleSQL, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return err
	}

	return nil
}

func (s *RoleService) GetByUserId(userId string) ([]RoleModel, error) {
	roles := make([]RoleModel, 0)

	roleSQL := `
		SELECT
			id,
			name,
			created_at,
		FROM
			role
		INNER JOIN account_role ar ON role.id = ar.role_id
		WHERE ar.user_id = $1
	`
	rows, err := s.db.Query(roleSQL, userId)
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during query")
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		r := roles[i]
		err := rows.Scan(&r.Id, &r.Name, &r.CreatedAt)
		if err != nil {
			mylog.Log.WithField("error", err).Error("error during scan")
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		mylog.Log.WithField("error", err).Error("error during rows processing")
		return nil, err
	}

	mylog.Log.Debug("roles found")
	return roles, nil
}

const grantUserRolesSQL = `
	INSERT INTO account_role (user_id, role_id)
	SELECT DISTINCT a.id, r.id
	FROM account a
	INNER JOIN role r ON r.name = ANY($1)
	WHERE a.id = $2
`

func (s *RoleService) GrantUser(userId string, roles ...Role) error {
	if len(roles) > 0 {
		roleArgs := make([]string, len(roles))
		for i, r := range roles {
			roleArgs[i] = r.String()
		}
		_, err := prepareExec(
			s.db,
			"grantUserRoles",
			grantUserRolesSQL,
			roleArgs,
			userId,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error during execution")
			return err
		}
	}

	return nil
}
