# Storage

A wrapper for working with files in popular cloud storage

## Storage types:


### Supported:
- Direct links (```bypass```)
- Local (```local```)
- S3 (```s3```)
- Cloudfront (```cloudfront```)
- Yandex Object Storage (```yos```)

Types values:
- TypeBypass = "bypass"
- TypeLocal = "local"
- TypeS3 = "s3"
- TypeCloudFront = "cf"
- TypeCloudFrontSigned = "cfs"
- TypeYOS = "yos"

Each of the types implements the interface:
```golang
Storage interface {
	Store(filePath, path string) (cLink string, err error)
	GetURL(cLink string, options ...interface{}) (URL string)
	Remove(cLink string) (err error)
	GetCLink(path string) (cLink string)
	StoreByCLink(filePath, cLink string) (err error)
}
```

You can create and use each of the types of storages separately.

Example:
```golang
import "github.com/rosberry/storage/local"

lStorage := local.New(&local.Config{
	StorageKey: cfg["storageKey"],
	Endpoint:   cfg["endpoint"],
	Root:       cfg["root"],
	BufferSize: 32 * 1024,
})

cLink, err := lStorage.Store(filePath, path)
...

url := lStorage.GetURL(cLink)
...
```

### Create
#### Bypass

```golang
import "github.com/rosberry/storage/bypass"

bpStorage := bypass.New()
```

#### Local
```golang
import "github.com/rosberry/storage/local"

lStorage := local.New(&local.Config{
	StorageKey: cfg["storageKey"],
	Endpoint:   cfg["endpoint"],
	Root:       cfg["root"],
	BufferSize: 32 * 1024,
})
```

#### S3
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

#### Cloudfront
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

#### YOS
```golang
import "github.com/rosberry/storage/yos/v2"

yosStorage := yos.New(&yos.Config{
	StorageKey:      cfg["storage_key"],
	Region:          cfg["region"],
	AccessKeyID:     cfg["access_key_id"],
	SecretAccessKey: cfg["secret_access_key"],
	BucketName:      cfg["bucket_name"],
	Prefix:          cfg["prefix"],
})
```

## Abstract storage

But the correct use of the library - use of an abstract storage, which includes one or more implementations of storage types.
You do not have to use specific types of methods. Use abstract storage methods.

#### Global instance (classic flow)

```golang
import (
	"github.com/rosberry/storage"
	"github.com/rosberry/storage/bypass"
	"github.com/rosberry/storage/local"
	"github.com/rosberry/storage/s3"
	"github.com/rosberry/storage/yos/v2"
	"github.com/rosberry/storage/cloudfront"
)

// Init storage types
// ....

// Add storage types
// bypass
storage.AddStorage("http", bypass.New())
storage.AddStorage("https", bypass.New())
// local
storage.AddStorage(localStorageKey, lStorage)
// s3
storage.AddStorage(s3StorageKey, s3Storage)
// yos
storage.AddStorage(yosStoragKey, yosStorage)
// cloudfront
storage.AddStorage(cfStorageKey, cfStorage)
storage.AddStorage(cfsStorageKey, cfsStorage)

// usage
cLink, err := storage.CreateCLinkInStorage(filePath, path, s3StorageKey)
...

url := storage.GetURL(cLink)
...
```

#### Local instance (experimental flow)
You can create an unlimited number of the local instances of the abstract storage and use the same methods.
```golang
// ...
import "github.com/rosberry/storage/core"

localInstance := core.New()

// Init storage types
// ....

// Add storage types
// bypass
localInstance.AddStorage("http", bypass.New())
localInstance.AddStorage("https", bypass.New())
// local
localInstance.AddStorage(localStorageKey, lStorage)
// ...

// usage
cLink, err := localInstance.CreateCLinkInStorage(filePath, path, s3StorageKey)
...

url := localInstance.GetURL(cLink)
...

```

You can also use instance creation with automatic configuration parsing
```golang
// ...
import "github.com/rosberry/storage/core"

//...
var config *core.StoragesConfig

// Parsing config from file
//...

localInstance := storage.NewWithConfig(&config)

// usage
cLink, err := localInstance.CreateCLinkInStorage(filePath, path, s3StorageKey)
...

url := localInstance.GetURL(cLink)
...

```

The configuration is the type:
```golang
// package github.com/rosberry/storage/core

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
```

### Abstract storage methods

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

#### Prepare cLink and upload in two step
You can create cLink without upload file and then upload file later.

Prepare cLink (by analogy with the `CreateCLink...`):
```golang
//default storage
func PrepareCLinkInStorage(path, storageKey string) (cLink string, err error)

//selected storage
func PrepareCLink(path string) (cLink string, err error)
```

Upload file:
```golang
func UploadByCLink(filePath, cLink string) (err error) 
```

## Example

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

## Restrictions and well-known problems
- You can not use the '_' symbol in the key
- The key must be in the lower case

## Code glossary
// Glossary
//	path - The name with which the user wants to save the file
//	internalPath - The path to the file in the repository (includes the prefix)
//	cLink - Formed as a storage key + ':' + path

## About

<img src="https://github.com/rosberry/Foundation/blob/master/Assets/full_logo.png?raw=true" height="100" />

This project is owned and maintained by [Rosberry](http://rosberry.com). We build mobile apps for users worldwide 🌏.

Check out our [open source projects](https://github.com/rosberry), read [our blog](https://medium.com/@Rosberry) or give us a high-five on 🐦 [@rosberryapps](http://twitter.com/RosberryApps).

## License

This project is available under the MIT license. See the LICENSE file for more info.


