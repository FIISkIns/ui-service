package main

import "github.com/kelseyhightower/envconfig"

type ConfigurationSpec struct {
	Port      int    `default:"7311"`
	CourseUrl string `default:"http://127.0.0.1:7310" envconfig:"COURSE_URL"`
}

var config ConfigurationSpec

func initConfig() {
	envconfig.MustProcess("ui", &config)
}
