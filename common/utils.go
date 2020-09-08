package common

import (
	"log"
	"net/http"
	"os"
)

func GetFileContentType(out *os.File) string {
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		log.Println(err)
		return ""
	}

	out.Seek(0, 0)

	contentType := http.DetectContentType(buffer)
	return contentType
}
