package service

import (
	"github.com/go-redis/redis/v7"
	"go-ddd/domain/article/entity"
	"go-ddd/domain/article/event"
	"go-ddd/domain/article/repository/facade"
	"go-ddd/domain/article/repository/persistence"
	"go-ddd/domain/article/repository/po"
	"go-ddd/infrastructure/common/context"
	baseEvent "go-ddd/infrastructure/common/event"
)

type ArticleDomainService struct {
	*context.Context
	articleFactory     ArticleFactory
	articleCache       ArticleCache
	articleRepositoryI facade.ArticleRepositoryI
	publisherI         baseEvent.PublisherI
}

func NewArticleDomainService(ctx *context.Context) ArticleDomainService {
	return ArticleDomainService{
		Context:            ctx,
		articleFactory:     NewArticleFactory(),
		articleCache:       NewArticleCache(),
		articleRepositoryI: persistence.NewArticleRepository(),
		publisherI:         baseEvent.NewBasePublisher(),
	}
}

func (a *ArticleDomainService) GetById(id int64) (entity.Article, error) {
	defer func() {
		if err := a.publisherI.AddFunc(event.ArticleRead(
			func() error {
				return a.articleRepositoryI.HitsIncr(id)
			},
			func() error {
				return a.articleCache.HitsIncr(id)
			},
		)).Publish(a.Context); err != nil {
			a.Logger.ErrorL("发布ArticleRead事件失败", id, err.Error())
		}
	}()

	var err error
	article := po.Article{}

	article, err = a.articleCache.GetDetail(id)
	if err != nil && err != redis.Nil {
		a.Logger.ErrorL("获取文章详情缓存失败", id, err.Error())
		return entity.Article{}, err
	}
	if err == nil {
		return a.articleFactory.CreateArticleEntity(article), nil
	}

	article, err = a.articleRepositoryI.GetById(id)
	if err != nil {
		a.Logger.ErrorL("获取文章详情失败", id, err.Error())
		return entity.Article{}, err
	}

	article.Hits++
	if err := a.articleCache.SetDetail(id, article); err != nil {
		a.Logger.ErrorL("缓存文章详情失败", id, err.Error())
	}

	return a.articleFactory.CreateArticleEntity(article), nil
}
