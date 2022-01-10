package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rosberry/storage"
	"github.com/rosberry/storage/core"
	"github.com/rosberry/storage/example/config"
)


func main() {
	log.Print("Hello!")

	log.Print("Init storage")
	storageInstance := storage.NewWithConfig(&config.App.Storages)
	if storageInstance == nil {
		log.Fatal("Failed to initialize storage")
	}

	log.Print("Generate test file")
	testFile, err := createTestFile()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Save test file in all storage instances")
	cLinks := saveFile(storageInstance, testFile)
	if !mapNotEmpty(cLinks) {
		log.Fatal("Failed to save test file")
	}

	printMap("cLinks:", cLinks)

	log.Print("Get urls to files in all storage")
	urls := getLinks(storageInstance, cLinks)
	if !mapNotEmpty(urls) {
		log.Fatal("Failed to get urls")
	}

	printMap("URLs:", urls)

	log.Print("End")
}

func createTestFile() (string, error) {
	f, err := os.Create(fmt.Sprintf("test_file_%s.txt", time.Now().Format("20060602030405")))
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Storage lib test file\n%s", time.Now().Format(time.RFC3339)))
    if err != nil {
        return "", err
    }

	return f.Name(), nil
}

func saveFile(storageInstance *core.AbstractStorage, filePath string) map[string]string {
	cLinks := map[string]string{}

	for _, st := range config.App.Storages.Instances {
		cLink, err := storageInstance.CreateCLinkInStorage(filePath, filePath, st.Key)
		if err != nil {			
			log.Printf("Failed save test file in storage %s: %v", st.Key, err)
			cLink = ""
		} else {
			log.Printf("Success save in storage %s: %s", st.Key, cLink)
		}

		cLinks[st.Key] = cLink
	}

	return cLinks
}

func getLinks(storageInstance *core.AbstractStorage, cLinks map[string]string) map[string]string {
	links := map[string]string{}

	for storageKey, cLink := range cLinks {
		if cLink == "" {
			links[storageKey] = ""
			continue
		}

		link := storageInstance.GetURL(cLink)
		if link == "" {
			log.Printf("Failed to get link for storage %s", storageKey)
		}

		links[storageKey] = link
	}

	return links
}

func mapNotEmpty(m map[string]string) bool {
	for _, v := range m {
		if v != "" {
			return true
		}
	}

	return false
}

func printMap(title string, m map[string]string) {
	sb := strings.Builder{}
	sb.WriteString(title)
	sb.WriteString("\n")

	for key, val := range m {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", key, val))	
	}

	log.Print(sb.String())
}