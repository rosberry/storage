package core

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type (
	Storage interface {
		Store(filePath, path string) (cLink string, err error)
		GetURL(cLink string, options ...interface{}) (URL string)
		Remove(cLink string) (err error)
		GetCLink(path string) (cLink string)
		StoreByCLink(filePath, cLink string) (err error)
	}

	ExpirationVerifier interface {
		GetAccessExpireTime(cLink string) (expires time.Time)
	}
)

type AbstractStorage struct {
	defaultStorageKey *string
	storages          map[string]Storage
}

type (
	StoragesConfig struct {
		Default   string          `json:"default" yaml:"default"`
		Instances []StorageConfig `json:"instances" yaml:"instances"`
	}

	StorageConfig struct {
		Key  string            `json:"key" yaml:"key"`
		Type string            `json:"type" yaml:"type"`
		Cfg  map[string]string `json:"config" yaml:"config"`
	}
)

var (
	ErrStorageNotFound   = errors.New("Storage not found")
	ErrStorageNil        = errors.New("Storage pointer is nil")
	ErrStorageKeyIsEmpty = errors.New("Storage key is empty")
	ErrNoDefaultStorage  = errors.New("Default storage not specified")
	ErrCLinkError        = errors.New("CLink error")
)

func New() *AbstractStorage {
	return &AbstractStorage{
		storages: make(map[string]Storage),
	}
}

//AddStorage - add new storage to storages list
func (aStorage *AbstractStorage) AddStorage(storageKey string, storage Storage) (err error) {
	if storageKey == "" {
		return ErrStorageKeyIsEmpty
	}
	if storage == nil {
		return ErrStorageNil
	}

	aStorage.storages[strings.ToLower(storageKey)] = storage
	return
}

//GetStorage - get storage from list by storage by
func (aStorage *AbstractStorage) GetStorage(storageKey string) (s Storage, err error) {
	var ok bool
	if s, ok = aStorage.storages[storageKey]; !ok {
		return nil, ErrStorageNotFound
	}
	return
}

func (aStorage *AbstractStorage) getStorage(storageKey string) (s Storage, err error) {
	var ok bool
	if s, ok = aStorage.storages[strings.ToLower(storageKey)]; !ok {
		return nil, ErrStorageNotFound
	}
	return
}

func (aStorage *AbstractStorage) getStorageByCLink(cLink string) (s Storage, err error) {
	u, e := url.Parse(cLink)
	if e != nil || u.Scheme == "" {
		return nil, ErrCLinkError
	}
	return aStorage.getStorage(u.Scheme)
}

//CreateCLinkInStorage - save file and create clink in selected storage by storageKey
func (aStorage *AbstractStorage) CreateCLinkInStorage(filePath, path, storageKey string) (cLink string, err error) {
	s, e := aStorage.getStorage(storageKey)
	if e != nil {
		return "", e
	}
	return s.Store(filePath, path)
}

//CreateCLink - save file and create cLink in default storage
func (aStorage *AbstractStorage) CreateCLink(filePath, path string) (cLink string, err error) {
	if aStorage.defaultStorageKey == nil {
		err = ErrNoDefaultStorage
		return
	}

	return aStorage.CreateCLinkInStorage(filePath, path, *aStorage.defaultStorageKey)
}


func (aStorage *AbstractStorage) PrepareCLinkInStorage(path, storageKey string) (cLink string, err error) {
	s, e := aStorage.getStorage(storageKey)
	if e != nil {
		return "", e
	}
	return s.GetCLink(path), nil
}

func (aStorage *AbstractStorage) PrepareCLink(path string) (cLink string, err error) {
	if aStorage.defaultStorageKey == nil {
		err = ErrNoDefaultStorage
		return
	}
	return aStorage.PrepareCLinkInStorage(path, *aStorage.defaultStorageKey)
}

//GetURL - return http link by cLink
func (aStorage *AbstractStorage) GetURL(cLink string, options ...interface{}) (URL string) {
	s, err := aStorage.getStorageByCLink(cLink)
	if err != nil {
		return ""
	}
	return s.GetURL(cLink, options...)
}

//Delete - delete file in storage by cLink
func (aStorage *AbstractStorage) Delete(cLink string) (err error) {
	s, e := aStorage.getStorageByCLink(cLink)
	if e != nil {
		return e
	}
	return s.Remove(cLink)
}

//SetDefaultStorage - set storage as default
func (aStorage *AbstractStorage) SetDefaultStorage(storageKey string) (err error) {
	if _, ok := aStorage.storages[storageKey]; !ok {
		return ErrStorageNotFound
	}
	aStorage.defaultStorageKey = &storageKey
	return
}

func (aStorage *AbstractStorage) UploadByCLink(filePath, cLink string) (err error) {
	s, err := aStorage.getStorageByCLink(cLink)
	if err != nil {
		return fmt.Errorf("failed get storage by cLink: %w", err)
	}

	err = s.StoreByCLink(filePath, cLink)
	return err
}

func (aStorage *AbstractStorage) GetPathByCLink(cLink string) (path string) {
	return cLink[strings.LastIndex(cLink, ":")+1:]
}