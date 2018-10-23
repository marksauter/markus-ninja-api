package route

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

const (
	MB = 1 << 20
)

type Sizer interface {
	Size() int64
}

var UploadAssetsCors = cors.New(cors.Options{
	AllowCredentials: true,
	AllowedHeaders:   []string{"Authorization", "Content-Type"},
	AllowedMethods:   []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins:   []string{"ma.rkus.ninja", "http://localhost:*"},
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
	if err := req.ParseMultipartForm(10 * MB); err != nil {
		mylog.Log.WithError(err).Error("failed to parse multipart form")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	// Limit upload size
	req.Body = http.MaxBytesReader(rw, req.Body, 10*MB)
	file, multipartFileHeader, err := req.FormFile("file")
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get form file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	defer file.Close()

	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		mylog.Log.WithError(err).Error("failed to read file header")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	// Set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		mylog.Log.WithError(err).Error("failed to seek file to start")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	contentType := http.DetectContentType(fileHeader)
	isImage := strings.HasPrefix(contentType, "image")
	if !isImage {
		mylog.Log.WithField("type", contentType).Error("attempt to upload non-image file")
		response := myhttp.InvalidRequestErrorResponse("file is not an image")
		myhttp.WriteResponseTo(rw, response)
		return
	} else {
		types := strings.SplitN(contentType, "/", 2)
		subtype := types[1]
		if subtype != "png" && subtype != "jpeg" && subtype != "gif" {
			response := myhttp.InvalidRequestErrorResponse("image files must be of type 'png', 'jpeg', or 'gif'")
			myhttp.WriteResponseTo(rw, response)
			return
		}
	}

	fileSize := file.(Sizer).Size()
	if fileSize > 10*MB {
		response := myhttp.InvalidRequestErrorResponse("file size must not exceed 10 MB")
		myhttp.WriteResponseTo(rw, response)
		return
	}

	saveStr := req.FormValue("save")
	save, err := strconv.ParseBool(saveStr)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to parse form `save`")
		response := myhttp.InvalidRequestErrorResponse("invalid save")
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if save {
		img, _, err := image.DecodeConfig(bytes.NewReader(fileHeader))
		if err != nil {
			mylog.Log.WithError(err).Error("failed to decode image file")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
		if img.Width < 400 || img.Height < 400 {
			response := myhttp.InvalidRequestErrorResponse("saved images must be greater than 400 x 400")
			myhttp.WriteResponseTo(rw, response)
			return
		}
		if img.Width > 10000 || img.Height > 10000 {
			response := myhttp.InvalidRequestErrorResponse("saved images must be less than 10000 x 10000")
			myhttp.WriteResponseTo(rw, response)
			return
		}
	}

	viewer, ok := myctx.UserFromContext(req.Context())
	if !ok {
		mylog.Log.WithError(err).Error("failed to get user from context")
		response := myhttp.AccessDeniedErrorResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}
	uploadResponse, err := h.StorageSvc.Upload(&viewer.ID, file, contentType, fileSize)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to upload file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	filename := strings.Replace(multipartFileHeader.Filename, ` `, "_", -1)
	var assetPermit *repo.AssetPermit
	if uploadResponse.IsNewObject {
		asset, err := data.NewAssetFromFile(
			&viewer.ID,
			uploadResponse.Key,
			file,
			filename,
			contentType,
			fileSize,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to create asset from file")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		assetPermit, err = h.Repos.Asset().Create(req.Context(), asset)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to create asset")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
	} else {
		assetPermit, err = h.Repos.Asset().GetByKey(req.Context(), uploadResponse.Key)
		if err != nil && err != data.ErrNotFound {
			mylog.Log.WithError(err).Error("failed to get asset")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		} else if err == data.ErrNotFound {
			mylog.Log.Info("asset not found")
			asset, err := data.NewAssetFromFile(
				&viewer.ID,
				uploadResponse.Key,
				file,
				filename,
				contentType,
				fileSize,
			)
			if err != nil {
				mylog.Log.WithError(err).Error("failed to create asset from file")
				response := myhttp.InternalServerErrorResponse(err.Error())
				myhttp.WriteResponseTo(rw, response)
				return
			}

			assetPermit, err = h.Repos.Asset().Create(req.Context(), asset)
			if err != nil {
				mylog.Log.WithError(err).Error("failed to create asset")
				response := myhttp.InternalServerErrorResponse(err.Error())
				myhttp.WriteResponseTo(rw, response)
				return
			}
		}
	}

	assetID, err := assetPermit.ID()
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get asset id")
		response := myhttp.AccessDeniedErrorResponse()
		myhttp.WriteResponseTo(rw, response)
		return
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
		ID:          strconv.FormatInt(assetID, 10),
		Name:        filename,
		Size:        fileSize,
	}

	response := &UploadAssetsSuccessResponse{Asset: assetResponse}
	myhttp.WriteResponseTo(rw, response)
	return
}

type Asset struct {
	ContentType string `json:"content_type,omitempty"`
	Href        string `json:"href,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

type UploadAssetsSuccessResponse struct {
	Asset Asset `json:"asset,omitempty"`
}

func (r *UploadAssetsSuccessResponse) StatusHTTP() int {
	return http.StatusCreated
}
