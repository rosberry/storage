package local

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
)

type (
	Config struct {
		StorageKey string
		Endpoint   string
		Root       string
		BufferSize int // bytes
	}

	Local struct {
		cfg Config
	}
)

const (
	defaultStorageKey = "file"
	defaultEndpoint   = "http://localhost:8080/"
	defaultRoot       = ""
	defaultBufferSize = 16 * 1024
)

var (
	ErrMethodNotImplemented = errors.New("method is not implemented")
	ErrFailedGetFilePath    = errors.New("failed to get file path")
)

func New(cfg *Config) *Local {
	if cfg == nil {
		cfg = &Config{
			StorageKey: defaultStorageKey,
			Endpoint:   defaultEndpoint,
			Root:       defaultRoot,
			BufferSize: defaultBufferSize,
		}
	}

	return &Local{
		cfg: *cfg,
	}
}

func (b *Local) Store(filePath, path string) (cLink string, err error) {
	return b.storeByPath(filePath, path)
}

func (b *Local) StoreByCLink(filePath, cLink string) (err error) {
	path := b.cLinkToPath(cLink)

	_, err = b.storeByPath(filePath, path)

	return
}

func (b *Local) GetURL(cLink string, options ...interface{}) string {
	if !checkStorageKey(cLink, b.cfg.StorageKey) {
		log.Println("Failed check storage key:", cLink, b.cfg.StorageKey)
		return ""
	}

	u, err := url.Parse(endSlash(b.cfg.Endpoint) + b.cLinkToPath(cLink))
	if err != nil {
		log.Println("Parse err:", err)
		return ""
	}

	return u.String()
}

func (b *Local) Remove(cLink string) (err error) {
	path := b.cLinkToPath(cLink)
	if path == "" {
		return ErrFailedGetFilePath
	}

	internalPath := b.pathToInternalPath(path)
	if internalPath == "" {
		return ErrFailedGetFilePath
	}

	err = os.Remove(internalPath)
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

func (b *Local) GetCLink(path string) (cLink string) {
	return b.pathToCLink(path)
}
