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
	// hash := sha1.New()
	// io.Copy(hash, file)
	//
	// filename := fmt.Sprintf("%x", hash.Sum(nil))
	// fmt.Printf("File name %s \n", header.Filename)

	// file.Seek(0, 0)
	// basepath := path.Dir("test")
	// if err := os.MkdirAll(basepath, 0666); err != nil {
	//   response := myhttp.InternalServerErrorResponse(err.Error())
	//   myhttp.WriteResponseTo(rw, response)
	//   return
	// }
	// f, err := os.Create(strings.Join([]string{basepath, filename}, "/"))
	// if err != nil {
	//   response := myhttp.InternalServerErrorResponse(err.Error())
	//   myhttp.WriteResponseTo(rw, response)
	//   return
	// }
	// defer f.Close()
	// io.Copy(f, file)

	rw.WriteHeader(http.StatusCreated)
	return
}
