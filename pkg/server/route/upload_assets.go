package route

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func UploadAssets() http.Handler {
	uploadAssetsHandler := UploadAssetsHandler{}
	return middleware.CommonMiddleware.Then(uploadAssetsHandler)
}

type UploadAssetsHandler struct{}

func (h UploadAssetsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(32 << 20)
	file, header, err := req.FormFile("file")
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	defer file.Close()
	svc, err := service.NewUploadService()
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	err = svc.Upload(file, header)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	return
}

type Asset struct {
	Id          string
	Name        string
	Size        int64
	ContentType string
	Href        string
}

type UploadSuccessResponse struct {
	UploadUrl string
	Asset     Asset
}
