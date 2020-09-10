# Storage

## Supported storages:
- Direct links (```bypass```)
- S3 (```s3```)
- Cloudfront (```cf```)
- Yandex Object Storage (```yos```)

## Usage:
### Storage init
```golang
import "github.com/rosberry/storage"

//INITS
//direct links
storage.AddStorage("http", bypass.New())
storage.AddStorage("https", bypass.New())

//AWS S3
storage.AddStorage(instance.Key, s3.New(&s3.Config{
				StorageKey:      instance.Key,
				Region:          instance.Cfg["region"],
				AccessKeyID:     instance.Cfg["access_key_id"],
				SecretAccessKey: instance.Cfg["secret_access_key"],
				BucketName:      instance.Cfg["bucket_name"],
				Prefix:          instance.Cfg["prefix"],
			}))


//Yandex Object Storage
storage.AddStorage(instance.Key, yos.New(&yos.Config{
				StorageKey:      instance.Key,
				Region:          instance.Cfg["region"],
				AccessKeyID:     instance.Cfg["access_key_id"],
				SecretAccessKey: instance.Cfg["secret_access_key"],
				BucketName:      instance.Cfg["bucket_name"],
				Prefix:          instance.Cfg["prefix"],
			}))
        
//Cloudfront
storage.AddStorage(instance.Key, cf.New(&cf.Config{
				StorageKey: instance.Key,
				DomainName: instance.Cfg["domain_name"],
				CFPrefix:   instance.Cfg["cf_prefix"],
				StorageCtl: s3.New(&s3.Config{
					StorageKey:      instance.Key,
					Region:          instance.Cfg["region"],
					AccessKeyID:     instance.Cfg["access_key_id"],
					SecretAccessKey: instance.Cfg["secret_access_key"],
					BucketName:      instance.Cfg["bucket_name"],
					Prefix:          instance.Cfg["prefix"],
				}),
            }))

//Cloudfront with signed url's
storage.AddStorage(instance.Key, cf.New(&cf.Config{
				StorageKey:   instance.Key,
				DomainName:   instance.Cfg["domain_name"],
				CFPrefix:     instance.Cfg["cf_prefix"],
				SignURLs:     true,
				PrivateKeyID: instance.Cfg["private_key_id"],
				PrivateKey:   instance.Cfg["private_key"],
				StorageCtl: s3.New(&s3.Config{
					StorageKey:      instance.Key,
					Region:          instance.Cfg["region"],
					AccessKeyID:     instance.Cfg["access_key_id"],
					SecretAccessKey: instance.Cfg["secret_access_key"],
					BucketName:      instance.Cfg["bucket_name"],
					Prefix:          instance.Cfg["prefix"],
				}),
			}))
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


