package config

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go-ddd/infrastructure/util/nacos"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
)

var (
	nacosParams *_nacosParams
	nacosClient *nacos.NacosConfig
)

const (
	defaultPollTime = 60 * time.Second
	defaultTimeout  = 30 * time.Second
)

type _nacosParams struct {
	address   string
	username  string
	password  string
	dataId    string
	group     string
	namespace string
	pollTime  time.Duration
	timeout   time.Duration
}

type logger struct {
}

func (l *logger) Error(v ...interface{}) {
	log.Println(v...)
}

func (l *logger) Info(v ...interface{}) {
	log.Println(v...)
}

//func (l *logger) Debug(v ...interface{}) {
//
//}

func loadNacosParams() (*_nacosParams, error) {
	viper.SetConfigName("nacos")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 读环境变量
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	namespace := viper.GetString("NACOS_NAMESPACE")
	conf := &_nacosParams{
		address:   viper.GetString("NACOS_SERVER"),
		username:  viper.GetString("NACOS_USERNAME"),
		password:  viper.GetString("NACOS_PASSWORD"),
		namespace: namespace,
		dataId:    viper.GetString(namespace + ".data_id"),
		group:     viper.GetString(namespace + ".group"),
		pollTime:  viper.GetDuration(namespace + ".poll_time"),
		timeout:   viper.GetDuration(namespace + ".timeout"),
	}

	if conf.address == "" || conf.username == "" || conf.password == "" || conf.namespace == "" {
		return nil, errors.New("环境变量中缺少nacos配置")
	}
	if conf.dataId == "" || conf.group == "" {
		return nil, errors.New("本地配置文件中缺少nacos配置")
	}
	if conf.pollTime == 0 {
		conf.pollTime = defaultPollTime
	}
	if conf.timeout == 0 {
		conf.timeout = defaultTimeout
	}

	return conf, nil
}

func SetupNacosClient() {
	var err error
	nacosParams, err = loadNacosParams()
	if err != nil {
		log.Fatalf("加载nacos配置失败: %s", err.Error())
	}
	nacosClient = nacos.NewNacosConfig(func(c *nacos.NacosConfig) {
		c.ServerAddr = nacosParams.address
		c.Username = nacosParams.username
		c.Password = nacosParams.password
		c.PollTime = nacosParams.pollTime
		c.HttpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: nacosParams.timeout,
		}
		c.Logger = &logger{}
	})
}

func DownloadNacosConfig() {
	var err error
	for i := 0; i < 2; i++ {
		if _, err = downloadNacosConfig(); err == nil {
			log.Println("nacos本地配置文件已更新")
			return
		}
		time.Sleep(time.Second)
	}

	log.Fatalf("nacos配置文件下载失败: %s", err.Error())
}

// ListenNacos 监控nacos
func ListenNacos(callbacks ...func(cnf string)) {
	nacosClient.ListenAsync(nacosParams.namespace, nacosParams.group, nacosParams.dataId, func(cnf string) {
		log.Println("nacos监听到远程配置文件有改变，开始获取")

		_, err := downloadNacosConfig()
		if err != nil {
			log.Println("nacos获取远程配置后更新本地配置文件失败", err.Error())
			return
		}

		// 执行callback
		for _, callbackFunc := range callbacks {
			callbackFunc(cnf)
		}
	})
}

func downloadNacosConfig() (string, error) {
	content, err := nacosClient.Get(nacosParams.namespace, nacosParams.group, nacosParams.dataId)
	if err != nil {
		return "", errors.Wrapf(err, "获取nacos远程配置失败")
	}
	if content == "" {
		return "", errors.New("获取的nacos远程配置为空")
	}

	if err := writeFile(DefaultRelationPath, content); err != nil {
		return "", errors.Wrapf(err, "更新本地配置文件失败")
	}

	return content, nil
}

func writeFile(configPath, configContent string) (err error) {
	// 打开配置文件
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return errors.Wrapf(err, "本地配置文件打开失败")
	}

	defer func() {
		_ = file.Close()
	}()

	// 阻塞模式下，加排他锁
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return errors.Wrapf(err, "文件加锁失败")
	}
	defer func() {
		if err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			err = errors.Wrapf(err, "文件解锁失败")
		}
	}()

	// 加载配置信息
	_, err = file.WriteString(configContent)
	if err != nil {
		return errors.Wrapf(err, "写入本地配置文件失败")
	}

	return nil
}
