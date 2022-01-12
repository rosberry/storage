package yos

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/rosberry/storage/common"
)

var (
	testStorageKey = "yosfile"
	prefix         = "test"

	yosRegion    string
	yosAccessKey string
	yosSecretKey string
	yosBucket    string

	testStorage *YandexObjStorage
)

func TestMain(m *testing.M) {
	yosRegion = os.Getenv("YOS_REGION")
	yosAccessKey = os.Getenv("YOS_ACCESS_KEY")
	yosSecretKey = os.Getenv("YOS_SECRET_KEY")
	yosBucket = os.Getenv("YOS_BUCKET")

	yosRegion = "ru"
	yosAccessKey = "xxx"
	yosSecretKey = "xxx"
	yosBucket = "bucket"

	if yosRegion == "" || yosAccessKey == "" || yosSecretKey == "" || yosBucket == "" {
		log.Print(`Please use environments for test cloudfront storage:
			YOS_REGION
			YOS_ACCESS_KEY
			YOS_SECRET_KEY
			YOS_BUCKET
		`)

		os.Exit(1)
	}

	testStorage = New(&Config{
		StorageKey:      testStorageKey,
		Region:          yosRegion,
		AccessKeyID:     yosAccessKey,
		SecretAccessKey: yosSecretKey,
		BucketName:      yosBucket,
		Prefix:          prefix,
		NoSSL:           true,
	})

	code := m.Run()

	// clearBucket()

	os.Exit(code)
}

func TestGetUrlPublic(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "file1.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "file1.jpg")},
		{testStorageKey + ":" + "file 1.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "file%201.jpg")},
		{testStorageKey + ":" + "iFile1.jpg", urlFormat(testStorage.endpoint, testStorage.cfg.BucketName, prefix, "iFile1.jpg")},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := testStorage.GetURL(tt.in, PublicLink)
			if s == tt.out {
				t.Logf("got %v, want %v", s, tt.out)
			} else {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}

func urlFormat(endpoint, bucket, prefix, in string) string {
	return fmt.Sprintf("https://%s/%s/%s/%s", endpoint, bucket, prefix, in)
}

func TestStore(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{"/folder/file1.jpg", testStorageKey + ":" + "folder/file1.jpg"},
		{"/folder/folder2/file2.jpg", testStorageKey + ":" + "folder/folder2/file2.jpg"},
		{"folder/file11.jpg", testStorageKey + ":" + "folder/file11.jpg"},
		{"folder/folder2/file12.jpg", testStorageKey + ":" + "folder/folder2/file12.jpg"},
		{"/file1.jpg", testStorageKey + ":" + "file1.jpg"},
		{"/file 1.jpg", testStorageKey + ":" + "file 1.jpg"},
		{"file4.jpg", testStorageKey + ":" + "file4.jpg"},
		{"file4", testStorageKey + ":" + "file4"},
	}

	tmp := "file1"
	defer os.Remove(tmp)

	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0o644)

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			cLink, err := testStorage.Store(tmp, tt.in)
			if err != nil {
				t.Errorf("Store err: %v", err)
			}
			if cLink != tt.out {
				t.Errorf("got %v, want %v", cLink, tt.out)
			} else {
				t.Logf("got %v, want %v", cLink, tt.out)
			}

			// check file exists
			err = checkFile(tmp, cLink)
			if err != nil {
				t.Errorf("failed check file %s: %v", tt.in, err)
			}
		})
	}
}

func TestStoreByCLink(t *testing.T) {
	flagtests := []struct {
		in string
	}{
		{testStorageKey + ":loadByClink/file1.jpg"},
		{testStorageKey + ":loadByClink/folder2/file2.jpg"},
	}

	tmp := "file1"
	defer os.Remove(tmp)

	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0o644)

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			err := testStorage.StoreByCLink(tmp, tt.in)
			if err != nil {
				t.Errorf("Store err: %v", err)
			}

			// check file exists
			err = checkFile(tmp, tt.in)
			if err != nil {
				t.Errorf("failed check file %s: %v", tt.in, err)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tmp := "ifile1"
	defer os.Remove(tmp)

	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0o644)
	cLink, _ := testStorage.Store(tmp, "/r_test/ifile.txt")

	path := common.CLinkToPath(testStorage.cfg.StorageKey, cLink)
	err := checkFile(tmp, cLink)
	if err != nil {
		t.Errorf("failed check file %s: %v", path, err)
	}

	err = testStorage.Remove(cLink)
	if err != nil {
		t.Errorf("Remove err: %v", err)
	}

	err = checkFile(tmp, cLink)
	if err == nil {
		t.Errorf("failed check file after delete %s: %v", path, err)
	}
}

func TestAccessForURL(t *testing.T) {
	tmp := "ifile1"
	defer os.Remove(tmp)

	ioutil.WriteFile(tmp, []byte("hello\ngo\n"), 0o644)
	cLink, _ := testStorage.Store(tmp, "/r_test/ifile.txt")

	path := common.CLinkToPath(testStorage.cfg.StorageKey, cLink)
	err := checkFile(tmp, cLink)
	if err != nil {
		t.Errorf("failed check file %s: %v", path, err)
	}

	url := testStorage.GetURL(cLink)
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("failed get resource %s -> %s: %v", cLink, url, err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("response code should be 200: %d", resp.StatusCode)
	}
}

func checkFile(tmp string, cLink string) error {
	path := common.CLinkToPath(testStorage.cfg.StorageKey, cLink)
	internalPath := common.PathToInternalPath(prefix, path)

	// check file exists
	object, err := testStorage.client.GetObject(context.TODO(), yosBucket, internalPath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed check '%s' object: %v", cLink, err)
	}

	tmpFileInfo, err := os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("check tmp file err: %w", err)
	}

	objectInfo, err := object.Stat()
	if err != nil {
		return fmt.Errorf("check object file stat err: %w", err)
	}

	if objectInfo.Size != tmpFileInfo.Size() {
		return fmt.Errorf("Not equal size: %v expected: %v", objectInfo.Size, tmpFileInfo.Size())
	}

	return nil
}

func clearBucket() {
	// TODO: Clear bucket after tests
}
