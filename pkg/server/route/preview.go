package route

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

var PreviewCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Authorization", "Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "http://localhost:*"},
})

type PreviewHandler struct {
	Repos *repo.Repos
}

func (h PreviewHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	query := req.URL.Query()
	study := query.Get("study")
	mylog.Log.Info(study)
	subject := query.Get("subject")
	mylog.Log.Info(subject)
	subjectType := query.Get("subject_type")
	mylog.Log.Info(subjectType)

	if err := req.ParseMultipartForm(10 * MB); err != nil {
		mylog.Log.WithError(err).Error("failed to parse multipart form")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	text := req.FormValue("text")
	previewHTML := util.MarkdownToHTML([]byte(text))
	rw.Write(previewHTML)

	return
}
