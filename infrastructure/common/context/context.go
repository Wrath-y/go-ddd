package context

import (
	"context"
	"go-ddd/infrastructure/util/logging"
	"time"
)

const Ctx = "__context__"

type Context struct {
	context.Context
	*logging.Logger
}

func NewContext(ctx context.Context) *Context {
	c := &Context{}
	c.Context = context.WithValue(ctx, Ctx, c)

	return c
}

func GetContext(c context.Context) *Context {
	return c.Value(Ctx).(*Context)
}

func (c *Context) GetString(key string) string {
	s, ok := c.Value(key).(string)
	if !ok {
		return ""
	}

	return s
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *Context) Err() error {
	return c.Context.Err()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *Context) Value(key any) any {
	return c.Context.Value(key)
}
