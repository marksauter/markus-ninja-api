package route

import (
	"net/http"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func UploadAssets(svcs *service.Services, repos *repo.Repos) http.Handler {
	authMiddleware := middleware.Authenticate{
		Svcs: svcs,
	}
	uploadAssetsHandler := UploadAssetsHandler{Repos: repos}
	return middleware.CommonMiddleware.Append(
		authMiddleware.Use,
	).Then(uploadAssetsHandler)
}

type UploadAssetsHandler struct {
	Repos *repo.Repos
}

func (h UploadAssetsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	file, header, err := req.FormFile("file")
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get form file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image") {
		mylog.Log.WithField("type", contentType).Error("attempt to upload non-image file")
		response := myhttp.InvalidRequestErrorResponse("file is not an image")
		myhttp.WriteResponseTo(rw, response)
		return
	}

	userAssetPermit, err := h.Repos.UserAsset().Upload(req.Context(), file, header)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to upload file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	href, err := userAssetPermit.Href()
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get user asset href")
		response := myhttp.AccessDeniedErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	assetId, err := userAssetPermit.ID()
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get user asset id")
		response := myhttp.AccessDeniedErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	asset := Asset{
		ContentType: contentType,
		Href:        href,
		id:          assetId.Short,
		name:        header.Filename,
		size:        header.Size,
	}

	response := &UploadAssetsSuccessResponse{Asset: asset}
	myhttp.WriteResponseTo(rw, response)
	return
}

type Asset struct {
	ContentType string `json:"content_type,omitempty"`
	Href        string `json:"href,omitempty"`
	id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

type UploadAssetsSuccessResponse struct {
	Asset Asset `json:"asset,omitempty"`
}

func (r *UploadAssetsSuccessResponse) StatusHTTP() int {
	return http.StatusCreated
}
