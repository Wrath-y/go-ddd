package facade

import "go-ddd/domain/article/repository/po"

type ArticleRepositoryI interface {
	GetById(id int64) (po.Article, error)
	GetHitsById(id int64) (po.Article, error)
	HitsIncr(id int64) error
}
