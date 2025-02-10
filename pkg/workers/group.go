package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Worker interface {
	Name() string
	Start(context.Context) error
}

type Group []Worker

func (g Group) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	startCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	var wg sync.WaitGroup
	errCh := make(chan error, len(g))
	wg.Add(len(g))
	for _, s := range g {
		go func(s Worker) {
			defer wg.Done()
			if err := s.Start(startCtx); err != nil {
				errCh <- fmt.Errorf("%s: %v", s.Name(), err)
				cancelFn()
			}
		}(s)
	}

	<-startCtx.Done()
	wg.Wait()

	var err error
	close(errCh)
	for srvErr := range errCh {
		err = multierror.Append(err, srvErr)
	}
	return err
}
