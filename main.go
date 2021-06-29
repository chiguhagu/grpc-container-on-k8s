package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.TODO()

	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	s := &http.Server{
		Addr:    ":19090",
		Handler: m,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	fmt.Println("start OpenMetrics server")
	wg.Go(func() error {
		return s.ListenAndServe()
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
	select {
	case <-sigCh:
		fmt.Println("catch os.Interrupt")
	case <-ctx.Done():
	}

	cancel()

	fmt.Println("start to shutdown OpenMetrics server gracefully")
	if err := s.Shutdown(context.TODO()); err != nil {
		fmt.Println("failed to shutdown OpenMetrics server gracefully", err)
	}
	fmt.Println("finish to shutdown OpenMetrics server gracefully")

	if err := wg.Wait(); err != nil {
		os.Exit(1)
	}
}
