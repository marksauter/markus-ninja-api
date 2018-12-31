package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type key string

const (
	activityRepoKey           key = "activity"
	activityAssetRepoKey      key = "activity_asset"
	appledRepoKey             key = "appled"
	assetRepoKey              key = "asset"
	commentRepoKey            key = "comment"
	commentDraftBackupRepoKey key = "comment_draft_backup"
	courseRepoKey             key = "course"
	courseLessonRepoKey       key = "course_lesson"
	emailRepoKey              key = "email"
	enrolledRepoKey           key = "enrolled"
	evtRepoKey                key = "evt"
	labelRepoKey              key = "label"
	labeledRepoKey            key = "labeled"
	lessonRepoKey             key = "lesson"
	lessonDraftBackupRepoKey  key = "lesson_draft_backup"
	notificationRepoKey       key = "notification"
	permRepoKey               key = "perm"
	prtRepoKey                key = "prt"
	eventRepoKey              key = "event"
	studyRepoKey              key = "study"
	topicRepoKey              key = "topic"
	topicableRepoKey          key = "topicable"
	topicedRepoKey            key = "topiced"
	userRepoKey               key = "user"
	userAssetRepoKey          key = "user_asset"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")
var ErrFieldAccessDenied = errors.New("field access denied")

type FieldPermissionFunc = func(field string) bool

var AdminPermissionFunc FieldPermissionFunc = func(field string) bool { return true }

type Repo interface {
	Open(*Permitter) error
	Close()
}

type Repos struct {
	conf   *myconf.Config
	db     data.Queryer
	lookup map[key]Repo
}

func NewRepos(db data.Queryer, conf *myconf.Config) *Repos {
	return &Repos{
		conf: conf,
		db:   db,
		lookup: map[key]Repo{
			activityRepoKey:           NewActivityRepo(conf),
			activityAssetRepoKey:      NewActivityAssetRepo(conf),
			appledRepoKey:             NewAppledRepo(conf),
			assetRepoKey:              NewAssetRepo(conf),
			commentRepoKey:            NewCommentRepo(conf),
			commentDraftBackupRepoKey: NewCommentDraftBackupRepo(conf),
			courseRepoKey:             NewCourseRepo(conf),
			courseLessonRepoKey:       NewCourseLessonRepo(conf),
			emailRepoKey:              NewEmailRepo(conf),
			enrolledRepoKey:           NewEnrolledRepo(conf),
			evtRepoKey:                NewEVTRepo(conf),
			labelRepoKey:              NewLabelRepo(conf),
			labeledRepoKey:            NewLabeledRepo(conf),
			lessonRepoKey:             NewLessonRepo(conf),
			lessonDraftBackupRepoKey:  NewLessonDraftBackupRepo(conf),
			notificationRepoKey:       NewNotificationRepo(conf),
			prtRepoKey:                NewPRTRepo(conf),
			eventRepoKey:              NewEventRepo(conf),
			studyRepoKey:              NewStudyRepo(conf),
			topicRepoKey:              NewTopicRepo(conf),
			topicedRepoKey:            NewTopicedRepo(conf),
			userRepoKey:               NewUserRepo(conf),
			userAssetRepoKey:          NewUserAssetRepo(conf),
		},
	}
}

var ErrNilPermitter = errors.New("permitter is nil")

func (r *Repos) OpenAll(p *Permitter) error {
	for _, repo := range r.lookup {
		if err := repo.Open(p); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
	}
	return nil
}

func (r *Repos) CloseAll() {
	for _, repo := range r.lookup {
		repo.Close()
	}
}

func (r *Repos) Activity() *ActivityRepo {
	repo, _ := r.lookup[activityRepoKey].(*ActivityRepo)
	return repo
}

func (r *Repos) ActivityAsset() *ActivityAssetRepo {
	repo, _ := r.lookup[activityAssetRepoKey].(*ActivityAssetRepo)
	return repo
}

func (r *Repos) Appled() *AppledRepo {
	repo, _ := r.lookup[appledRepoKey].(*AppledRepo)
	return repo
}

func (r *Repos) Asset() *AssetRepo {
	repo, _ := r.lookup[assetRepoKey].(*AssetRepo)
	return repo
}

func (r *Repos) Comment() *CommentRepo {
	repo, _ := r.lookup[commentRepoKey].(*CommentRepo)
	return repo
}

func (r *Repos) CommentDraftBackup() *CommentDraftBackupRepo {
	repo, _ := r.lookup[commentDraftBackupRepoKey].(*CommentDraftBackupRepo)
	return repo
}

func (r *Repos) Course() *CourseRepo {
	repo, _ := r.lookup[courseRepoKey].(*CourseRepo)
	return repo
}

func (r *Repos) CourseLesson() *CourseLessonRepo {
	repo, _ := r.lookup[courseLessonRepoKey].(*CourseLessonRepo)
	return repo
}

func (r *Repos) Email() *EmailRepo {
	repo, _ := r.lookup[emailRepoKey].(*EmailRepo)
	return repo
}

func (r *Repos) Enrolled() *EnrolledRepo {
	repo, _ := r.lookup[enrolledRepoKey].(*EnrolledRepo)
	return repo
}

func (r *Repos) Event() *EventRepo {
	repo, _ := r.lookup[eventRepoKey].(*EventRepo)
	return repo
}

func (r *Repos) EVT() *EVTRepo {
	repo, _ := r.lookup[evtRepoKey].(*EVTRepo)
	return repo
}

func (r *Repos) Label() *LabelRepo {
	repo, _ := r.lookup[labelRepoKey].(*LabelRepo)
	return repo
}

func (r *Repos) Labeled() *LabeledRepo {
	repo, _ := r.lookup[labeledRepoKey].(*LabeledRepo)
	return repo
}

func (r *Repos) Lesson() *LessonRepo {
	repo, _ := r.lookup[lessonRepoKey].(*LessonRepo)
	return repo
}

func (r *Repos) LessonDraftBackup() *LessonDraftBackupRepo {
	repo, _ := r.lookup[lessonDraftBackupRepoKey].(*LessonDraftBackupRepo)
	return repo
}

func (r *Repos) Notification() *NotificationRepo {
	repo, _ := r.lookup[notificationRepoKey].(*NotificationRepo)
	return repo
}

func (r *Repos) PRT() *PRTRepo {
	repo, _ := r.lookup[prtRepoKey].(*PRTRepo)
	return repo
}

func (r *Repos) Study() *StudyRepo {
	repo, _ := r.lookup[studyRepoKey].(*StudyRepo)
	return repo
}

func (r *Repos) Topic() *TopicRepo {
	repo, _ := r.lookup[topicRepoKey].(*TopicRepo)
	return repo
}

func (r *Repos) Topiced() *TopicedRepo {
	repo, _ := r.lookup[topicedRepoKey].(*TopicedRepo)
	return repo
}

func (r *Repos) User() *UserRepo {
	repo, _ := r.lookup[userRepoKey].(*UserRepo)
	return repo
}

func (r *Repos) UserAsset() *UserAssetRepo {
	repo, _ := r.lookup[userAssetRepoKey].(*UserAssetRepo)
	return repo
}

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		permitter := NewPermitter(r, r.conf)
		defer permitter.ClearCache()
		r.OpenAll(permitter)
		defer r.CloseAll()

		ctx := myctx.NewQueryerContext(req.Context(), r.db)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// Cross repo methods

func (r *Repos) GetAppleable(
	ctx context.Context,
	appleableID *mytype.OID,
) (NodePermit, error) {
	switch appleableID.Type {
	case "Course":
		return r.Course().Get(ctx, appleableID.String)
	case "Study":
		return r.Study().Get(ctx, appleableID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for appleable id", appleableID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetCommentable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for commentable id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetCreateable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Activity":
		return r.Activity().Get(ctx, nodeID.String)
	case "Course":
		return r.Course().Get(ctx, nodeID.String)
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "Study":
		return r.Study().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for createable id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetEnrollable(
	ctx context.Context,
	enrollableID *mytype.OID,
) (NodePermit, error) {
	switch enrollableID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, enrollableID.String)
	case "Study":
		return r.Study().Get(ctx, enrollableID.String)
	case "User":
		return r.User().Get(ctx, enrollableID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for enrollable id", enrollableID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetPublishable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Course":
		return r.Course().Get(ctx, nodeID.String)
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "Comment":
		return r.Comment().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for publishable id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetReferenceable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for referenceable id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetRenameable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for renameable id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetLabelable(
	ctx context.Context,
	labelableID *mytype.OID,
) (NodePermit, error) {
	switch labelableID.Type {
	case "Comment":
		return r.Comment().Get(ctx, labelableID.String)
	case "Lesson":
		return r.Lesson().Get(ctx, labelableID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, labelableID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for labelable id", labelableID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetNode(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Activity":
		return r.Activity().Get(ctx, nodeID.String)
	case "Comment":
		return r.Comment().Get(ctx, nodeID.String)
	case "Course":
		return r.Course().Get(ctx, nodeID.String)
	case "Email":
		return r.Email().Get(ctx, nodeID.String)
	case "Event":
		return r.Event().Get(ctx, nodeID.String)
	case "Label":
		return r.Label().Get(ctx, nodeID.String)
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "Notification":
		return r.Notification().Get(ctx, nodeID.String)
	case "Study":
		return r.Study().Get(ctx, nodeID.String)
	case "Topic":
		return r.Topic().Get(ctx, nodeID.String)
	case "User":
		return r.User().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for node id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetNotificationSubject(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for notification subject id", nodeID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) GetTopicable(
	ctx context.Context,
	topicableID *mytype.OID,
) (NodePermit, error) {
	switch topicableID.Type {
	case "Course":
		return r.Course().Get(ctx, topicableID.String)
	case "Study":
		return r.Study().Get(ctx, topicableID.String)
	default:
		err := fmt.Errorf("invalid type '%s' for topicable id", topicableID.Type)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
}

func (r *Repos) ReplaceMarkdownRefsWithLinks(
	ctx context.Context,
	markdown mytype.Markdown,
	studyID string,
) (*mytype.Markdown, error, bool) {
	updated := false
	body := markdown.String
	studyPermit, err := r.Study().Get(ctx, studyID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err, false
	}
	study := studyPermit.Get()

	userPermit, err := r.User().Get(ctx, study.UserID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err, false
	}
	user := userPermit.Get()

	userAssetRefToLink := func(s string) string {
		result := mytype.AssetRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		name := result[1]
		class := ""
		caption := ""
		query := ""
		var queryValues url.Values
		if len(result) >= 3 {
			query = result[2]
			queryValues, err = url.ParseQuery(query)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return s
			}
			class = queryValues.Get("class")
			class = util.RemoveQuotes(class)
			caption = queryValues.Get("caption")
			caption = util.RemoveQuotes(caption)
			queryValues.Del("class")
			queryValues.Del("caption")
			query = queryValues.Encode()
		}
		userAssetPermit, err := r.UserAsset().GetByName(
			ctx,
			studyID,
			name,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return s
		}
		userAsset := userAssetPermit.Get()

		updated = true
		src := fmt.Sprintf(
			r.conf.ImagesURL+"/%s/%s",
			userAsset.UserID.Short,
			userAsset.Key.String,
		)
		if query != "" {
			src += "?" + query
		}
		href := fmt.Sprintf(
			r.conf.ClientURL+"/u/%s/%s/asset/%s",
			user.Login.String,
			study.Name.String,
			userAsset.Name.String,
		)
		link := `<figure`
		if class != "" {
			link += ` class="` + class + `"`
		}
		link += `><!---USER_ASSET_LINK--->`
		link += `<a href="` + href + `" rel="nofollow"><img src="` +
			src + `" alt="` + name + `"/></a>`
		if caption != "" {
			link += `<figcaption>` + caption + `</figcaption>`
		}
		link += `</figure>`
		link = fmt.Sprintf("\n\n%s\n\n", link)

		return util.ReplaceWithPadding(s, link)
	}
	body = mytype.AssetRefRegexp.ReplaceAllStringFunc(body, userAssetRefToLink)

	lessonNumberRefToLink := func(s string) string {
		result := mytype.NumberRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		number := result[1]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Lesson().ExistsByNumber(ctx, studyID, int32(n))
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return s
		}
		if !exists {
			return s
		}

		updated = true
		href := fmt.Sprintf(
			r.conf.ClientURL+"/u/%s/%s/lesson/%d",
			user.Login.String,
			study.Name.String,
			n,
		)
		return util.ReplaceWithPadding(s, fmt.Sprintf("<!---LESSON_LINK--->[#%d](%s)", n, href))
	}
	body = mytype.NumberRefRegexp.ReplaceAllStringFunc(body, lessonNumberRefToLink)

	crossStudyRefToLink := func(s string) string {
		result := mytype.CrossStudyRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		owner := result[1]
		name := result[2]
		number := result[3]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Lesson().ExistsByOwnerStudyAndNumber(ctx, owner, name, int32(n))
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return s
		}
		if !exists {
			return s
		}

		updated = true
		link := fmt.Sprintf("%s/%s#%d", owner, name, n)
		href := fmt.Sprintf(
			r.conf.ClientURL+"/u/%s/%s/lesson/%d",
			owner,
			name,
			n,
		)
		return util.ReplaceWithPadding(s, fmt.Sprintf("<!---STUDY_LINK--->[%s](%s)", link, href))
	}
	body = mytype.CrossStudyRefRegexp.ReplaceAllStringFunc(body, crossStudyRefToLink)

	userRefToLink := func(s string) string {
		result := mytype.AtRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		name := result[1]
		exists, err := r.User().ExistsByLogin(ctx, name)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return s
		}
		if !exists {
			return s
		}

		updated = true
		href := fmt.Sprintf(r.conf.ClientURL+"/u/%s", user.Login.String)
		return util.ReplaceWithPadding(s, fmt.Sprintf("<!---USER_LINK--->[@%s](%s)", name, href))
	}
	body = mytype.AtRefRegexp.ReplaceAllStringFunc(body, userRefToLink)

	if err := markdown.Set(body); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err, false
	}

	return &markdown, nil, updated
}

var UserLinkRegexp = regexp.MustCompile(`<!---USER_LINK--->\[(@\w+)\]\(.*?\)`)
var UserAssetLinkRegexp = regexp.MustCompile(`<figure(?: class="(.*)")?><!---USER_ASSET_LINK---><a href=".*" rel="nofollow"><img src="(.*)" alt="(.*)"\/><\/a>(?:<figcaption>(.*)<\/figcaption>)?<\/figure>`)
var LessonLinkRegexp = regexp.MustCompile(`<!---LESSON_LINK--->\[(#\d+)\]\(.*?\)`)
var StudyLinkRegexp = regexp.MustCompile(`<!---STUDY_LINK--->\[(\w+\/[\w-]{1,39}#\d+)\]\(.*?\)`)

func (r *Repos) ReplaceMarkdownLinksWithRefs(
	ctx context.Context,
	markdown,
	studyID string,
) (string, error, bool) {
	updated := false

	userAssetLinkToRef := func(s string) string {
		result := UserAssetLinkRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		class := ""
		name := ""
		src := ""
		query := ""
		caption := ""
		if len(result) > 1 {
			class = result[1]
		}
		if len(result) > 2 {
			src = result[2]
			uri, err := url.ParseRequestURI(src)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return s
			}
			query = uri.RawQuery
		}
		if len(result) > 3 {
			name = result[3]
		}
		if len(result) > 4 {
			caption = result[4]
		}
		updated = true
		ref := `$$` + name
		queries := []string{query, `class="` + class + `"`, `caption="` + caption + `"`}
		queries = util.RemoveEmptyStrings(queries)
		if len(queries) > 0 {
			ref += `?` + strings.Join(queries, "&") + `?`
		}
		return ref
	}
	markdown = UserAssetLinkRegexp.ReplaceAllStringFunc(markdown, userAssetLinkToRef)

	lessonNumberLinkToRef := func(s string) string {
		result := LessonLinkRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		updated = true
		return result[1]
	}
	markdown = LessonLinkRegexp.ReplaceAllStringFunc(markdown, lessonNumberLinkToRef)

	crossStudyLinkToRef := func(s string) string {
		result := StudyLinkRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		updated = true
		return result[1]
	}
	markdown = StudyLinkRegexp.ReplaceAllStringFunc(markdown, crossStudyLinkToRef)

	userLinkToRef := func(s string) string {
		result := UserLinkRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		updated = true
		return result[1]
	}
	markdown = UserLinkRegexp.ReplaceAllStringFunc(markdown, userLinkToRef)

	return markdown, nil, updated
}

func (r *Repos) ParseLessonBodyForEvents(
	ctx context.Context,
	body *mytype.Markdown,
	lessonID,
	studyID,
	userID *mytype.OID,
) error {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	userAssetRefs := body.AssetRefs()
	if len(userAssetRefs) > 0 {
		names := make([]string, len(userAssetRefs))
		for i, ref := range userAssetRefs {
			names[i] = ref.Name
		}
		userAssets, err := r.UserAsset().BatchGetByName(
			ctx,
			studyID.String,
			names,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		for _, a := range userAssets {
			aID, err := a.ID()
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			payload, err := data.NewUserAssetReferencedPayload(aID, lessonID)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			event, err := data.NewUserAssetEvent(payload, studyID, userID, true)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			if _, err = r.Event().Create(ctx, event); err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
		}
	}
	lessonNumberRefs, err := body.NumberRefs()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if len(lessonNumberRefs) > 0 {
		numbers := make([]int32, len(lessonNumberRefs))
		for i, ref := range lessonNumberRefs {
			numbers[i] = ref.Number
		}
		lessons, err := r.Lesson().BatchGetByNumber(
			ctx,
			studyID.String,
			numbers,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		for _, l := range lessons {
			lID, err := l.ID()
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			if lID.String != lessonID.String {
				payload, err := data.NewLessonReferencedPayload(lID, lessonID)
				if err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
				event, err := data.NewLessonEvent(payload, studyID, userID, true)
				if err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
				if _, err = r.Event().Create(ctx, event); err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
			}
		}
	}
	crossStudyRefs, err := body.CrossStudyRefs()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	for _, ref := range crossStudyRefs {
		l, err := r.Lesson().GetByOwnerStudyAndNumber(
			ctx,
			ref.Owner,
			ref.Name,
			ref.Number,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		lID, err := l.ID()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		if lID.String != lessonID.String {
			payload, err := data.NewLessonReferencedPayload(lID, lessonID)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			event, err := data.NewLessonEvent(payload, studyID, userID, true)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			if _, err = r.Event().Create(ctx, event); err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
		}
	}
	userRefs := body.AtRefs()
	if len(userRefs) > 0 {
		names := make([]string, len(userRefs))
		for i, ref := range userRefs {
			names[i] = ref.Name
		}
		users, err := r.User().BatchGetByLogin(
			ctx,
			names,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		for _, u := range users {
			uID, err := u.ID()
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			if uID.String != userID.String {
				payload, err := data.NewLessonMentionedPayload(lessonID)
				if err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
				event, err := data.NewLessonEvent(payload, studyID, userID, true)
				if err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
				if _, err = r.Event().Create(ctx, event); err != nil {
					mylog.Log.WithError(err).Error(util.Trace(""))
					return err
				}
			}
		}
	}

	if newTx {
		err = data.CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
	}

	return nil
}
