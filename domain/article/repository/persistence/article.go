package persistence

import (
	"go-ddd/domain/article/repository/facade"
	"go-ddd/domain/article/repository/po"
	"go-ddd/infrastructure/util/db"
)

type ArticleRepository struct{}

func NewArticleRepository() facade.ArticleRepositoryI {
	return &ArticleRepository{}
}

func (*ArticleRepository) GetById(id int64) (po.Article, error) {
	article := po.Article{}
	return article, db.Orm.First(&article, id).Error
}

func (*ArticleRepository) GetHitsById(id int64) (po.Article, error) {
	article := po.Article{}
	return article, db.Orm.Raw("select id, hits from article where id = ?", id).First(&article).Error
}

func (*ArticleRepository) HitsIncr(id int64) error {
	return db.Orm.Exec("update article set hits = hits + 1 where id = ?", id).Error
}
