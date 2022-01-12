package s3

import (
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rosberry/storage/common"
)

func (s *S3Storage) getSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(s.cfg.Region),
		Credentials: credentials.NewStaticCredentials(s.cfg.AccessKeyID, s.cfg.SecretAccessKey, ""),
	}))
}

func (s *S3Storage) storeByPath(filePath string, path string) (cLink string, err error) {
	err = s.storeByInternalPath(filePath, common.PathToInternalPath(s.cfg.Prefix, path))
	if err != nil {
		return "", fmt.Errorf("failed to store %s: %w", path, err)
	}

	cLink = s.GetCLink(path)

	return
}

func (s *S3Storage) storeByInternalPath(filePath, internalPath string) (err error) {
	uploader := s3manager.NewUploader(s.getSession())

	f, _ := os.Open(filePath)
	defer f.Close()

	mimetype := common.GetFileContentType(f)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.cfg.BucketName),
		Key:         aws.String(internalPath),
		Body:        f,
		ContentType: aws.String(mimetype),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *S3Storage) prepareURL(cLink string) (u *url.URL, err error) {
	var uc *url.URL

	uc, err = url.Parse(cLink)
	if err != nil {
		return
	}

	if uc.Scheme != s.cfg.StorageKey {
		err = ErrStorageKeyNotMatch
		return
	}

	u = &url.URL{}

	u.Scheme = s.scheme
	u.Host = fmt.Sprintf(S3HostTemplate, s.cfg.BucketName)

	u.Path = uc.Path
	if uc.Opaque != "" {
		u.Path = "/" + uc.Opaque
	}

	u.Path = s.cfg.Prefix + u.Path

	return
}
