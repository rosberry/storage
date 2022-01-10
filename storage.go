package storage

import (
	"github.com/rosberry/storage/core"
)

var aStorage = core.New()

//AddStorage - add new storage to storages list
func AddStorage(storageKey string, storage core.Storage) (err error) {
	return aStorage.AddStorage(storageKey, storage)
}

//GetStorage - get storage from list by storage by
func GetStorage(storageKey string) (s core.Storage, err error) {
	return aStorage.GetStorage(storageKey)
}

//CreateCLinkInStorage - save file and create clink in selected storage by storageKey
func CreateCLinkInStorage(filePath, path, storageKey string) (cLink string, err error) {
	return aStorage.CreateCLinkInStorage(filePath, path, storageKey)
}

//CreateCLink - save file and create cLink in default storage
func CreateCLink(filePath, path string) (cLink string, err error) {
	return aStorage.CreateCLink(filePath, path)
}

func PrepareCLinkInStorage(path, storageKey string) (cLink string, err error) {
	return aStorage.PrepareCLinkInStorage(path, storageKey)
}

func PrepareCLink(path string) (cLink string, err error) {
	return aStorage.PrepareCLink(path)
}

//GetURL - return http link by cLink
func GetURL(cLink string, options ...interface{}) (URL string) {
	return aStorage.GetURL(cLink, options...)
}

//Delete - delete file in storage by cLink
func Delete(cLink string) (err error) {
	return aStorage.Delete(cLink)
}

//SetDefaultStorage - set storage as default
func SetDefaultStorage(storageKey string) (err error) {
	return aStorage.SetDefaultStorage(storageKey)
}

func UploadByCLink(filePath, cLink string) (err error) {
	return aStorage.UploadByCLink(filePath, cLink)
}

func GetPathByCLink(cLink string) (path string) {
	return aStorage.GetPathByCLink(cLink)
}

