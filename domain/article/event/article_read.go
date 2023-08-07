package event

import (
	"go-ddd/infrastructure/common/context"
)

func ArticleRead(fList ...func() error) func(ctx *context.Context) {
	return func(ctx *context.Context) {
		for _, f := range fList {
			if err := f(); err != nil {
				ctx.Logger.ErrorL("增加文章访问次数失败", "", err.Error())
			}
		}
	}
}
