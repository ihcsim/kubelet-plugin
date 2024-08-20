package generic

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ v1beta1.DevicePluginServer = &DevicePlugin{}

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
	errCh := make(chan error)
	go func() {
		errCh <- p.serve()
	}()

	go func() {
		<-ctx.Done()
		p.gserver.GracefulStop()
	}()

	return <-errCh
}
