package goredis

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"log"
)

var Client *redis.Client

func Setup() {
	Client = newClient("default")
}

func newClient(store string) *redis.Client {
	cfg := viper.Sub("redis." + store)
	if cfg == nil {
		log.Fatal("redis配置缺失", store)
	}

	var tlsConfig *tls.Config
	cert := cfg.GetString("cert")
	key := cfg.GetString("key")
	ca := cfg.GetString("ca")
	if cert != "" && key != "" && ca != "" {
		tlsConfig = &tls.Config{}
		certificate, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			log.Fatal(err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(ca)) {
			log.Fatal("failed to parse root certificate")
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
		tlsConfig.RootCAs = pool
	}

	var cli *redis.Client
	if cfg.GetString("address") != "" {
		cli = redis.NewClient(&redis.Options{
			Addr:         cfg.GetString("address"),
			Password:     cfg.GetString("password"),
			DB:           cfg.GetInt("db"),
			TLSConfig:    tlsConfig,
			PoolSize:     cfg.GetInt("pool_size"),
			MinIdleConns: cfg.GetInt("min_idle_conns"),
			MaxRetries:   2,
		})
	} else if cfg.GetString("mastername") != "" {
		cli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.GetString("mastername"),
			Username:         cfg.GetString("username"),
			Password:         cfg.GetString("password"),
			DB:               cfg.GetInt("db"),
			SentinelAddrs:    cfg.GetStringSlice("sentineladdrs"),
			SentinelUsername: cfg.GetString("sentinelusername"),
			SentinelPassword: cfg.GetString("sentinelpassword"),
			TLSConfig:        tlsConfig,
			PoolSize:         cfg.GetInt("pool_size"),
			MinIdleConns:     cfg.GetInt("min_idle_conns"),
			MaxRetries:       2,
		})
	} else {
		log.Fatal("redis配置不全", store)
	}

	if err := cli.Ping().Err(); err != nil {
		log.Fatal(err)
	}
	return cli
}
