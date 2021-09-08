package local

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	testEndpoint            = "http://url.com/files"
	testStorageKey          = "localFile"
	testRoot                = "test/"
	testRootWithoutEndSlash = strings.TrimRight(testRoot, "/")

	testCfg = &Config{
		StorageKey: testStorageKey,
		Endpoint:   testEndpoint,
		Root:       testRoot,
		BufferSize: 32 * 1024,
	}

	testStorage = New(testCfg)
)

func TestEndSlash(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{"/folder/file1.jpg", "/folder/file1.jpg/"},
		{"/file1.jpg", "/file1.jpg/"},
		{"http://ya.ru", "http://ya.ru/"},
		{"http://ya.ru/", "http://ya.ru/"},
		{"http://ya.ru/page", "http://ya.ru/page/"},
		{"http://ya.ru/page/", "http://ya.ru/page/"},
		{"file1", "file1/"},
		{"file1/", "file1/"},
		{"file1//", "file1/"},
		{"file1///", "file1/"},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := endSlash(tt.in)
			if s == tt.out {
				t.Logf("got %q, want %q", s, tt.out)
			} else {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

func TestPathToCLink(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{"/folder/file1.jpg", testStorageKey + ":" + "folder/file1.jpg"},
		{"/folder/folder2/file2.jpg", testStorageKey + ":" + "folder/folder2/file2.jpg"},
		{"folder/file1.jpg", testStorageKey + ":" + "folder/file1.jpg"},
		{"folder/folder2/file2.jpg", testStorageKey + ":" + "folder/folder2/file2.jpg"},
		{"/file1.jpg", testStorageKey + ":" + "file1.jpg"},
		{"/file 1.jpg", testStorageKey + ":" + "file 1.jpg"},
		{"file1.jpg", testStorageKey + ":" + "file1.jpg"},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := testStorage.pathToCLink(tt.in)
			if s == tt.out {
				t.Logf("got %q, want %q", s, tt.out)
			} else {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

func TestCLinkToUrl(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{"localFile:folder/file1.jpg", testRootWithoutEndSlash + "/folder/file1.jpg"},
		{"localFile:folder/folder2/file2.jpg", testRootWithoutEndSlash + "/folder/folder2/file2.jpg"},
		{"localFile:file1.jpg", testRootWithoutEndSlash + "/file1.jpg"},
		{"localFile:file 1.jpg", testRootWithoutEndSlash + "/file 1.jpg"},
		{"anotherKey:file1.jpg", ""},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := testStorage.cLinkToPath(tt.in)
			if s == tt.out {
				t.Logf("got %q, want %q", s, tt.out)
			} else {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

func TestGetUrl(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{testStorageKey + ":" + "folder/file1.jpg", endSlash(testEndpoint) + "folder/file1.jpg"},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", endSlash(testEndpoint) + "folder/folder2/file2.jpg"},
		{testStorageKey + ":" + "folder/file1.jpg", endSlash(testEndpoint) + "folder/file1.jpg"},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", endSlash(testEndpoint) + "folder/folder2/file2.jpg"},
		{testStorageKey + ":" + "file1.jpg", endSlash(testEndpoint) + "file1.jpg"},
		{testStorageKey + ":" + "file 1.jpg", endSlash(testEndpoint) + "file%201.jpg"},
		{testStorageKey + ":" + "iFile1.jpg", endSlash(testEndpoint) + "iFile1.jpg"},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := testStorage.GetURL(tt.in)
			if s == tt.out {
				t.Logf("got %q, want %q", s, tt.out)
			} else {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

func TestStore(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{"/folder/file1.jpg", testStorageKey + ":" + "folder/file1.jpg"},
		{"/folder/folder2/file2.jpg", testStorageKey + ":" + "folder/folder2/file2.jpg"},
		{"folder/file1.jpg", testStorageKey + ":" + "folder/file1.jpg"},
		{"folder/folder2/file2.jpg", testStorageKey + ":" + "folder/folder2/file2.jpg"},
		{"/file1.jpg", testStorageKey + ":" + "file1.jpg"},
		{"/file 1.jpg", testStorageKey + ":" + "file 1.jpg"},
		{"file1.jpg", testStorageKey + ":" + "file1.jpg"},
	}

	tmp := "file1"
	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0644)

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			cLink, err := testStorage.Store(tmp, tt.in)
			if err != nil {
				t.Errorf("Store err: %q", err)
			}
			if cLink != tt.out {
				t.Errorf("got %q, want %q", cLink, tt.out)
			} else {
				t.Logf("got %q, want %q", cLink, tt.out)
			}

			fullPath := testStorage.cLinkToPath(cLink)
			fileInfo, err := os.Stat(fullPath)
			if err != nil {
				t.Errorf("Check file err: %q", err)
			}
			tmpFileInfo, _ := os.Stat(tmp)
			if fileInfo.Size() != tmpFileInfo.Size() {
				t.Errorf("Not equal size: %q", cLink)
			}
		})
	}

	// clear files
	os.Remove(tmp)
	os.RemoveAll(testRoot)
}

func TestRemove(t *testing.T) {
	tmp := "ifile1"
	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0644)
	cLink, _ := testStorage.Store(tmp, "/r_test/ifile.txt")

	fullPath := testStorage.cLinkToPath(cLink)
	_, err := os.Stat(fullPath)
	if err != nil {
		t.Errorf("Check file err: %q", err)
	}

	err = testStorage.Remove(cLink)
	if err != nil {
		t.Errorf("Remove err: %q", err)
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		t.Errorf("Check removed file err: %q", err)
	}

	// clear
	os.Remove(tmp)
	os.RemoveAll(testRoot)
}
