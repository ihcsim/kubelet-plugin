package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ihcsim/kvm-device-plugin/pkg/plugins/kvm"
	"github.com/rs/zerolog"
)

func main() {
	var (
		log    = logger()
		plugin = kvm.NewPlugin(log)
	)
	defer func() {
		if err := plugin.Cleanup(); err != nil {
			log.Error().Err(err).Msg("cleanup failed")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info().Msg("received signal, attempting graceful shutdown")
		cancel()
	}()

	if err := plugin.Run(ctx); err != nil {
		if !errors.Is(ctx.Err(), context.Canceled) {
			log.Error().Err(err).Send()
			return
		}
	}

	log.Info().Msg("shutdown completed successfully")
}

func logger() *zerolog.Logger {
	w := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	l := zerolog.New(w).With().Timestamp().Logger()
	return &l
}
