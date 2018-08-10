package service

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	minio "github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
)

type StorageService struct {
	bucket string
	svc    *minio.Client
}

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
	return &StorageService{
		bucket: "markus-ninja-development-user-asset-us-east-1",
		svc:    svc,
	}, nil
}

func (s *StorageService) Get(
	contentType string,
	userId *mytype.OID,
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
			userId.Short,
			objectName,
		},
		"/",
	)

	mylog.Log.Info("found object")

	return s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
}

type UploadResponse struct {
	Key         string
	IsNewObject bool
}

func (s *StorageService) Upload(
	userId *mytype.OID,
	file multipart.File,
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
		userId.Short,
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
		} else {
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
	}

	return &UploadResponse{
		Key:         key,
		IsNewObject: false,
	}, nil
}
