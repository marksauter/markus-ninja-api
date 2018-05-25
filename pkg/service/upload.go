package service

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	minio "github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type UploadService struct {
	bucket string
	svc    *minio.Client
}

func NewUploadService() (*UploadService, error) {
	credentials := credentials.NewFileAWSCredentials("", "")
	useSSL := true
	svc, err := minio.NewWithCredentials(
		"s3.amazonaws.com",
		credentials,
		useSSL,
		myaws.AWSRegion,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to new minio client")
		return nil, err
	}
	return &UploadService{
		bucket: "markus-ninja-development-user-asset-us-east-1",
		svc:    svc,
	}, nil
}

func (s *UploadService) Upload(
	file multipart.File,
	header *multipart.FileHeader,
) error {
	// Hash of the file contents to be used as the s3 object 'key'.
	hash := sha1.New()
	io.Copy(hash, file)
	hashHex := fmt.Sprintf("%x", hash.Sum(nil))

	objectSize := header.Size
	contentType := header.Header.Get("Content-Type")
	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		hashHex[:2],
		hashHex[3:5],
		hashHex[6:8],
		hashHex[9:],
	)
	objectPath := strings.Join([]string{contentType, objectName}, "/")

	n, err := s.svc.PutObject(
		s.bucket,
		objectPath,
		file,
		objectSize,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to upload file")
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"filename": header.Filename,
		"size":     n,
	}).Info("Successfully uploaded file")
	return nil
}
