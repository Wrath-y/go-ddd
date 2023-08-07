package facade

import (
	"context"
	"go-ddd/application/service"
	grpcCtx "go-ddd/infrastructure/common/context"
	"go-ddd/infrastructure/common/errcode"
	"go-ddd/interfaces/assembler"
	"go-ddd/interfaces/proto"
	"go-ddd/launch/grpc/resp"
)

type Article struct{}

func (*Article) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.Response, error) {
	res, err := service.NewArticleApplicationService(grpcCtx.GetContext(ctx)).GetById(req.Id)
	if err != nil {
		return resp.FailWithErrCode(errcode.ArticleNotExists)
	}
	return resp.Success(assembler.ToArticleDTO(res))
}
