package httpcmd

import (
	"context"
	"log"
	"os"
	"os/signal"
)

func attachInterruptForHttpRunner(ctx context.Context, softInt, hardInt func()) {
	attachInterrupt(ctx, func(nth int) {
		if nth == 0 {
			softInt()
			log.Printf("sigint: terminating after finishing existing requests")
		} else if nth == 1 {
			hardInt()
			log.Printf("sigint: terminating immediately")
		}
	})
}

func attachInterrupt(ctx context.Context, callback func(nth int)) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt)
	go func() {
		defer signal.Stop(ch)
		for times := 0; ; times++ {
			select {
			case <-ch:
				callback(times)
			case <-ctx.Done():
				return
			}
		}
	}()
}
