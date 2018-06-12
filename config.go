package main

import (
	"github.com/FIISkIns/ui-service/external"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type ConfigurationSpec struct {
	Port       int    `default:"7311"`
	SessionKey string `split_words:"true"`
}

var config ConfigurationSpec

func initConfig() {
	envconfig.MustProcess("ui", &config)
	log.Printf("main -> %+v\n", config)

	external.InitConfig()
}
