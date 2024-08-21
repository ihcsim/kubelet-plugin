package generic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var (
	_ v1beta1.DevicePluginServer = &DevicePlugin{}

	readyTimeoutDuration = 10 * time.Second
)

type DevicePlugin struct {
	gserver *grpc.Server
	log     *zerolog.Logger
	socket  string
}

func NewPlugin(socket string, log *zerolog.Logger) *DevicePlugin {
	gserver := grpc.NewServer()
	plugin := &DevicePlugin{
		gserver: gserver,
		log:     log,
		socket:  socket,
	}

	v1beta1.RegisterDevicePluginServer(gserver, plugin)

	return plugin
}

func (p *DevicePlugin) Run(ctx context.Context) error {
	var (
		errCh = make(chan error)
		errs  error
	)

	go func() {
		for err := range errCh {
			errs = errors.Join(errs, err)
		}
	}()

	p.log.Info().Str("addr", p.socket).Msg("starting grpc server1")
	p.grpcStartServe(ctx, errCh)

	ready, cancel := context.WithTimeout(ctx, readyTimeoutDuration)
	defer cancel()
	if err := p.grpcReady(ready); err != nil {
		return err
	}
	p.log.Info().Str("addr", p.socket).Msg("grpc server ready")

	kubeletAddr := fmt.Sprintf("unix://%s", v1beta1.KubeletSocket)
	if err := p.registerKubelet(kubeletAddr); err != nil {
		return err
	}

	return errs
}

func (p *DevicePlugin) registerKubelet(addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := v1beta1.NewRegistrationClient(conn)
	request := &v1beta1.RegisterRequest{
		Version:      v1beta1.Version,
		Endpoint:     p.socket,
		ResourceName: "generic",
	}

	if _, err := client.Register(context.Background(), request); err != nil {
		return err
	}

	return nil
}
