package route

import (
	"context"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"
	"strconv"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type PreviewHandler struct {
	Conf  *myconf.Config
	Repos *repo.Repos
}

func (h PreviewHandler) Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{http.MethodOptions, http.MethodPost},
		AllowedOrigins:   []string{h.Conf.ClientURL},
	})
}

func (h PreviewHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.Conf == nil || h.Repos == nil {
		err := errors.New("route inproperly setup")
		mylog.Log.WithError(err).Error(util.Trace(""))
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	query := req.URL.Query()
	study := query.Get("study")

	if err := req.ParseMultipartForm(10 * MB); err != nil {
		mylog.Log.WithError(err).Error("failed to parse multipart form")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	text := req.FormValue("text")
	if text == "" {
		rw.Write([]byte(""))
		return
	}

	carriageReturnNewline := regexp.MustCompile(`\r\n`)
	text = carriageReturnNewline.ReplaceAllString(text, "\n")

	html, err := h.textToHTML(req.Context(), text, study)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to convert text to html")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.Write([]byte(html))

	return
}

func (h PreviewHandler) textToHTML(ctx context.Context, text, studyID string) (string, error) {
	markdown := mytype.Markdown{}
	if err := markdown.Set(text); err != nil {
		return "", err
	}
	markdownStr := markdown.String

	userAssetRefToLink := func(s string) string {
		result := mytype.AssetRefRegexp.FindStringSubmatch(s)
		if len(result) < 2 {
			return s
		}
		name := result[1]
		userAsset, err := h.Repos.UserAsset().GetByName(
			ctx,
			studyID,
			name,
		)
		if err != nil && err != data.ErrNotFound {
			return s
		}
		userID, err := userAsset.UserID()
		if err != nil {
			return s
		}
		key, err := userAsset.Key()
		if err != nil {
			return s
		}
		href := fmt.Sprintf(
			h.Conf.APIURL+"/user/assets/%s/%s",
			userID.Short,
			key,
		)
		return util.ReplaceWithPadding(s, fmt.Sprintf("![%s](%s)", name, href))
	}
	markdownStr = mytype.AssetRefRegexp.ReplaceAllStringFunc(markdownStr, userAssetRefToLink)

	study, err := h.Repos.Study().Get(ctx, studyID)
	if err != nil {
		return "", err
	}
	studyName, err := study.Name()
	if err != nil {
		return "", err
	}
	userID, err := study.UserID()
	if err != nil {
		return "", err
	}
	user, err := h.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return "", err
	}
	userLogin, err := user.Login()
	if err != nil {
		return "", err
	}

	lessonNumberRefToLink := func(s string) string {
		result := mytype.NumberRefRegexp.FindStringSubmatch(s)
		if len(result) < 2 {
			return s
		}
		number := result[1]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := h.Repos.Lesson().ExistsByNumber(ctx, studyID, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[#%[4]d](%[1]s/%[2]s/%[3]s/lesson/%[4]d)",
			h.Conf.ClientURL,
			userLogin,
			studyName,
			n,
		))
	}
	markdownStr = mytype.NumberRefRegexp.ReplaceAllStringFunc(markdownStr, lessonNumberRefToLink)

	crossStudyRefToLink := func(s string) string {
		result := mytype.CrossStudyRefRegexp.FindStringSubmatch(s)
		if len(result) < 4 {
			return s
		}
		owner := result[1]
		name := result[2]
		number := result[3]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := h.Repos.Lesson().ExistsByOwnerStudyAndNumber(ctx, owner, name, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[%[2]s/%[3]s#%[4]d](%[1]s/%[2]s/%[3]s/lesson/%[4]d)",
			h.Conf.ClientURL,
			owner,
			name,
			n,
		))
	}
	markdownStr = mytype.CrossStudyRefRegexp.ReplaceAllStringFunc(markdownStr, crossStudyRefToLink)

	userRefToLink := func(s string) string {
		result := mytype.AtRefRegexp.FindStringSubmatch(s)
		if len(result) < 2 {
			return s
		}
		name := result[1]
		exists, err := h.Repos.User().ExistsByLogin(ctx, name)
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[@%[2]s](%[1]s/%[2]s)",
			h.Conf.ClientURL,
			name,
		))
	}
	markdownStr = mytype.AtRefRegexp.ReplaceAllStringFunc(markdownStr, userRefToLink)

	if err := markdown.Set(markdownStr); err != nil {
		return "", err
	}

	return markdown.ToHTML(), nil
}
