package event

import (
	"encoding/json"
	"go-ddd/infrastructure/common/context"
	"time"
)

type Base struct {
	*context.Context
	f          func(*context.Context)
	Id         int       `json:"id"`
	Source     string    `json:"source"`
	Data       string    `json:"data"`
	CreateTime time.Time `json:"create_time"`
}

func NewBasePublisher() PublisherI {
	return new(Base)
}

func (b *Base) AddFunc(f func(*context.Context)) PublisherI {
	b.f = f
	return b
}

func (b *Base) Publish(ctx *context.Context) error {
	if b.f != nil {
		return b.handleFunc(ctx)
	}

	return b.sendToMQ(ctx)
}

func (b *Base) handleFunc(ctx *context.Context) error {
	go b.f(ctx)

	return nil
}

func (b *Base) sendToMQ(ctx *context.Context) error {
	// send to mq
	bt, err := json.Marshal(b)
	if err != nil {
		return err
	}
	println(bt)

	return nil
}
