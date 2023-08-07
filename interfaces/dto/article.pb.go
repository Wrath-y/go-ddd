package dto

import "time"

type H struct{}

type ArticlesItem struct {
	Id         int64     `json:"id"`
	Title      string    `json:"title"`
	Image      string    `json:"image"`
	Intro      string    `json:"intro"`
	Hits       int       `json:"hits"`
	Source     int       `json:"source"`
	Tags       string    `json:"tags"`
	CreateTime time.Time `json:"create_time"`
}

type Article struct {
	Id         int64     `json:"id"`
	Title      string    `json:"title"`
	Image      string    `json:"image"`
	Html       string    `json:"html"`
	Hits       int       `json:"hits"`
	Source     int       `json:"source"`
	Tags       string    `json:"tags"`
	CreateTime time.Time `json:"create_time"`
}
