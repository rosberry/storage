package yos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	cm "github.com/rosberry/storage/common"
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

	YandexObjStorage struct {
		cfg      Config
		scheme   string
		endpoint string
	}
)

type (
	YandexOption int
)

const (
	PublicLink YandexOption = iota + 1
	LinkForPutObject
)

const (
	SchemeHTTPWithSSL    = "https"
	SchemeHTTPWithoutSSL = "http"

	endpoint = "storage.yandexcloud.net"
)

var (
	ErrStorageKeyNotMatch   = errors.New("Storage Key did not match!")
	ErrMethodNotImplemented = errors.New("Method is not implemented")
)

var Instance = &YandexObjStorage{}

func New(cfg *Config) *YandexObjStorage {
	scheme := SchemeHTTPWithSSL
	if cfg.NoSSL {
		scheme = SchemeHTTPWithoutSSL
	}
	return &YandexObjStorage{
		cfg:      *cfg,
		scheme:   scheme,
		endpoint: endpoint,
	}
}

func (y *YandexObjStorage) Store(filePath, path string) (cLink string, err error) {
	// Initialize minio client object.
	minioClient, err := minio.New(y.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(y.cfg.AccessKeyID, y.cfg.SecretAccessKey, ""),
		Secure: !y.cfg.NoSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	f, _ := os.Open(filePath)
	defer f.Close()

	mimetype := cm.GetFileContentType(f)
	internalPath := path

	_, err = minioClient.FPutObject(context.Background(), y.cfg.BucketName, internalPath, filePath, minio.PutObjectOptions{ContentType: mimetype})
	if err != nil {
		log.Println(err)
	}

	cLink = fmt.Sprintf("%s:%s", y.cfg.StorageKey, internalPath)
	return
}

func (y *YandexObjStorage) GetURL(cLink string, options ...interface{}) (URL string) {
	if strings.Index(cLink, "http") > -1 {
		return cLink
	}

	obj := strings.Replace(cLink, y.cfg.StorageKey+":", "", 1)

	s3Client, err := minio.New(y.endpoint, &minio.Options{
		Region: y.cfg.Region,
		Creds:  credentials.NewStaticV4(y.cfg.AccessKeyID, y.cfg.SecretAccessKey, ""),
		Secure: !y.cfg.NoSSL,
	})
	if err != nil {
		log.Println("Failed create new minio client:", err)
		return ""
	}

	for _, o := range options {
		if option, ok := o.(YandexOption); ok {
			switch option {
			case PublicLink:
				downloadLink := fmt.Sprintf("https://%v/%v/%v", y.endpoint, y.cfg.BucketName, obj)
				return downloadLink
			case LinkForPutObject:
				presignedURL, err := s3Client.PresignedPutObject(context.Background(), y.cfg.BucketName, obj, time.Duration(30)*time.Minute)
				if err != nil {
					log.Println("Failed generate presignedURL: ", err)
					return ""
				}
				return presignedURL.String()
			}
		}
	}

	//check exist object
	headUrl, err := s3Client.PresignedHeadObject(context.Background(), y.cfg.BucketName, obj, time.Duration(600)*time.Second, nil)
	if !checkExistObject(headUrl.String()) {
		return ""
	}

	presignedURL, err := s3Client.PresignedGetObject(context.Background(), y.cfg.BucketName, obj, time.Duration(24)*time.Hour, nil)
	if err != nil {
		log.Println("Failed generate presignedURL: ", err)
	}

	return presignedURL.String()
}

func (b *YandexObjStorage) Remove(cLink string) (err error) {
	return ErrMethodNotImplemented
}

func checkExistObject(headUrl string) (exist bool) {
	req, err := http.NewRequest("HEAD", headUrl, nil)
	if err != nil {
		log.Println(err)
	}

	// send request with headers
	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
