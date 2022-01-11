package storage

import (
	"log"

	"github.com/rosberry/storage/bypass"
	"github.com/rosberry/storage/cloudfront"
	"github.com/rosberry/storage/core"
	"github.com/rosberry/storage/local"
	"github.com/rosberry/storage/s3"
	"github.com/rosberry/storage/yos/v2"
)

const (
	TypeBypass = "bypass"
	TypeLocal = "local"
	TypeS3 = "s3"
	TypeCloudFront = "cf"
	TypeCloudFrontSigned = "cfs"
	TypeYOS = "yos"
)

func NewWithConfig(config *core.StoragesConfig) *core.AbstractStorage {
	aStorage := core.New()

	aStorage.AddStorage("http", bypass.New())
	aStorage.AddStorage("https", bypass.New())

	if config == nil {
		return aStorage
	}

	for _, instance := range config.Instances {
		//key := strings.ToLower(instance.Key)
		key := instance.Key
		
		switch instance.Type {
		case TypeS3:
			aStorage.AddStorage(key, s3.New(&s3.Config{
				StorageKey:      key,
				Region:          instance.Cfg["region"],
				AccessKeyID:     instance.Cfg["access_key_id"],
				SecretAccessKey: instance.Cfg["secret_access_key"],
				BucketName:      instance.Cfg["bucket_name"],
				Prefix:          instance.Cfg["prefix"],
			}))
		case TypeCloudFront:
			aStorage.AddStorage(key, cloudfront.New(&cloudfront.Config{
				StorageKey: key,
				DomainName: instance.Cfg["domain_name"],
				CFPrefix:   instance.Cfg["cf_prefix"],
				StorageCtl: s3.New(&s3.Config{
					StorageKey:      key,
					Region:          instance.Cfg["region"],
					AccessKeyID:     instance.Cfg["access_key_id"],
					SecretAccessKey: instance.Cfg["secret_access_key"],
					BucketName:      instance.Cfg["bucket_name"],
					Prefix:          instance.Cfg["prefix"],
				}),
			}))
		case TypeCloudFrontSigned:
			aStorage.AddStorage(key, cloudfront.New(&cloudfront.Config{
				StorageKey:   key,
				DomainName:   instance.Cfg["domain_name"],
				CFPrefix:     instance.Cfg["cf_prefix"],
				SignURLs:     true,
				PrivateKeyID: instance.Cfg["private_key_id"],
				PrivateKey:   instance.Cfg["private_key"],
				StorageCtl: s3.New(&s3.Config{
					StorageKey:      key,
					Region:          instance.Cfg["region"],
					AccessKeyID:     instance.Cfg["access_key_id"],
					SecretAccessKey: instance.Cfg["secret_access_key"],
					BucketName:      instance.Cfg["bucket_name"],
					Prefix:          instance.Cfg["prefix"],
				}),
			}))
		case TypeYOS:
			aStorage.AddStorage(key, yos.New(&yos.Config{
				StorageKey:      key,
				Region:          instance.Cfg["region"],
				AccessKeyID:     instance.Cfg["access_key_id"],
				SecretAccessKey: instance.Cfg["secret_access_key"],
				BucketName:      instance.Cfg["bucket_name"],
				Prefix:          instance.Cfg["prefix"],
			}))
		case TypeLocal:
			aStorage.AddStorage(key, local.New(&local.Config{
				StorageKey: key,
				Endpoint:   instance.Cfg["endpoint"],
				Root:       instance.Cfg["root"],
				BufferSize: 32 * 1024, // TODO: Config?
			}))
		default:
			log.Printf("Storage type '%s' not supported!", instance.Type)
		}
	}

	if config.Default != "" {
		aStorage.SetDefaultStorage(config.Default)
	}

	return aStorage
}