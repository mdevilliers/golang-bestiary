package retry

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"golang.org/x/net/context"
)

func TestRetry(t *testing.T) {

	cnt := 1

	retryAble := func() error {

		if cnt == 5 {

			return nil
		}
		cnt++

		return errors.New("boom")

	}

	var notifee = func(err error) {
		if err != nil {
			t.Error("Error returned")
		}
	}

	retry := NewRetry(retryAble, 5, 1)
	retry.Execute(context.Background(), notifee)

}

func TestRetryCancel(t *testing.T) {

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)

	retryAble := func() error {

		fmt.Println("TestRetryCancel")
		cancel()
		wg.Done()

		//
		return errors.New("!")
	}

	var notifee = func(err error) {
		if err != nil {
			t.Error("Error returned")
		}
	}

	retry := NewRetry(retryAble, 5, 1)
	go retry.Execute(ctx, notifee)

	wg.Wait()

	if ctx.Err() != context.Canceled {
		t.Error("Should have been cancelled!")
	}

}
