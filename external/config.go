package external

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type ConfigurationSpec struct {
	LoginUrl         string `default:"http://127.0.0.1:7312" split_words:"true"`
	CourseManagerUrl string `default:"http://127.0.0.1:7313" split_words:"true"`
	StatsUrl         string `default:"http://127.0.0.1:7315" split_words:"true"`
	AchievementsUrl  string `default:"http://127.0.0.1:7316" split_words:"true"`
}

var config ConfigurationSpec

func InitConfig() {
	envconfig.MustProcess("", &config)
	log.Printf("external -> %+v\n", config)
}
