package s3

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rosberry/storage/common"
)

var (
	testStorageKey = "s3file"
	prefix         = "test"

	s3Region    string
	s3AccessKey string
	s3SecretKey string
	s3Bucket    string

	testStorage *S3Storage
)

func TestMain(m *testing.M) {
	s3Region = os.Getenv("S3_REGION")
	s3AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3SecretKey = os.Getenv("S3_SECRET_KEY")
	s3Bucket = os.Getenv("S3_BUCKET")

	if s3Region == "" || s3AccessKey == "" || s3SecretKey == "" || s3Bucket == "" {
		log.Print(`Please use environments for test s3 storage:
			S3_REGION
			S3_ACCESS_KEY
			S3_SECRET_KEY
			S3_BUCKET
		`)
		os.Exit(1)
	}

	testCfg := &Config{
		StorageKey:      testStorageKey,
		Region:          s3Region,
		AccessKeyID:     s3AccessKey,
		SecretAccessKey: s3SecretKey,
		BucketName:      s3Bucket,
		Prefix:          prefix,
		NoSSL:           true,
	}

	testStorage = New(testCfg)

	code := m.Run()

	clearBucket()

	os.Exit(code)
}

func TestGetUrl(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "file1.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "file1.jpg")},
		{testStorageKey + ":" + "file 1.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "file%201.jpg")},
		{testStorageKey + ":" + "iFile1.jpg", urlFormat(S3HostTemplate, s3Bucket, testStorage.cfg.Prefix, "iFile1.jpg")},
	}

	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			s := testStorage.GetURL(tt.in)
			if s == tt.out {
				t.Logf("got %v, want %v", s, tt.out)
			} else {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}

func urlFormat(host, bucket, prefix, in string) string {
	return fmt.Sprintf("http://%s/%s/%s", fmt.Sprintf(host, bucket), prefix, in)
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
	}

	tmp := "file1"
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

	os.Remove(tmp)
}

func TestStoreByCLink(t *testing.T) {
	flagtests := []struct {
		in string
	}{
		{"s3file:loadByClink/file1.jpg"},
		{"s3file:loadByClink/folder2/file2.jpg"},
	}

	tmp := "file1"
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

	os.Remove(tmp)
}

func TestRemove(t *testing.T) {
	tmp := "ifile1"
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

	os.Remove(tmp)
}

func checkFile(tmp string, cLink string) error {
	path := common.CLinkToPath(testStorage.cfg.StorageKey, cLink)
	internalPath := common.PathToInternalPath(testStorage.cfg.Prefix, path)

	// check file exists
	uploader := s3manager.NewUploader(testStorage.getSession())

	object, err := uploader.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(internalPath),
	})
	if err != nil {
		return fmt.Errorf("failed check '%s' object: %v", cLink, err)
	}

	tmpFileInfo, err := os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("Check tmp file err: %v", err)
	}

	if object.ContentLength == nil || *object.ContentLength != tmpFileInfo.Size() {
		return fmt.Errorf("Not equal size: %v", cLink)
	}

	return nil
}

func clearBucket() {
	svc := s3.New(testStorage.getSession())

	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(s3Bucket),
	})

	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
		log.Printf("Unable to delete objects from bucket %v, %v", s3Bucket, err)
	}
}
