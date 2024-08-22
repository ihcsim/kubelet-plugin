package pflex

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	socketName   = "pflex.sock"
	resourceName = "pflex.io/block"
)

var (
	_ v1beta1.DevicePluginServer = &DevicePlugin{}

	socketPath = v1beta1.DevicePluginPath + socketName
)

type DevicePlugin struct {
	gserver *grpc.Server
	log     *zerolog.Logger
}

func NewPlugin(log *zerolog.Logger) *DevicePlugin {
	gserver := grpc.NewServer()
	plugin := &DevicePlugin{
		gserver: gserver,
		log:     log,
	}

	v1beta1.RegisterDevicePluginServer(gserver, plugin)
	return plugin
}

func (p *DevicePlugin) Run(ctx context.Context) error {
	p.log.Info().Str("addr", socketPath).Msg("starting grpc server")
	if err := p.grpcServe(ctx); err != nil {
		return err
	}

	ready, cancel := context.WithTimeout(ctx, grpcReadyTimeoutDuration)
	defer cancel()
	if err := p.grpcReady(ready); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("grpc server ready")

	p.log.Info().Str("addr", socketPath).Msg("registering with kubelet")
	kubeletAddr := fmt.Sprintf("unix://%s", v1beta1.KubeletSocket)
	if err := p.registerKubelet(ctx, kubeletAddr); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("registration completed successfully")

	return nil
}

func (p *DevicePlugin) registerKubelet(ctx context.Context, addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := v1beta1.NewRegistrationClient(conn)
	request := &v1beta1.RegisterRequest{
		Version:      v1beta1.Version,
		Endpoint:     socketName,
		ResourceName: resourceName,
	}

	if _, err := client.Register(ctx, request); err != nil {
		return err
	}

	return nil
}
