package service

import (
	"go-ddd/domain/article/entity"
	"go-ddd/domain/article/repository/po"
)

type ArticleFactory struct {
}

func NewArticleFactory() ArticleFactory {
	return ArticleFactory{}
}

func (*ArticleFactory) CreateArticlePO(article entity.Article) po.Article {
	return po.Article{
		Id:         article.Id,
		Title:      article.Title,
		Image:      article.Image,
		Intro:      article.Intro,
		Html:       article.Html,
		Con:        article.Con,
		Hits:       article.Hits,
		Status:     article.Status,
		Source:     article.Source,
		Tags:       article.Tags,
		CreateTime: article.CreateTime,
		UpdateTime: article.UpdateTime,
	}
}

func (a *ArticleFactory) CreateArticleEntities(poList []*po.Article) []*entity.Article {
	res := make([]*entity.Article, 0, len(poList))
	for _, v := range poList {
		tmp := a.CreateArticleEntity(*v)
		res = append(res, &tmp)
	}

	return res
}

func (*ArticleFactory) CreateArticleEntity(article po.Article) entity.Article {
	return entity.Article{
		Id:         article.Id,
		Title:      article.Title,
		Image:      article.Image,
		Intro:      article.Intro,
		Html:       article.Html,
		Con:        article.Con,
		Hits:       article.Hits,
		Status:     article.Status,
		Source:     article.Source,
		Tags:       article.Tags,
		CreateTime: article.CreateTime,
		UpdateTime: article.UpdateTime,
	}
}
