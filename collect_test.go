package asyncutil_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sanggonlee/asyncutil"
)

////////////////////////////////////////////////////////////////////////////
// Examples.
////////////////////////////////////////////////////////////////////////////

func ExampleCollect() {
	doWork := func(url string) chan error {
		errs := make(chan error)
		go func(url string) {
			// Some expensive work..
			_, err := http.Get(url)
			if err != nil {
				errs <- err
			}
		}(url)
		return errs
	}

	for err := range asyncutil.Collect(
		doWork("1"),
		doWork("2"),
	) {
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
// Tests.
////////////////////////////////////////////////////////////////////////////

func TestCollect_NoFunctions(t *testing.T) {
	errs := asyncutil.Collect()
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
		}
	}
	if numErrors != 0 {
		t.Errorf("Expected no errors but got %d", numErrors)
	}
}

func TestCollect_OneOfTheFunctionsReturnsError(t *testing.T) {
	errs := asyncutil.Collect(
		func() chan error {
			errch := make(chan error)
			go func() {
				close(errch)
			}()
			return errch
		}(),
		func() chan error {
			errch := make(chan error)
			go func() {
				errch <- errors.New("err")
				close(errch)
			}()
			return errch
		}(),
		func() chan error {
			errch := make(chan error)
			go func() {
				close(errch)
			}()
			return errch
		}(),
	)
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
			if err.Error() != "err" {
				t.Errorf("Expected error %s but got %s", "err", err.Error())
			}
		}
	}
	if numErrors != 1 {
		t.Fatalf("Expected %d errors but got %d", 1, numErrors)
	}
}

func TestCollect_MultipleFunctionsReturnErrors(t *testing.T) {
	errs := asyncutil.Collect(
		func() chan error {
			errch := make(chan error)
			go func() {
				errch <- errors.New("err1")
				close(errch)
			}()
			return errch
		}(),
		func() chan error {
			errch := make(chan error)
			go func() {
				errch <- errors.New("err2")
				close(errch)
			}()
			return errch
		}(),
		func() chan error {
			errch := make(chan error)
			go func() {
				errch <- errors.New("err3")
				close(errch)
			}()
			return errch
		}(),
	)
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
		}
	}
	if numErrors != 3 {
		t.Fatalf("Expected %d errors but got %d", 3, numErrors)
	}
}

func TestCollect_ErrorChannelIsClosedAlready(t *testing.T) {
	errs := asyncutil.Collect(
		func() chan error {
			errch := make(chan error)
			close(errch)
			return errch
		}(),
		func() chan error {
			errch := make(chan error)
			go func() {
				errch <- errors.New("err2")
				close(errch)
			}()
			return errch
		}(),
	)
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
		}
	}
	if numErrors != 1 {
		t.Fatalf("Expected %d errors but got %d", 1, numErrors)
	}
}

func TestCollectContext_NoFunctionsAndNoDeadline(t *testing.T) {
	errs := asyncutil.CollectContext(context.Background())
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
		}
	}
	if numErrors != 0 {
		t.Errorf("Expected no errors but got %d", numErrors)
	}
}

func TestCollectContext_NoFunctionsWithDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()
	errs := asyncutil.CollectContext(ctx)
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++
		}
	}
	if numErrors != 0 {
		t.Errorf("Expected no errors but got %d", numErrors)
	}
}

func TestCollectContext_FunctionExceedsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100*time.Millisecond))
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()
	errs := asyncutil.CollectContext(ctx,
		func() chan error {
			errch := make(chan error)
			defer close(errch)
			time.Sleep(1 * time.Second)
			return errch
		}(),
	)
	var numErrors int
	for err := range errs {
		if err != nil {
			numErrors++

			if err != context.DeadlineExceeded {
				t.Errorf("Expected deadline exceeded error but got %s", err.Error())
			}
		}
	}
	if numErrors != 1 {
		t.Errorf("Expected no errors but got %d", numErrors)
	}
}

func TestCollectContext_ContextCancelledBeforeAnyFunctionReturns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	errs := asyncutil.CollectContext(ctx,
		func(ctx context.Context) chan error {
			errch := make(chan error)
			go func() {
				time.Sleep(1 * time.Second)
				errch <- errors.New("err1")
				close(errch)
			}()
			return errch
		}(ctx),
		func(ctx context.Context) chan error {
			errch := make(chan error)
			go func() {
				time.Sleep(1 * time.Second)
				errch <- errors.New("err2")
				close(errch)
			}()
			return errch
		}(ctx),
		func(ctx context.Context) chan error {
			errch := make(chan error)
			go func() {
				time.Sleep(1 * time.Second)
				errch <- errors.New("err3")
				close(errch)
			}()
			return errch
		}(ctx),
	)
	var numErrors, numCancelledError int
	for err := range errs {
		if err != nil {
			numErrors++

			if err == context.Canceled {
				numCancelledError++
			}
		}
	}
	if numCancelledError != 1 {
		t.Errorf("Expected %d context cancelled but got %d", 1, numCancelledError)
	}
	if numErrors != 4 {
		t.Fatalf("Expected %d errors but got %d", 4, numErrors)
	}
}

func TestCollectContext_CancelledContextIsPassed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	cancel()

	errs := asyncutil.CollectContext(ctx,
		func(ctx context.Context) chan error {
			errch := make(chan error)
			go func() {
				time.Sleep(1 * time.Second)
				errch <- errors.New("err1")
				close(errch)
			}()
			return errch
		}(ctx),
		func(ctx context.Context) chan error {
			errch := make(chan error)
			go func() {
				time.Sleep(1 * time.Second)
				errch <- errors.New("err2")
				close(errch)
			}()
			return errch
		}(ctx),
	)
	var numErrors, numCancelledError int
	for err := range errs {
		if err != nil {
			numErrors++

			if err == context.Canceled {
				numCancelledError++

			}
		}
	}
	if numCancelledError != 1 {
		t.Errorf("Expected %d context cancelled but got %d", 1, numCancelledError)
	}
	if numErrors != 1 {
		t.Fatalf("Expected %d errors but got %d", 1, numErrors)
	}
}

////////////////////////////////////////////////////////////////////////////
// Benchmarks.
////////////////////////////////////////////////////////////////////////////

func BenchmarkCollect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		results := make([]<-chan error, 0, len(durations))
		for _, dur := range durations {
			results = append(results, work(dur))
		}
		for err := range asyncutil.Collect(results...) {
			if err != nil {
				b.Fatal(err.Error())
			}
		}
	}
}

func BenchmarkSequential(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, dur := range durations {
			for err := range work(dur) {
				if err != nil {
					b.Fatal(err.Error())
				}
			}
		}
	}
}

func work(dur time.Duration) chan error {
	errch := make(chan error)
	go func() {
		defer close(errch)
		time.Sleep(dur)
		errch <- nil
	}()
	return errch
}

var durations = []time.Duration{
	50 * time.Millisecond,
	50 * time.Millisecond,
	50 * time.Millisecond,
	50 * time.Millisecond,
	50 * time.Millisecond,
}
