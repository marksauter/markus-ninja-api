package route

import (
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"

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
	studyID := query.Get("study")

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

	markdown := &mytype.Markdown{}
	if err := markdown.Set(text); err != nil {
		mylog.Log.WithError(err).Error("failed to set markdown to text")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	markdown, err, _ := h.Repos.ReplaceMarkdownRefsWithLinks(req.Context(), *markdown, studyID)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to replace markdown refs with links")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.Write([]byte(markdown.ToHTML()))

	return
}
