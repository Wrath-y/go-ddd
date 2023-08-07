package main

import (
	"flag"
	"github.com/spf13/viper"
	"go-ddd/infrastructure/config"
	"go-ddd/infrastructure/util/consul"
	"go-ddd/infrastructure/util/db"
	"go-ddd/infrastructure/util/def"
	"go-ddd/infrastructure/util/goredis"
	"go-ddd/infrastructure/util/logging"
	"go-ddd/launch/grpc"
	"log"
	"syscall"
)

// setupConfigYaml 就绪配置文件
// 环境变量配置 NACOS_SKIP="Y", 可跳过下载配置
// 环境变量:
// NACOS_USE=false
// NACOS_NAMESPACE=""
// NACOS_SERVER=""
// NACOS_USERNAME=""
// NACOS_PASSWORD=""
func setupConfigYaml() {
	viper.AutomaticEnv()
	if envUse := viper.GetBool("NACOS_USE"); !envUse {
		log.Println("跳过从nacos下载配置文件")
		return
	}

	config.SetupNacosClient()
	config.DownloadNacosConfig()

	env := viper.GetString("app.env")
	if env != def.EnvDevelopment && env != def.EnvTesting && env != def.EnvProduction {
		log.Fatal("app.env异常")
	}
	// 监听nacos（已经被使用的变量变了也不会体现出变化）
	config.ListenNacos(func(cnf string) {
		if env == def.EnvDevelopment || env == def.EnvTesting || env == def.EnvProduction {
			return
		}

		// when use k8s
		log.Println("当前进程将被停止")
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	})
}

func setup() {
	config.Setup()
	logging.Setup()
	goredis.Setup()
	db.Setup()
	consul.Setup()
}

func main() {
	flag.Parse()
	setupConfigYaml()

	setup()

	grpc.RunGrpc()

	logging.New().Info("Server exited", "", nil)
}
