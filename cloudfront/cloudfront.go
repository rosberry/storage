package cloudfront

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/rosberry/storage/common"
	"github.com/rosberry/storage/core"
)

type (
	Config struct {
		StorageKey   string
		DomainName   string
		CFPrefix     string
		NoSSL        bool
		SignURLs     bool
		StorageCtl   core.Storage
		PrivateKeyID string
		PrivateKey   string
	}

	CFStorage struct {
		cfg    Config
		scheme string
	}
)

const (
	SchemeHTTPWithSSL    = "https"
	SchemeHTTPWithoutSSL = "http"
)

func New(cfg *Config) *CFStorage {
	scheme := SchemeHTTPWithSSL

	if cfg.NoSSL {
		scheme = SchemeHTTPWithoutSSL
	}

	return &CFStorage{
		cfg:    *cfg,
		scheme: scheme,
	}
}

func (c *CFStorage) GetCLink(path string) (cLink string) {
	return c.cfg.StorageCtl.GetCLink(path)
}

func (c *CFStorage) Store(filePath, path string) (cLink string, err error) {
	cLink, err = c.cfg.StorageCtl.Store(filePath, path)
	if err != nil {
		return "", fmt.Errorf("failed to store %s: %w", path, err)
	}

	return
}

func (c *CFStorage) GetURL(cLink string, options ...interface{}) string {
	uc, err := url.Parse(cLink)
	if err != nil || uc.Scheme != c.cfg.StorageKey {
		return ""
	}

	u := &url.URL{}

	u.Scheme = c.scheme
	u.Host = c.cfg.DomainName

	u.Path = uc.Path
	if uc.Opaque != "" {
		u.Path = "/" + uc.Opaque
	}

	u.Path = c.cfg.CFPrefix + u.Path

	URL := u.String()

	if !c.cfg.SignURLs {
		return URL
	}

	var signed bool

	for _, op := range options {
		if expirationVerifier, ok := op.(core.ExpirationVerifier); ok {
			expire := expirationVerifier.GetAccessExpireTime(URL)
			block, _ := pem.Decode([]byte(c.cfg.PrivateKey))

			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				signer := sign.NewURLSigner(c.cfg.PrivateKeyID, privateKey)

				URL, err = signer.Sign(URL, expire)
				if err == nil {
					signed = true
				}
			}
		}
	}

	if !signed {
		URL = ""
	}

	return URL
}

func (c *CFStorage) Remove(cLink string) (err error) {
	return c.cfg.StorageCtl.Remove(cLink) // nolint:wrapcheck
}

func (c *CFStorage) StoreByCLink(filePath, cLink string) (err error) {
	path := common.CLinkToPath(c.cfg.StorageKey, cLink)

	_, err = c.cfg.StorageCtl.Store(filePath, path)
	if err != nil {
		return fmt.Errorf("failed to store %s: %w", path, err)
	}

	return
}
