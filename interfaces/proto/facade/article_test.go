package facade

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go-ddd/infrastructure/util/consul"
	"go-ddd/interfaces/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
	"time"
)

func TestFindAll(t *testing.T) {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../../")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	consul.Setup()

	instance, err := consul.Client.GetHealthRandomInstance("article")

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", instance.GetAddress(), instance.GetPort(), grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Error(err.Error())
	}
	defer conn.Close()

	var grpcClient proto.ArticleClient
	grpcClient = proto.NewArticleClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	r, err := grpcClient.GetById(ctx, &proto.GetByIdReq{
		Id: 27,
	})
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(r)
}
