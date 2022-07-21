package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yimi-go/logging"
	"github.com/yimi-go/logging/hook"
)

func TestWithRecoverGo(t *testing.T) {
	t.Run("no_panic", func(t *testing.T) {
		ch := make(chan struct{}, 10)
		err := WithRecoverGo(func() {}, func(rerr any) error {
			ch <- struct{}{}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
		select {
		case <-ch:
			t.Errorf("want noting, got one")
		case <-time.After(time.Millisecond):
		}
	})
	t.Run("handle_ok", func(t *testing.T) {
		ch := make(chan struct{}, 10)
		err := WithRecoverGo(func() {
			panic("abc")
		}, func(rerr any) error {
			ch <- struct{}{}
			return nil
		})
		select {
		case <-ch:
		case <-time.After(time.Millisecond):
			t.Errorf("timeout")
		}
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})
	t.Run("handle_err", func(t *testing.T) {
		ch := make(chan struct{}, 10)
		err := WithRecoverGo(func() {
			panic("abc")
		}, func(rerr any) error {
			ch <- struct{}{}
			return fmt.Errorf("%v", rerr)
		})
		select {
		case <-ch:
		case <-time.After(time.Millisecond):
			t.Errorf("timeout")
		}
		if err == nil {
			t.Errorf("expect err, got nil")
		}
	})
}

type record struct {
	meth   string
	params []any
}

func TestGo(t *testing.T) {
	ch := make(chan *record, 10)
	originFactory := logging.SwapFactory(hook.Hooked(logging.NewNopLoggerFactory(), func(meth string, param ...any) {
		ch <- &record{meth, param}
	}))
	defer func() {
		logging.SwapFactory(originFactory)
	}()
	err := Go(context.Background(), func() {
		panic("abc")
	})
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	select {
	case r := <-ch:
		if r.meth != "Errorw" {
			t.Errorf("expect Errorf, got %v", r.meth)
		}
		t.Logf("%s, %+v", r.params[0].(string), r.params[1].([]logging.Field))
	case <-time.After(time.Millisecond):
		t.Errorf("timeout")
	}
}
