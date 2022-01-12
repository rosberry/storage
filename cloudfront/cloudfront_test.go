package cloudfront

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rosberry/storage/common"
	s3storage "github.com/rosberry/storage/s3"
)

var (
	testStorageKey = "cfs3file"
	prefix         = "test"

	s3Region    string
	s3AccessKey string
	s3SecretKey string
	s3Bucket    string

	domain = "d1kxflqjbui4hw.cloudfront.net"

	s3StorageInstance *s3storage.S3Storage
	testStorage       *CFStorage

	S3HostTemplate = s3storage.S3HostTemplate
)

func TestMain(m *testing.M) {
	s3Region = os.Getenv("S3_REGION")
	s3AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3SecretKey = os.Getenv("S3_SECRET_KEY")
	s3Bucket = os.Getenv("S3_BUCKET")

	domain = os.Getenv("CF_DOMAIN")

	if s3Region == "" || s3AccessKey == "" || s3SecretKey == "" || s3Bucket == "" || domain == "" {
		log.Print(`Please use environments for test cloudfront storage:
			S3_REGION
			S3_ACCESS_KEY
			S3_SECRET_KEY
			S3_BUCKET
			CF_DOMAIN
		`)
		os.Exit(1)
	}

	s3TestCfg := &s3storage.Config{
		StorageKey:      testStorageKey,
		Region:          s3Region,
		AccessKeyID:     s3AccessKey,
		SecretAccessKey: s3SecretKey,
		BucketName:      s3Bucket,
		Prefix:          prefix,
		NoSSL:           true,
	}

	s3StorageInstance = s3storage.New(s3TestCfg)
	// TODO: Credentials for cf
	testStorage = New(&Config{
		StorageKey:   testStorageKey,
		DomainName:   domain,
		CFPrefix:     prefix,
		NoSSL:        true,
		SignURLs:     false,
		StorageCtl:   s3StorageInstance,
		PrivateKeyID: "",
		PrivateKey:   "",
	})

	code := m.Run()

	clearBucket()

	os.Exit(code)
}

func TestGetUrl(t *testing.T) {
	flagtests := []struct {
		in  string
		out string
	}{
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(domain, prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(domain, prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "folder/file1.jpg", urlFormat(domain, prefix, "folder/file1.jpg")},
		{testStorageKey + ":" + "folder/folder2/file2.jpg", urlFormat(domain, prefix, "folder/folder2/file2.jpg")},
		{testStorageKey + ":" + "file1.jpg", urlFormat(domain, prefix, "file1.jpg")},
		{testStorageKey + ":" + "file 1.jpg", urlFormat(domain, prefix, "file%201.jpg")},
		{testStorageKey + ":" + "iFile1.jpg", urlFormat(domain, prefix, "iFile1.jpg")},
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

func urlFormat(domain, prefix, in string) string {
	return fmt.Sprintf("http://%s/%s/%s", domain, prefix, in)
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
	uploader := s3manager.NewUploader(s3Session(s3Region, s3AccessKey, s3SecretKey))

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
	svc := s3.New(s3Session(s3Region, s3AccessKey, s3SecretKey))

	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(s3Bucket),
	})

	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
		log.Printf("Unable to delete objects from bucket %v, %v", s3Bucket, err)
	}
}

func s3Session(region, accessKey, secretKey string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))
}
