package sync

import (
	"context"
	"fmt"

	"github.com/yimi-go/logging"
)

type RecoverHandler func(rerr any) error

func Go(ctx context.Context, fn func()) error {
	return WithRecoverGo(fn, func(rerr any) error {
		logger := logging.GetFactory().Logger("sync.go")
		logger = logging.WithContextField(ctx, logger)
		logger.Errorw(fmt.Sprintf("recover from panic: %v", rerr), logging.Stack("recover_at"))
		return nil
	})
}

func WithRecoverGo(fn func(), handler RecoverHandler) (err error) {
	defer func() {
		rerr := recover()
		if rerr == nil {
			return
		}
		err = handler(rerr)
	}()
	fn()
	return
}
