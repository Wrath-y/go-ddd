package service

import (
	"go-ddd/domain/article/entity"
	"go-ddd/domain/article/service"
	"go-ddd/infrastructure/common/context"
)

type ArticleApplicationService struct {
	*context.Context
	articleDomainService service.ArticleDomainService
}

func NewArticleApplicationService(ctx *context.Context) *ArticleApplicationService {
	return &ArticleApplicationService{
		ctx,
		service.NewArticleDomainService(ctx),
	}
}

func (a *ArticleApplicationService) GetById(id int64) (entity.Article, error) {
	return a.articleDomainService.GetById(id)
}
