package rabbitmq

import (
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"log"
)

var Conn *amqp.Connection

func Setup() {
	Conn = getConnection("default")
}

func getConnection(store string) *amqp.Connection {
	dbViper := viper.Sub("rabbitmq." + store)
	if dbViper == nil {
		log.Fatal("rabbitmq配置缺失", store)
	}

	address := dbViper.GetString("address")
	username := dbViper.GetString("username")
	password := dbViper.GetString("password")

	url := "amqp://" + username + ":" + password + "@" + address + "/"

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal("amqp连接失败", err)
	}

	return conn
}
