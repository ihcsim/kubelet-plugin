package generic

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ v1beta1.DevicePluginServer = &DevicePlugin{}

var (
	readyTimeoutDuration = 10 * time.Second
)

type DevicePlugin struct {
	gserver   *grpc.Server
	log       *zerolog.Logger
	pluginDir string
	socket    string
}

func NewPlugin(pluginDir, socket string, log *zerolog.Logger) *DevicePlugin {
	gserver := grpc.NewServer()
	plugin := &DevicePlugin{
		gserver:   gserver,
		log:       log,
		pluginDir: pluginDir,
		socket:    socket,
	}

	v1beta1.RegisterDevicePluginServer(gserver, plugin)

	plugin.log.Info().Str("socket", socket).Msg("plugin created")
	return plugin
}

func (p *DevicePlugin) Run(ctx context.Context) error {
	grpcErr := make(chan error)
	if err := p.startGRPC(ctx, grpcErr); err != nil {
		return err
	}

	for err := range grpcErr {
		return err
	}

	return nil
}

func (p *DevicePlugin) startGRPC(ctx context.Context, errCh chan<- error) error {
	ready, cancel := context.WithTimeout(ctx, readyTimeoutDuration)
	defer func() {
		close(errCh)
		cancel()
	}()

	go func() {
		errCh <- p.serve()
	}()

	go func() {
		<-ctx.Done()
		p.gracefulStop()
	}()

	return p.grpcReady(ready)
}
