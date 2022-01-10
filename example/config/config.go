package config

import (
	"log"

	"github.com/jinzhu/configor"
	"github.com/rosberry/storage/core"
)


type (
	Config struct {
		Storages   core.StoragesConfig  `json:"storages" yaml:"storages"`
	}
)


func app() *Config {
	cfg := Config{}

	err := configor.Load(&cfg, ".env.yml")
	if err != nil {
		log.Fatal(err)
	}


	log.Printf("config: %+v", cfg)

	return &cfg
}

var App = app() // nolint:gochecknoglobals