package event

import "go-ddd/infrastructure/common/context"

type PublisherI interface {
	AddFunc(f func(*context.Context)) PublisherI
	Publish(ctx *context.Context) error
}
