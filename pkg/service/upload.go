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

// const (
//   maxPartSize = int64(512 * 1000)
//   maxRetries  = 3
// )

type UploadService struct {
	bucket string
	// svc    s3iface.S3API
	svc *minio.Client
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
		// svc:    myaws.NewS3(),
		svc: svc,
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

// func NewMockUploadService() *UploadService {
//   return &UploadService{
//     bucket: "markus-ninja-development-user-asset-us-east-1",
//     svc:    myaws.NewMockS3(),
//   }
// }

// func (s *UploadService) MultipartUpload(file multipart.File, header *multipart.FileHeader) (int64, error) {
//   // Total data read and writtern to server.
//   // Should be equal to 'size' at the end of the call
//   var totalUploadedSize int64
//
//   totalPartsCount, partSize, _, err := optimalPartInfo(-1)
//   if err != nil {
//     return 0, err
//   }
//
//   // Hash of the file contents to be used as the s3 object 'key'.
//   hash := sha1.New()
//   io.Copy(hash, file)
//   hashHex := fmt.Sprintf("%x", hash.Sum(nil))
//
//   size := header.Size
//   filetype := header.Header.Get("Content-Type")
//   filename := fmt.Sprintf(
//     "%s/%s/%s/%s",
//     hashHex[:2],
//     hashHex[3:5],
//     hashHex[6:8],
//     hashHex[9:],
//   )
//   path := strings.Join([]string{filetype, filename}, "/")
//
//   input := &s3.CreateMultipartUploadInput{
//     Bucket:      aws.String(s.bucket),
//     Key:         aws.String(path),
//     ContentType: aws.String(filetype),
//   }
//   output, err := s.svc.CreateMultipartUpload(input)
//   if err != nil {
//     mylog.Log.WithError(err).Error("failed to create multipart upload")
//     return err
//   }
//
//   var curr, partLength int64
//   var remaining = size
//   var completedParts []*s3.CompletedPart
//   partNumber := int64(1)
//   for curr = 0; remaining != 0; curr += partLength {
//     if remaining < maxPartSize {
//       partLength = remaining
//     } else {
//       partLength = maxPartSize
//     }
//     completedPart, err := s.uploadPart(output, buffer[curr:curr+partLength], partNumber)
//     if err != nil {
//       mylog.Log.WithError(err).Error("failed to upload part")
//       if err := s.abortMultipartUpload(output); err != nil {
//         mylog.Log.WithError(err).Error("failed to abort multipart upload")
//         return err
//       }
//       return err
//     }
//     remaining -= partLength
//     partNumber++
//     completedParts = append(completedParts, completedPart)
//   }
//
//   response, err := s.completeMultipartUpload(output, completedParts)
//   if err != nil {
//     mylog.Log.WithError(err).Error("failed to complete multipart upload")
//     return err
//   }
//
//   mylog.Log.WithField("response", response.String()).Info("Successfully uploaded file")
//
//   return nil
// }
//
// func (s *UploadService) abortMultipartUpload(
//   o *s3.CreateMultipartUploadOutput,
// ) error {
//   mylog.Log.WithField("upload_id", o.UploadId).Info("Aborting multipart upload")
//   input := &s3.AbortMultipartUploadInput{
//     Bucket:   o.Bucket,
//     Key:      o.Key,
//     UploadId: o.UploadId,
//   }
//   _, err := s.svc.AbortMultipartUpload(input)
//   return err
// }
//
// func (s *UploadService) completeMultipartUpload(
//   o *s3.CreateMultipartUploadOutput,
//   completedParts []*s3.CompletedPart,
// ) (*s3.CompleteMultipartUploadOutput, error) {
//   input := &s3.CompleteMultipartUploadInput{
//     Bucket:   o.Bucket,
//     Key:      o.Key,
//     UploadId: o.UploadId,
//     MultipartUpload: &s3.CompletedMultipartUpload{
//       Parts: completedParts,
//     },
//   }
//   return s.svc.CompleteMultipartUpload(input)
// }
//
// func (s *UploadService) uploadPart(
//   o *s3.CreateMultipartUploadOutput,
//   fileBytes []byte,
//   partNumber int64,
// ) (*s3.CompletedPart, error) {
//   tryNum := 1
//   input := &s3.UploadPartInput{
//     Body:          bytes.NewReader(fileBytes),
//     Bucket:        o.Bucket,
//     Key:           o.Key,
//     PartNumber:    aws.Int64(partNumber),
//     UploadId:      o.UploadId,
//     ContentLength: aws.Int64(int64(len(fileBytes))),
//   }
//
//   for tryNum <= maxRetries {
//     output, err := s.svc.UploadPart(input)
//     if err != nil {
//       if tryNum == maxRetries {
//         if aerr, ok := err.(awserr.Error); ok {
//           return nil, aerr
//         }
//         return nil, err
//       }
//       mylog.Log.WithField("n", partNumber).Info("Retrying to upload part...")
//       tryNum++
//     } else {
//       mylog.Log.WithField("n", partNumber).Info("Uploaded part")
//       return &s3.CompletedPart{
//         ETag:       output.ETag,
//         PartNumber: aws.Int64(partNumber),
//       }, nil
//     }
//   }
//
//   return nil, nil
// }
