package service

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"go-ddd/domain/article/repository/po"
	"go-ddd/infrastructure/util/def"
	"go-ddd/infrastructure/util/goredis"
	"go-ddd/infrastructure/util/util/highperf"
	"strconv"
	"time"
)

const (
	ListStrKey    = "blog:article:list:%d:%d"
	DetailHashKey = "blog:article:%d"
)

type ArticleCache struct {
}

func NewArticleCache() ArticleCache {
	return ArticleCache{}
}

func (*ArticleCache) GetList(id int64, size int32) ([]*po.Article, error) {
	articles := make([]*po.Article, 0)
	b, err := goredis.Client.Get(fmt.Sprintf(ListStrKey, id, size)).Bytes()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (*ArticleCache) SetList(id int64, size int32, list []*po.Article) error {
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return goredis.Client.Set(fmt.Sprintf(ListStrKey, id, size), highperf.Bytes2str(b), time.Hour*24*7).Err()
}

func (*ArticleCache) GetDetail(id int64) (po.Article, error) {
	article := new(po.Article)
	m, err := goredis.Client.HGetAll(fmt.Sprintf(DetailHashKey, id)).Result()
	if err != nil {
		return *article, err
	}
	if len(m) == 0 {
		return *article, redis.Nil
	}

	article.Id, err = strconv.ParseInt(m["id"], 10, 64)
	article.Title = m["title"]
	article.Image = m["image"]
	article.Intro = m["intro"]
	article.Html = m["html"]
	article.Con = m["con"]
	article.Hits, err = strconv.Atoi(m["hits"])
	article.Tags = m["tags"]
	status, err := strconv.Atoi(m["status"])
	article.Status = int8(status)
	article.Source, err = strconv.Atoi(m["source"])
	article.CreateTime, err = time.Parse(def.ISO8601Layout, m["create_time"])
	article.UpdateTime, err = time.Parse(def.ISO8601Layout, m["update_time"])

	return *article, nil
}

func (*ArticleCache) SetDetail(id int64, detail po.Article) error {
	b, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	m := make(map[string]any)
	if err = json.Unmarshal(b, &m); err != nil {
		return err
	}

	key := fmt.Sprintf(DetailHashKey, id)
	if err := goredis.Client.HSet(key, m).Err(); err != nil {
		return err
	}

	return goredis.Client.Expire(key, time.Hour*24*7).Err()
}

func (*ArticleCache) HitsIncr(id int64) error {
	key := fmt.Sprintf(DetailHashKey, id)
	if err := goredis.Client.HIncrBy(key, "hits", 1).Err(); err != nil {
		return err
	}

	return goredis.Client.Expire(key, time.Hour*24*7).Err()
}
