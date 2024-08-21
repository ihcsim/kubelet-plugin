package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ihcsim/kubelet-plugin/pkg/plugins/generic"
	"github.com/rs/zerolog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var socket = v1beta1.DevicePluginPath + "generic.sock"

func main() {
	var (
		log         = logger()
		plugin      = generic.NewPlugin(socket, log)
		ctx, cancel = context.WithCancel(context.Background())
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info().Msg("received signal, attempting graceful shutdown")
		cancel()
	}()

	if err := plugin.Run(ctx); err != nil {
		log.Error().Err(err).Send()
		return
	}

	log.Info().Msg("shutdown completed successfully")
}

func logger() *zerolog.Logger {
	w := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	l := zerolog.New(w).With().Timestamp().Logger()
	return &l
}
