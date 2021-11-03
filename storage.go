package storage

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

type (
	ExpirationVerifier interface {
		GetAccessExpireTime(cLink string) (expires time.Time)
	}

	Storage interface {
		Store(filePath, path string) (cLink string, err error)
		GetURL(cLink string, options ...interface{}) (URL string)
		Remove(cLink string) (err error)
		GetCLink(path string) (cLink string)
		StoreByCLink(filePath, cLink string) (err error)
	}

	abstractStorage struct {
		defaultStorageKey *string
		storages          map[string]Storage
	}
)

var (
	ErrStorageNotFound   = errors.New("Storage not found")
	ErrStorageNil        = errors.New("Storage pointer is nil")
	ErrStorageKeyIsEmpty = errors.New("Storage key is empty")
	ErrNoDefaultStorage  = errors.New("Default storage not specified")
	ErrCLinkError        = errors.New("CLink error")
)

var aStorage = &abstractStorage{
	storages: make(map[string]Storage),
}

//AddStorage - add new storage to storages list
func AddStorage(storageKey string, storage Storage) (err error) {
	if storageKey == "" {
		return ErrStorageKeyIsEmpty
	}
	if storage == nil {
		return ErrStorageNil
	}
	aStorage.storages[storageKey] = storage
	return
}

//GetStorage - get storage from list by storage by
func GetStorage(storageKey string) (s Storage, err error) {
	var ok bool
	if s, ok = aStorage.storages[storageKey]; !ok {
		return nil, ErrStorageNotFound
	}
	return
}

func (a *abstractStorage) getStorage(storageKey string) (s Storage, err error) {
	var ok bool
	if s, ok = a.storages[storageKey]; !ok {
		return nil, ErrStorageNotFound
	}
	return
}

func (a *abstractStorage) getStorageByCLink(cLink string) (s Storage, err error) {
	u, e := url.Parse(cLink)
	if e != nil || u.Scheme == "" {
		return nil, ErrCLinkError
	}
	return a.getStorage(u.Scheme)
}

//CreateCLinkInStorage - save file and create clink in selected storage by storageKey
func CreateCLinkInStorage(filePath, path, storageKey string) (cLink string, err error) {
	s, e := aStorage.getStorage(storageKey)
	if e != nil {
		return "", e
	}
	return s.Store(filePath, path)
}

//CreateCLink - save file and create cLink in default storage
func CreateCLink(filePath, path string) (cLink string, err error) {
	if aStorage.defaultStorageKey == nil {
		err = ErrNoDefaultStorage
		return
	}
	return CreateCLinkInStorage(filePath, path, *aStorage.defaultStorageKey)
}

func PrepareCLinkInStorage(path, storageKey string) (cLink string, err error) {
	s, e := aStorage.getStorage(storageKey)
	if e != nil {
		return "", e
	}
	return s.GetCLink(path), nil
}

func PrepareCLink(path string) (cLink string, err error) {
	if aStorage.defaultStorageKey == nil {
		err = ErrNoDefaultStorage
		return
	}
	return PrepareCLinkInStorage(path, *aStorage.defaultStorageKey)
}

//GetURL - return http link by cLink
func GetURL(cLink string, options ...interface{}) (URL string) {
	s, err := aStorage.getStorageByCLink(cLink)
	if err != nil {
		return ""
	}
	return s.GetURL(cLink, options...)
}

//Delete - delete file in storage by cLink
func Delete(cLink string) (err error) {
	s, e := aStorage.getStorageByCLink(cLink)
	if e != nil {
		return e
	}
	return s.Remove(cLink)
}

//SetDefaultStorage - set storage as default
func SetDefaultStorage(storageKey string) (err error) {
	if _, ok := aStorage.storages[storageKey]; !ok {
		return ErrStorageNotFound
	}
	aStorage.defaultStorageKey = &storageKey
	return
}

func UploadByCLink(filePath, cLink string) (err error) {
	s, err := aStorage.getStorageByCLink(cLink)
	if err != nil {
		return fmt.Errorf("failed get storage by cLink: %w", err)
	}


}