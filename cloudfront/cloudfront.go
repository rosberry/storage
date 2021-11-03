package cloudfront

import (
	"crypto/x509"
	"encoding/pem"
	"net/url"

	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/rosberry/storage"
)

type (
	Config struct {
		StorageKey   string
		DomainName   string
		CFPrefix     string
		NoSSL        bool
		SignURLs     bool
		StorageCtl   storage.Storage
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
	return c.cfg.StorageCtl.Store(filePath, path)
}

func (c *CFStorage) GetURL(cLink string, options ...interface{}) (URL string) {
	u, err := url.Parse(cLink)
	if err != nil || u.Scheme != c.cfg.StorageKey {
		return
	}
	u.Scheme = c.scheme
	u.Host = c.cfg.DomainName
	u.Path = c.cfg.CFPrefix + u.Path

	URL = u.String()

	if c.cfg.SignURLs {
		signed := false
		for _, op := range options {
			if expirationVerifier, ok := op.(storage.ExpirationVerifier); ok {
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
		if signed == false {
			URL = ""
		}
	}

	return
}

func (c *CFStorage) Remove(cLink string) (err error) {
	return c.cfg.StorageCtl.Remove(cLink)
}

func (c *CFStorage) StoreByCLink(filePath, cLink string) (err error) {
	return nil
}