package retry

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type RetryAble func() error
type Notify func(error)

var (
	MaxRetryExceeded = errors.New("MaxRetryExceeded")
)

type Retry struct {
	sync.RWMutex
	fn                  RetryAble
	maxRetrys           int
	currentRetry        int
	delayInMilliseconds int
}

func NewRetry(function RetryAble, maxRetrys int, delayInMilliseconds int) *Retry {
	return &Retry{
		fn:                  function,
		maxRetrys:           maxRetrys,
		currentRetry:        0,
		delayInMilliseconds: delayInMilliseconds,
	}
}

func (r *Retry) Execute(ctx context.Context, notify Notify) {
	r.Lock()
	defer r.Unlock()

	for {

		if err := r.fn(); err == nil {
			notify(err)
			return
		}

		r.currentRetry++

		if r.currentRetry == r.maxRetrys {
			notify(MaxRetryExceeded)
			return
		}

		select {
		case <-ctx.Done():
			notify(nil)
			return
		case <-time.After(time.Duration(r.delayInMilliseconds) * time.Millisecond):
			r.delayInMilliseconds = r.delayInMilliseconds + r.delayInMilliseconds
		}
	}
}
