package service

import (
	"crypto/sha1"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	// Allow processing of images
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	minio "github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
)

// StorageService - service used for storing assets
type StorageService struct {
	bucket string
	svc    *minio.Client
}

// NewStorageService - create a new storage service instance
func NewStorageService() (*StorageService, error) {
	credentials := credentials.NewFileAWSCredentials("", "")
	endPoint := "s3.amazonaws.com"
	useSSL := true
	svc, err := minio.NewWithCredentials(
		endPoint,
		credentials,
		useSSL,
		myaws.AWSRegion,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to new minio client")
		return nil, err
	}
	// svc.TraceOn(nil)
	return &StorageService{
		bucket: "markus-ninja-development-user-asset-us-east-1",
		svc:    svc,
	}, nil
}

// Get - get an object from passed contentType, userID, and key
func (s *StorageService) Get(
	contentType string,
	userID *mytype.OID,
	key string,
) (*minio.Object, error) {
	mylog.Log.WithField("key", key).Info("StorageService.Get()")

	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		key[:2],
		key[3:5],
		key[6:8],
		key[9:],
	)
	objectPath := strings.Join(
		[]string{
			contentType,
			userID.Short,
			objectName,
		},
		"/",
	)

	return s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
}

// GetThumbnail - get a thumbnail of an asset, and generate it first if
// necessary
func (s *StorageService) GetThumbnail(
	size int,
	contentType string,
	userID *mytype.OID,
	key string,
) (*minio.Object, error) {
	mylog.Log.WithField("key", key).Info("StorageService.GetThumbnail()")

	sizeStr := strconv.FormatInt(int64(size), 10)
	// Thumbnail objects are identified with a -'size' at the end of the key
	thumbKey := key + "--" + sizeStr
	thumbLocal := "./tmp/" + thumbKey + ".jpg"

	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		thumbKey[:2],
		thumbKey[3:5],
		thumbKey[6:8],
		thumbKey[9:],
	)
	objectPath := strings.Join([]string{
		contentType,
		userID.Short,
		objectName,
	}, "/")

	object, err := s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
	if err != nil {
		minioError := minio.ToErrorResponse(err)
		if minioError.Code != "NoSuchKey" {
			return nil, err
		}

		asset, err := s.Get(contentType, userID, key)
		if err != nil {
			return nil, err
		}

		img, err := imaging.Decode(asset)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to decode image")
			return nil, err
		}
		thumb := imaging.Thumbnail(img, size, size, imaging.CatmullRom)

		// create a new blank image
		dst := imaging.New(size, size, color.NRGBA{0, 0, 0, 0})

		// paste thumbnails into the new image
		dst = imaging.Paste(dst, thumb, image.Pt(0, 0))

		// ensure path is available
		dir := filepath.Dir(thumbLocal)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			mylog.Log.WithError(err).WithField("path", dir).Error("failed to create tmp directory")
			return nil, err
		}

		err = imaging.Save(dst, thumbLocal)
		if err != nil {
			mylog.Log.WithError(err).WithField("thumbnail", thumbLocal).Error("failed to generate thumbnail")
			return nil, err
		}

		thumbFile, err := os.Open(thumbLocal)
		if err != nil {
			mylog.Log.WithError(err).WithField("thumbnail", thumbLocal).Error("failed to open file")
			return nil, err
		}

		thumbStat, err := thumbFile.Stat()
		if err != nil {
			return nil, err
		}

		_, err = s.svc.PutObject(
			s.bucket,
			objectPath,
			thumbFile,
			thumbStat.Size(),
			minio.PutObjectOptions{ContentType: contentType},
		)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to put object")
			return nil, err
		}
	}

	return object, nil
}

// UploadResponse - response object from Upload
type UploadResponse struct {
	Key         string
	IsNewObject bool
}

// Upload - upload asset to storage service
func (s *StorageService) Upload(
	userID *mytype.OID,
	file io.Reader,
	contentType string,
	size int64,
) (*UploadResponse, error) {
	mylog.Log.Info("StorageService.Upload()")
	// Hash of the file contents to be used as the s3 object 'key'.
	hash := sha1.New()
	io.Copy(hash, file)
	key := fmt.Sprintf("%x", hash.Sum(nil))

	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		key[:2],
		key[3:5],
		key[6:8],
		key[9:],
	)
	objectPath := strings.Join([]string{
		contentType,
		userID.Short,
		objectName,
	}, "/")

	_, err := s.svc.StatObject(
		s.bucket,
		objectPath,
		minio.StatObjectOptions{},
	)
	if err != nil {
		minioError := minio.ToErrorResponse(err)
		if minioError.Code != "NoSuchKey" {
			return nil, err
		}
		n, err := s.svc.PutObject(
			s.bucket,
			objectPath,
			file,
			size,
			minio.PutObjectOptions{ContentType: contentType},
		)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to put object")
			return nil, err
		}

		mylog.Log.WithField("size", n).Info("Successfully uploaded file")

		return &UploadResponse{
			Key:         key,
			IsNewObject: true,
		}, nil
	}

	return &UploadResponse{
		Key:         key,
		IsNewObject: false,
	}, nil
}
