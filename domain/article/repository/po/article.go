package po

import (
	"time"
)

type Article struct {
	Id         int64     `json:"id"`
	Title      string    `json:"title"`
	Image      string    `json:"image"`
	Intro      string    `json:"intro"`
	Html       string    `json:"html"`
	Con        string    `json:"con"`
	Hits       int       `json:"hits"`
	Status     int8      `json:"status"`
	Source     int       `json:"source"`
	Tags       string    `json:"tags"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*Article) TableName() string {
	return "article"
}
