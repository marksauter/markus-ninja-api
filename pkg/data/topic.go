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
	StudyId     mytype.OID         `db:"study_id"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func NewTopicService(db Queryer) *TopicService {
	return &TopicService{db}
}

type TopicService struct {
	db Queryer
}

const countTopicByStudySQL = `
	SELECT COUNT(*)
	FROM study_topic
	WHERE study_id = $1
`

func (s *TopicService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("Topic.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countTopicByStudy",
		countTopicByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

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
		if within.Type != "Study" {
			// Only studies 'contain' topics, so return 0 otherwise
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

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *TopicService) get(name string, sql string, args ...interface{}) (*Topic, error) {
	var row Topic
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
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
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get topics")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getTopicByIdSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic_master
	WHERE id = $1
`

func (s *TopicService) Get(id string) (*Topic, error) {
	mylog.Log.WithField("id", id).Info("Topic.Get(id)")
	return s.get("getTopicById", getTopicByIdSQL, id)
}

const getTopicNamesByStudySQL = `
	SELECT
		array_agg(name) topic_names
	FROM study_topic_master
	WHERE study_id = $1
	GROUP BY study_id
`

func (s *TopicService) GetNamesByStudy(studyId string) (names []string, err error) {
	mylog.Log.WithField("study_id", studyId).Info("Topic.GetNamesByStudy(study_id)")
	topicNames := pgtype.TextArray{}
	err = prepareQueryRow(
		s.db,
		"getTopicNamesByStudy",
		getTopicNamesByStudySQL,
		studyId,
	).Scan(topicNames)
	if err == pgx.ErrNoRows {
		return
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get topic")
		return
	}

	err = topicNames.AssignTo(names)
	return
}

func (s *TopicService) GetByStudy(
	studyId string,
	po *PageOptions,
) ([]*Topic, error) {
	mylog.Log.WithField("study_id", studyId).Info("Topic.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_topic_master.study_id = ` + args.Append(studyId)

	selects := []string{
		"created_at",
		"description",
		"id",
		"name",
		"study_id",
		"updated_at",
	}
	from := "study_topic_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getTopicsByStudyId", sql)

	var rows []*Topic

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Topic
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.StudyId,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get topics")
		return nil, err
	}

	return rows, nil
}

const getTopicByNameSQL = `
	SELECT
		created_at,
		description,
		id,
		name,
		updated_at
	FROM topic_master
	WHERE LOWER(name) = LOWER($1)
`

func (s *TopicService) GetByName(name string) (*Topic, error) {
	mylog.Log.WithFields(logrus.Fields{
		"name": name,
	}).Info("Topic.GetByName(user_id, name)")
	return s.get("getTopicByName", getTopicByNameSQL, name)
}

const createStudyTopicSQL = `
	INSERT INTO study_topic(study_id, topic_id)
	VALUES ($1, $2)
`

func (s *TopicService) Create(row *Topic) (*Topic, error) {
	mylog.Log.Info("Topic.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Topic")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO topic(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT(lower("name")) DO UPDATE SET name=EXCLUDED.name RETURNING id
	`

	psName := preparedName("createTopic", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
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

	if row.StudyId.Status != pgtype.Undefined {
		commandTag, err := prepareExec(
			tx,
			"createStudyTopic",
			createStudyTopicSQL,
			row.StudyId.String,
			row.Id.String,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to create study topic")
			return nil, err
		}
		if commandTag.RowsAffected() != 1 {
			mylog.Log.WithError(err).Error("failed to create study topic")
			return nil, ErrNotFound
		}
	}

	topicSvc := NewTopicService(tx)
	topic, err := topicSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return topic, nil
}

const deleteTopicStudyRelationSQL = `
	DELETE FROM study_topic
	WHERE study_id = $1 AND topic_id = $2
`

func (s *TopicService) DeleteStudyRelation(studyId, topicId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"topic_id": topicId,
	}).Info("Topic.DeleteStudyRelation()")
	commandTag, err := prepareExec(
		s.db,
		"deleteTopicStudyRelation",
		deleteTopicStudyRelationSQL,
		studyId,
		topicId,
	)
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

func (s *TopicService) Search(query string, po *PageOptions) ([]*Topic, error) {
	mylog.Log.WithField("query", query).Info("Topic.Search(query)")
	selects := []string{
		"created_at",
		"description",
		"id",
		"name",
		"updated_at",
	}
	from := "topic_search_index"
	sql, args := SearchSQL(selects, from, nil, query, po)

	psName := preparedName("searchTopicIndex", sql)

	return s.getMany(psName, sql, args...)
}

func (s *TopicService) Update(row *Topic) (*Topic, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Topic.Update(id)")
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		UPDATE topic
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateTopic", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	topicSvc := NewTopicService(tx)
	topic, err := topicSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return topic, nil
}

func topicDelimeter(r rune) bool {
	return r == '-'
}
