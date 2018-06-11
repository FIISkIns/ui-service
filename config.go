package main

import (
	"github.com/FIISkIns/ui-service/external"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type ConfigurationSpec struct {
	Port       int    `default:"7311"`
	SessionKey string `split_words:"true"`
	CourseUrl  string `default:"http://127.0.0.1:7310" envconfig:"COURSE_URL"`
}

var config ConfigurationSpec

func initConfig() {
	envconfig.MustProcess("ui", &config)
	log.Printf("main -> %+v\n", config)

	external.InitConfig()
}
