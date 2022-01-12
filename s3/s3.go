package s3

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rosberry/storage/common"
)

type (
	Config struct {
		StorageKey      string
		Region          string
		AccessKeyID     string
		SecretAccessKey string
		BucketName      string
		Prefix          string
		NoSSL           bool
	}

	S3Storage struct { // nolint:golint
		cfg    Config
		scheme string
	}
)

const (
	SchemeHTTPWithSSL    = "https"
	SchemeHTTPWithoutSSL = "http"

	S3HostTemplate = "%s.s3.amazonaws.com"
)

var (
	ErrStorageKeyNotMatch = errors.New("storage Key did not match")
	ErrFailedGetFilePath  = errors.New("failed to get file path")
)

func New(cfg *Config) *S3Storage {
	scheme := SchemeHTTPWithSSL
	if cfg.NoSSL {
		scheme = SchemeHTTPWithoutSSL
	}

	return &S3Storage{
		cfg:    *cfg,
		scheme: scheme,
	}
}

func (s *S3Storage) Store(filePath, path string) (cLink string, err error) {
	return s.storeByPath(filePath, path)
}

func (s *S3Storage) StoreByCLink(filePath, cLink string) (err error) {
	path := common.CLinkToPath(s.cfg.StorageKey, cLink)

	_, err = s.storeByPath(filePath, path)

	return err
}

func (s *S3Storage) GetURL(cLink string, options ...interface{}) string {
	u, err := s.prepareURL(cLink)
	if err != nil {
		return ""
	}

	return u.String()
}

func (s *S3Storage) Remove(cLink string) (err error) {
	path := common.CLinkToPath(s.cfg.StorageKey, cLink)
	if path == "" {
		return fmt.Errorf("%s: %w", cLink, ErrFailedGetFilePath)
	}

	internalPath := common.PathToInternalPath(s.cfg.Prefix, path)
	if internalPath == "" {
		return fmt.Errorf("%s: %s: %w", cLink, path, ErrFailedGetFilePath)
	}

	svc := s3.New(s.getSession())
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(internalPath),
	})

	return
}

func (s *S3Storage) GetCLink(path string) (cLink string) {
	return common.PathToCLink(s.cfg.StorageKey, path)
}
