package facade

import (
	"context"
	"go-ddd/interfaces/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestFindAll(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Error(err.Error())
	}
	defer conn.Close()

	var grpcClient proto.ArticleClient
	grpcClient = proto.NewArticleClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	r, err := grpcClient.FindById(ctx, &proto.FindByIdReq{
		Id:   0,
		Size: 1,
	})
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(r)
}
