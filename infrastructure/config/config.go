package config

import (
	"go-ddd/infrastructure/util/def"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const DefaultRelationPath = "./conf.yaml"

func Setup() {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件被修改:", e.Name)
	})
	viper.WatchConfig()

	env := viper.GetString("app.env")
	if env != def.EnvDevelopment && env != def.EnvTesting && env != def.EnvProduction {
		log.Fatal("app.env异常")
	}
}
