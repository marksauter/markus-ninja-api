package route

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

var UploadAssetsCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Authorization", "Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "http://localhost:*"},
})

type UploadAssetsHandler struct {
	Repos      *repo.Repos
	StorageSvc *service.StorageService
}

func (h UploadAssetsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
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

	viewer, ok := myctx.UserFromContext(req.Context())
	if !ok {
		mylog.Log.WithError(err).Error("failed to get user from context")
		response := myhttp.AccessDeniedErrorResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}
	key, err := h.StorageSvc.Upload(&viewer.Id, file, header)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to upload file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	asset, err := data.NewAssetFromFile(&viewer.Id, key, file, header)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create asset from file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	assetPermit, err := h.Repos.Asset().Create(req.Context(), asset)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create asset")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	saveStr := req.FormValue("save")
	save, err := strconv.ParseBool(saveStr)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to parse form `save`")
		response := myhttp.InvalidRequestErrorResponse("invalid study_id")
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if save {
		formStudyId := req.FormValue("study_id")
		studyId, err := mytype.ParseOID(formStudyId)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to parse form `study_id")
			response := myhttp.InvalidRequestErrorResponse("invalid study_id")
			myhttp.WriteResponseTo(rw, response)
			return
		}

		userAsset, err := data.NewUserAsset(&viewer.Id, studyId, &asset.Id, header.Filename)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to create user asset")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		_, err = h.Repos.UserAsset().Create(req.Context(), userAsset)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to create user asset")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

	}

	href, err := assetPermit.Href()
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get asset href")
		response := myhttp.AccessDeniedErrorResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}

	assetResponse := Asset{
		ContentType: contentType,
		Href:        href,
		Id:          asset.Id.Short,
		Name:        header.Filename,
		Size:        header.Size,
	}

	response := &UploadAssetsSuccessResponse{Asset: assetResponse}
	myhttp.WriteResponseTo(rw, response)
	return
}

type Asset struct {
	ContentType string `json:"content_type,omitempty"`
	Href        string `json:"href,omitempty"`
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

type UploadAssetsSuccessResponse struct {
	Asset Asset `json:"asset,omitempty"`
}

func (r *UploadAssetsSuccessResponse) StatusHTTP() int {
	return http.StatusCreated
}
