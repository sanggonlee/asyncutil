package asyncutil

import (
	"context"
	"sync"
)

// Collect takes error channels and returns a new error channel where all non-nil
// errors from input error channels are funneled into.
func Collect(errchans ...<-chan error) <-chan error {
	return wait(context.TODO(), errchans)
}

// CollectContext is same as Collect, except it takes a context.
// If the context exceeds deadline or is cancelled, the resulting error channel
// receives the corresponding error, in addition to the errors collected by errchans.
// If the context was already cancelled by the time CollectContext is executed, the
// resulting error channel will only contain the context error, and not the errors
// collected by errchans.
func CollectContext(ctx context.Context, errchans ...<-chan error) <-chan error {
	return wait(ctx, errchans)
}

func wait(ctx context.Context, errchans []<-chan error) <-chan error {
	errs := make(chan error)
	var wg sync.WaitGroup
	closeChanEventually := func() {
		wg.Wait()
		close(errs)
	}

	if !isEmptyContext(ctx) && len(errchans) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			done := ctx.Done()
			if done == nil {
				return
			}

			<-done
			errs <- ctx.Err()
		}()

		if ctx.Err() != nil {
			go closeChanEventually()
			// Context deadline has already passed, return without waiting for errchans
			return errs
		}
	}

	for _, errchan := range errchans {
		wg.Add(1)
		go func(errch <-chan error) {
			defer wg.Done()

			if errch == nil {
				return
			}

			for err := range errch {
				if err != nil {
					errs <- err
				}
			}
		}(errchan)
	}

	go closeChanEventually()

	return errs
}

func isEmptyContext(ctx context.Context) bool {
	return ctx == context.TODO()
}
