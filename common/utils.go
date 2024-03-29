// Glossary
//	path - The name with which the user wants to save the file
//	internalPath - The path to the file in the repository (includes the prefix)
//	cLink - Formed as a storage key + ':' + path
package common

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const bufferSize = 512

func GetFileContentType(out *os.File) string {
	buffer := make([]byte, bufferSize)

	_, err := out.Read(buffer)
	if err != nil {
		log.Println(err)
		return ""
	}

	out.Seek(0, 0) // nolint:errcheck

	contentType := http.DetectContentType(buffer)

	return contentType
}

// path -> internalPath
//
// Transform path to internal path
// 		path - as user want save file
// 		internalPath - full path in storage
//
// For backward convertation use InternalPathToPath(prefix, internalPath string) (path string)
func PathToInternalPath(prefix, path string) (internalPath string) {
	return endSlash(prefix) + strings.Trim(path, "/")
}

// internalPath -> path
//
// Backward convertation after PathToInternalPath(prefix, path string) (internalPath string)
func InternalPathToPath(prefix, internalPath string) (path string) {
	return strings.TrimPrefix(internalPath, endSlash(prefix))
}

// path -> cLink
//
// Transform path to cLink
//
// For backward convertation use CLinkToPath(storageKey, cLink string) (path string)
func PathToCLink(storageKey, path string) (cLink string) {
	return fmt.Sprintf("%s:%s", storageKey, strings.Trim(path, "/"))
}

// cLink -> path
//
// Backward convertation after PathToCLink(storageKey, path string) (cLink string)
func CLinkToPath(storageKey, cLink string) (path string) {
	if !CheckStorageKey(cLink, storageKey) {
		return ""
	}

	return strings.TrimPrefix(cLink, storageKey+":")
}

// Checks the string to end with the slash
func endSlash(s string) string {
	return strings.TrimRight(s, "/") + "/"
}

// Checks the cLink contain storage key
func CheckStorageKey(cLink string, storageKey string) (ok bool) {
	return strings.Contains(cLink, storageKey+":")
}
