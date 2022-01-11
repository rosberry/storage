package local

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rosberry/storage/common"
)

const mkdirPerm = 0o770

var ErrFileNotRegular = errors.New("not a regular file")

func (b *Local) pathToCLink(path string) (cLink string) {
	return common.PathToCLink(b.cfg.StorageKey, path)
}

func (b *Local) cLinkToPath(cLink string) (path string) {
	return common.CLinkToPath(b.cfg.StorageKey, cLink)
}

func (b *Local) pathToInternalPath(path string) (internalPath string) {
	return common.PathToInternalPath(b.cfg.Root, path)
}

func (b *Local) internalPathToPath(internalPath string) (path string) {
	return common.InternalPathToPath(b.cfg.Root, internalPath)
}

func (b *Local) storeByPath(filePath string, path string) (cLink string, err error) {
	return b.storeByInternalPath(filePath, b.pathToInternalPath(path))
}

func (b *Local) storeByInternalPath(filePath, internalPath string) (cLink string, err error) {
	err = copyFile(filePath, internalPath, b.cfg.BufferSize)
	if err != nil {
		return "", err
	}

	path := b.internalPathToPath(internalPath)
	cLink = b.pathToCLink(path)

	return cLink, nil
}

func checkStorageKey(cLink string, storageKey string) (ok bool) {
	return strings.Contains(cLink, storageKey+":")
}

func endSlash(s string) string {
	return strings.TrimRight(s, "/") + "/"
}

func copyFile(src, dst string, bufferSize int) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat: %w", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s: %w", src, ErrFileNotRegular)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer source.Close()

	err = os.MkdirAll(filepath.Dir(dst), mkdirPerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	os.Remove(dst)

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read from source: %w", err)
		}

		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write to destination: %w", err)
		}
	}

	return nil
}
