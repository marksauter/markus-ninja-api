package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Topic struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
	UserLogin   pgtype.Text        `db:"user_login" permit:"read"`
}

func NewTopicService(db Queryer) *TopicService {
	return &TopicService{db}
}

type TopicService struct {
	db Queryer
}

const countTopicByUserSQL = `
	SELECT COUNT(*)
	FROM topic
	WHERE user_id = $1
`

func (s *TopicService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("Topic.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countTopicByUser",
		countTopicByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *TopicService) CountBySearch(within *mytype.OID, query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("Topic.CountBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM topic_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" {
			// Only users 'contain' studies, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND topic_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countTopicBySearch", sql)

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)
	return
}

func (s *TopicService) get(name string, sql string, args ...interface{}) (*Topic, error) {
	var row Topic
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get topic")
		return nil, err
	}

	return &row, nil
}

func (s *TopicService) getMany(name string, sql string, args ...interface{}) ([]*Topic, error) {
	var rows []*Topic

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
			&row.UserId,
			&row.UserLogin,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("found rows")

	return rows, nil
}

const getTopicByIdSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		updated_at,
		user_id,
		user_login
	FROM topic_master
	WHERE id = $1
`

func (s *TopicService) Get(id string) (*Topic, error) {
	mylog.Log.WithField("id", id).Info("Topic.Get(id)")
	return s.get("getTopicById", getTopicByIdSQL, id)
}

func (s *TopicService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*Topic, error) {
	mylog.Log.WithField("user_id", userId).Info("Topic.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `topic_master.user_id = ` + args.Append(userId)

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "topic_master"
	sql := po.SQL(selects, from, whereSQL, &args)

	psName := preparedName("getStudiesByUserId", sql)

	return s.getMany(psName, sql, args...)
}

const getTopicByNameSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		updated_at,
		user_id,
		user_login
	FROM topic_master
	WHERE user_id = $1 AND LOWER(name) = LOWER($2)
`

func (s *TopicService) GetByName(userId, name string) (*Topic, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"name":    name,
	}).Info("Topic.GetByName(user_id, name)")
	return s.get("getTopicByName", getTopicByNameSQL, userId, name)
}

const getTopicByUserAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.updated_at,
		s.user_id,
		a.login user_login
	FROM topic s
	INNER JOIN account a ON a.login = $1
	WHERE s.user_id = a.id AND LOWER(s.name) = LOWER($2)  
`

func (s *TopicService) GetByUserAndName(owner, name string) (*Topic, error) {
	mylog.Log.WithFields(logrus.Fields{
		"owner": owner,
		"name":  name,
	}).Info("Topic.GetByUserAndName(owner, name)")
	return s.get("getTopicByUserAndName", getTopicByUserAndNameSQL, owner, name)
}

func (s *TopicService) Create(row *Topic) (*Topic, error) {
	mylog.Log.Info("Topic.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Topic")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.AdvancedAt.Status != pgtype.Undefined {
		columns = append(columns, "advanced_at")
		values = append(values, args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, topicDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		INSERT INTO topic(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createTopic", sql)

	_, err = prepareExec(s.db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create topic")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		return nil, err
	}

	topicSvc := NewTopicService(tx)
	topic, err := topicSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return topic, nil
}

const deleteTopicSQL = `
	DELETE FROM topic
	WHERE id = $1
`

func (s *TopicService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Topic.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteTopic", deleteTopicSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshTopicSearchIndexSQL = `
	REFRESH MATERIALIZED VIEW CONCURRENTLY topic_search_index
`

func (s *TopicService) RefreshSearchIndex() error {
	mylog.Log.Info("Topic.RefreshSearchIndex()")
	_, err := prepareExec(
		s.db,
		"refreshTopicSearchIndex",
		refreshTopicSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TopicService) Search(within *mytype.OID, query string, po *PageOptions) ([]*Topic, error) {
	mylog.Log.WithField("query", query).Info("Topic.Search(query)")
	if within != nil {
		if within.Type != "User" {
			return nil, fmt.Errorf(
				"cannot search for studies within type `%s`",
				within.Type,
			)
		}
	}
	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "topic_search_index"
	sql, args := po.SearchSQL(selects, from, within, query)

	psName := preparedName("searchTopicIndex", sql)

	return s.getMany(psName, sql, args...)
}

func (s *TopicService) Update(row *Topic) (*Topic, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Topic.Update(id)")
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.AdvancedAt.Status != pgtype.Undefined {
		sets = append(sets, `advanced_at`+"="+args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.TextArray{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, topicDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		UPDATE topic
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateTopic", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to update topic")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		return nil, err
	}

	topicSvc := NewTopicService(tx)
	topic, err := topicSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return topic, nil
}

func topicDelimeter(r rune) bool {
	return r == '-' || r == '_'
}
