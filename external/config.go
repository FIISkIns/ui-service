package external

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type ConfigurationSpec struct {
	LoginUrl string `default:"http://127.0.0.1:7312" split_words:"true"`
}

var config ConfigurationSpec

func InitConfig() {
	envconfig.MustProcess("", &config)
	log.Printf("external -> %+v\n", config)
}
