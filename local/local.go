package local

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

var (
	defaultStorageKey = "file"
	defaultEndpoint   = "http://localhost:8080/"
	defaultRoot       = ""
	defaultBufferSize = 16 * 1024
)

var ErrMethodNotImplemented = errors.New("Method is not implemented")

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

func (b *Local) internalPath(path string) (internalPath string) {
	return endSlash(b.cfg.Root)+strings.TrimLeft(path, "/")
}

func (b *Local) Store(filePath, path string) (cLink string, err error) {
	return b.storeByInternalPath(filePath, b.internalPath(path))
}

func (b *Local) StoreByCLink(filePath, cLink string) (err error) {
	internalPath := b.cLinkToPath(cLink)

	_, err = b.storeByInternalPath(filePath, internalPath)

	return
}

func (b *Local) storeByInternalPath(filePath, internalPath string ) (cLink string, err error) {
	err = copy(filePath, internalPath, b.cfg.BufferSize)
	if err != nil {
		return "", err
	}

	cLink = b.internalPathToCLink(internalPath)
	return cLink, nil
}

func (b *Local) GetURL(cLink string, options ...interface{}) (URL string) {
	if !checkStorageKey(cLink, b.cfg.StorageKey) {
		log.Println("Failed check storage key:", cLink, b.cfg.StorageKey)
		return ""
	}

	u, err := url.Parse(endSlash(b.cfg.Endpoint) + strings.TrimPrefix(cLink, b.cfg.StorageKey+":"))
	if err != nil {
		log.Println("Parse err:", err)
		return ""
	}
	return u.String()
}

func (b *Local) Remove(cLink string) (err error) {
	path := b.cLinkToPath(cLink)
	err = os.Remove(path)

	return err
}

func copy(src, dst string, bufferSize int) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0770)
	if err != nil {
		return err
	}

	os.Remove(dst)

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}
func (b *Local) GetCLink(path string) (cLink string) {
	return b.pathToCLink(path)
}

func (b *Local) internalPathToCLink(internalPath string) (cLink string) {
	return b.pathToCLink(internalPath)
}

func (b *Local) pathToCLink(path string) (cLink string) {
	return fmt.Sprintf("%s:%s", b.cfg.StorageKey, strings.TrimLeft(path, "/"))
}

func (b *Local) cLinkToPath(cLink string) (path string) {
	if !checkStorageKey(cLink, b.cfg.StorageKey) {
		return ""
	}
	return endSlash(b.cfg.Root) + strings.TrimPrefix(cLink, b.cfg.StorageKey+":")
}

func checkStorageKey(cLink string, storageKey string) (ok bool) {
	return strings.Contains(cLink, storageKey+":")
}

func endSlash(s string) string {
	return strings.TrimRight(s, "/") + "/"
}
