# Storage

A wrapper for working with files in popular cloud storage

## Supported storages:
- Direct links (```bypass```)
- Local (```local```)
- S3 (```s3```)
- Cloudfront (```cloudfront```)
- Yandex Object Storage (```yos```)

## Init:

### Bypass

```golang
import "github.com/rosberry/storage/bypass"

bpStorage := bypass.New()
```

### Local
```golang
import "github.com/rosberry/storage/local"

lStorage := local.New(local.Config{
	StorageKey: cfg["storageKey"],
	Endpoint:   cfg["endpoint"],
	Root:       cfg["root"],
	BufferSize: 32 * 1024,
})
```

### S3
```golang
import "github.com/rosberry/storage/s3"

s3Storage := s3.New(&s3.Config{
	StorageKey:      cfg["storage_key"],
	Region:          cfg["region"],
	AccessKeyID:     cfg["access_key_id"],
	SecretAccessKey: cfg["secret_access_key"],
	BucketName:      cfg["bucket_name"],
	Prefix:          cfg["prefix"],
})
```

### Cloudfront
```golang
import "github.com/rosberry/storage/cloudfront"

cfStorage := cloudfront.New(&cloudfront.Config{
	StorageKey: cfg["storage_key"],
	DomainName: cfg["domain_name"],
	CFPrefix:   cfg["cf_prefix"],
	StorageCtl: s3Storage, // see section S3
})

// with signed url's
cfStorage := cloudfront.New(&cloudfront.Config{
	StorageKey:   cfg["storage_key"],
	DomainName:   cfg["domain_name"],
	CFPrefix:     cfg["cf_prefix"],
	SignURLs:     true,
	PrivateKeyID: cfg["private_key_id"],
	PrivateKey:   cfg["private_key"],
	StorageCtl: s3Storage,
})
```

### YOS
```golang
import "github.com/rosberry/storage/yos"

yosStorage := yos.New(&yos.Config{
	StorageKey:      cfg["storage_key"],
	Region:          cfg["region"],
	AccessKeyID:     cfg["access_key_id"],
	SecretAccessKey: cfg["secret_access_key"],
	BucketName:      cfg["bucket_name"],
	Prefix:          cfg["prefix"],
})
```

## Multiple storages
```golang
import "github.com/rosberry/storage"
import "github.com/rosberry/storage/bypass"
import "github.com/rosberry/storage/local"
import "github.com/rosberry/storage/s3"
import "github.com/rosberry/storage/yos"
import "github.com/rosberry/storage/cloudfront"

// bypass
storage.AddStorage("http", bypass.New())
storage.AddStorage("https", bypass.New())
// local
storage.AddStorage(cfg["local_storage_key"], lStorage)
// s3
storage.AddStorage(cfg["s3_storage_key"], s3Storage)
// yos
storage.AddStorage(cfg["yos_storage_key"], yosStorage)
// Cloudfront
storage.AddStorage(cfg["cf_storage_key"], cfStorage)

// usage
st := storage.GetStorage("s3_storage_key")
```

### Functions

Add storage to storage list
```golang
func AddStorage(storageKey string, storage Storage) (err error) 
```

Get storage from storage list
```golang
func GetStorage(storageKey string) (s Storage, err error)
```

Save file to storage and create storage link
```golang
//default storage
func CreateCLink(filePath, path string) (cLink string, err error)

//selected storage
func CreateCLinkInStorage(filePath, path, storageKey string) (cLink string, err error)
```

Return http link storage link
```golang
func GetURL(cLink string, options ...interface{}) (URL string)
```

Delete file in storage
```golang
func Delete(cLink string) (err error)
```

Set storage as default
```golang
func SetDefaultStorage(storageKey string) (err error)
```

### Example

```golang
if sp.Photo != "" {
	tmpfile, _ := ioutil.TempFile("", "avatar-*.jpg")
	tmpPath := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(tmpPath)

	err := downloadFile(tmpPath, sp.Photo)
	if err != nil {
		log.Println(err)
	}

	link, err := storage.CreateCLinkInStorage(tmpPath, fmt.Sprintf("users/%v", user.ID), "yos")
	if err != nil {
		log.Printf("Failed download user avatar for user %v\n", u.ID)
		user.Photo = sp.Photo
	} else {
		user.Photo = link
	}
}
```

## About

<img src="https://github.com/rosberry/Foundation/blob/master/Assets/full_logo.png?raw=true" height="100" />

This project is owned and maintained by [Rosberry](http://rosberry.com). We build mobile apps for users worldwide 🌏.

Check out our [open source projects](https://github.com/rosberry), read [our blog](https://medium.com/@Rosberry) or give us a high-five on 🐦 [@rosberryapps](http://twitter.com/RosberryApps).

## License

This project is available under the MIT license. See the LICENSE file for more info.


