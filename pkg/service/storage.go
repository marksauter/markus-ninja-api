package service

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
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
	userId *oid.OID,
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
	contentType := mime.TypeByExtension(filepath.Ext(objectName))
	objectPath := strings.Join([]string{contentType, userId.Short, objectName}, "/")

	mylog.Log.Info("found object")

	return s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
}

func (s *StorageService) Upload(
	userId *oid.OID,
	file multipart.File,
	header *multipart.FileHeader,
) (key string, err error) {
	mylog.Log.WithField("name", header.Filename).Info("StorageService.Upload()")
	// Hash of the file contents to be used as the s3 object 'key'.
	hash := sha1.New()
	io.Copy(hash, file)
	hashHex := fmt.Sprintf("%x", hash.Sum(nil))

	contentType := header.Header.Get("Content-Type")
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf(
		"%s/%s/%s/%s%s",
		hashHex[:2],
		hashHex[3:5],
		hashHex[6:8],
		hashHex[9:],
		ext,
	)
	objectPath := strings.Join([]string{
		contentType,
		userId.Short,
		objectName,
	}, "/")

	n, err := s.svc.PutObject(
		s.bucket,
		objectPath,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to put object")
		return
	}

	mylog.Log.WithField("size", n).Info("Successfully uploaded file")

	key = hashHex + ext
	return
}
