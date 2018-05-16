package route

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
)

func Upload() http.Handler {
	uploadHandler := UploadHandler{}
	return middleware.CommonMiddleware.Then(uploadHandler)
}

type UploadHandler struct{}

func (h UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	crutime := time.Now().Unix()
	hash := md5.New()
	io.WriteString(hash, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", hash.Sum(nil))

	t, err := template.ParseFiles("static/upload.gtpl")
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	t.Execute(rw, token)
}
