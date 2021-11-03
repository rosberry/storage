package s3

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	cm "github.com/rosberry/storage/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/nu7hatch/gouuid"
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

	S3Storage struct {
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
	ErrStorageKeyNotMatch = errors.New("Storage Key did not match!")
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

func (s *S3Storage) getSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(s.cfg.Region),
		Credentials: credentials.NewStaticCredentials(s.cfg.AccessKeyID, s.cfg.SecretAccessKey, ""),
	}))
}

func (s *S3Storage) internalPath(path string) string {
	u4, _ := uuid.NewV4()
	internalPath := "/" + path + u4.String()

	return internalPath
}

func (s *S3Storage) GetCLink(path string) (cLink string) {
	return s.getCLinkByInternalPath(s.internalPath(path))
}

func (s *S3Storage) getCLinkByInternalPath(internalPath string) (cLink string) {
	return fmt.Sprintf("%s:%s", s.cfg.StorageKey, internalPath)
}

func (s *S3Storage) getInternalPathByCLink(cLink string) (internalPath string) {
	internalPath = strings.TrimPrefix(cLink, s.cfg.StorageKey+":")

	log.Printf("internal path by clink '%s' = '%s'", cLink, internalPath)

	return internalPath
}

func (s *S3Storage) Store(filePath, path string) (cLink string, err error) {
	return s.storeByInternalPath(filePath, s.internalPath(path))
}

func (s *S3Storage) StoreByCLink(filePath, cLink string) (err error) {
	_, err = s.storeByInternalPath(filePath, s.getInternalPathByCLink(cLink))

	return err
}

func (s *S3Storage) storeByInternalPath(filePath, internalPath string) (cLink string, err error) {
	uploader := s3manager.NewUploader(s.getSession())

	f, _ := os.Open(filePath)
	defer f.Close()

	mimetype := cm.GetFileContentType(f)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.cfg.BucketName),
		Key:         aws.String(s.cfg.Prefix + internalPath),
		Body:        f,
		ContentType: aws.String(mimetype),
	})

	cLink = s.getCLinkByInternalPath(internalPath)
	return
}


func (s *S3Storage) prepareURL(cLink string) (u *url.URL, err error) {
	u, err = url.Parse(cLink)
	if err != nil {
		return
	}
	if u.Scheme != s.cfg.StorageKey {
		err = ErrStorageKeyNotMatch
		return
	}
	u.Scheme = s.scheme
	u.Host = fmt.Sprintf(S3HostTemplate, s.cfg.BucketName)
	u.Path = s.cfg.Prefix + u.Path
	return

}

func (s *S3Storage) GetURL(cLink string, options ...interface{}) (URL string) {
	u, err := s.prepareURL(cLink)
	if err != nil {
		return ""
	}
	return u.String()
}

func (s *S3Storage) Remove(cLink string) (err error) {
	u, e := s.prepareURL(cLink)
	if e != nil {
		return e
	}

	svc := s3.New(s.getSession())
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(u.Path),
	})
	return
}
