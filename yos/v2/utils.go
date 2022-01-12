package yos

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/rosberry/storage/common"
)

func (y *YandexObjStorage) prepareURL(cLink string, options ...interface{}) string {
	path := common.CLinkToPath(y.cfg.StorageKey, cLink)
	internalPath := common.PathToInternalPath(y.cfg.Prefix, path)

	for _, o := range options {
		if option, ok := o.(YandexOption); ok {
			switch option {
			case PublicLink:
				return y.preparePublicURL(internalPath)
			case LinkForPutObject:
				return y.preparePutObjectURL(internalPath)
			}
		}
	}

	// check exist object
	headURL, err := y.client.PresignedHeadObject(
		context.Background(),
		y.cfg.BucketName,
		internalPath,
		checkExistLinkLifeTime,
		nil)
	if err != nil {
		log.Printf("Failed generate presignedURL for check exist object: %v", err)
	}

	if !checkExistObject(headURL.String()) {
		return ""
	}

	presignedURL, err := y.client.PresignedGetObject(
		context.Background(),
		y.cfg.BucketName,
		internalPath,
		getObjectLinkLifeTime,
		nil)
	if err != nil {
		log.Printf("Failed generate presignedURL: %v", err)
	}

	return presignedURL.String()
}

func (y *YandexObjStorage) preparePublicURL(internalPath string) string {
	u := &url.URL{
		Scheme: SchemeHTTPWithSSL,
		Host:   y.endpoint,
		Path:   y.cfg.BucketName + "/" + internalPath,
	}

	return u.String()
}

func (y *YandexObjStorage) preparePutObjectURL(internalPath string) string {
	presignedURL, err := y.client.PresignedPutObject(
		context.Background(),
		y.cfg.BucketName,
		internalPath,
		putLinkLifeTime,
	)
	if err != nil {
		log.Print("Failed generate presignedURL: ", err)
		return ""
	}

	return presignedURL.String()
}

func checkExistObject(headURL string) (exist bool) {
	req, err := http.NewRequestWithContext(context.TODO(), "HEAD", headURL, nil)
	if err != nil {
		log.Print(err)
	}

	// send request with headers
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Print("failed do request: v", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
