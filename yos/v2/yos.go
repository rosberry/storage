package yos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	YandexObjStorage struct {
		cfg      Config
		scheme   string
		endpoint string

		client *minio.Client
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
	putLinkLifeTime        = 30 * time.Minute
	checkExistLinkLifeTime = 10 * time.Minute
	getObjectLinkLifeTime  = 24 * time.Hour
)

const (
	SchemeHTTPWithSSL    = "https"
	SchemeHTTPWithoutSSL = "http"

	endpoint = "storage.yandexcloud.net"
)

var (
	ErrStorageKeyNotMatch   = errors.New("storage Key did not match")
	ErrMethodNotImplemented = errors.New("method is not implemented")
)

func New(cfg *Config) *YandexObjStorage {
	scheme := SchemeHTTPWithSSL

	if cfg.NoSSL {
		scheme = SchemeHTTPWithoutSSL
	}

	y := &YandexObjStorage{
		cfg:      *cfg,
		scheme:   scheme,
		endpoint: endpoint,
	}

	minioClient, err := minio.New(y.endpoint, &minio.Options{
		Region: y.cfg.Region,
		Creds:  credentials.NewStaticV4(y.cfg.AccessKeyID, y.cfg.SecretAccessKey, ""),
		Secure: !y.cfg.NoSSL,
	})
	if err != nil {
		log.Printf("failed init minio client: %v", err)
	}

	y.client = minioClient

	return y
}

func (y *YandexObjStorage) Store(filePath, path string) (cLink string, err error) {
	f, _ := os.Open(filePath)
	defer f.Close()

	mimetype := common.GetFileContentType(f)
	internalPath := common.PathToInternalPath(y.cfg.Prefix, path)

	_, err = y.client.FPutObject(
		context.Background(),
		y.cfg.BucketName,
		internalPath,
		filePath,
		minio.PutObjectOptions{ContentType: mimetype})
	if err != nil {
		log.Print(err)
	}

	cLink = common.PathToCLink(y.cfg.StorageKey, path)

	return
}

func (y *YandexObjStorage) GetURL(cLink string, options ...interface{}) string {
	if strings.Contains(cLink, "http") {
		return cLink
	}

	return y.prepareURL(cLink, options...)
}

func (y *YandexObjStorage) Remove(cLink string) (err error) {
	return ErrMethodNotImplemented
}

func (y *YandexObjStorage) GetCLink(path string) (cLink string) {
	return common.PathToCLink(y.cfg.StorageKey, path)
}

func (y *YandexObjStorage) StoreByCLink(filePath, cLink string) (err error) {
	path := common.CLinkToPath(y.cfg.StorageKey, cLink)

	_, err = y.Store(filePath, path)
	if err != nil {
		return fmt.Errorf("failed store file: %w", err)
	}

	return nil
}
