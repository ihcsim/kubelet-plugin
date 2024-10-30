package kvm

import (
	"context"
	"errors"
	"fmt"

	"github.com/ihcsim/kvm-device-plugin/pkg/plugins"
	"github.com/rs/zerolog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	socketName   = "kvm.sock"
	resourceName = "github.com.ihcsim/kvm"
)

var (
	_          v1beta1.DevicePluginServer = &DevicePlugin{}
	socketPath                            = v1beta1.DevicePluginPath + socketName
)

type DevicePlugin struct {
	cache  *plugins.DeviceState
	server *plugins.Server
	log    *zerolog.Logger
}

func NewPlugin(log *zerolog.Logger) *DevicePlugin {
	server := plugins.NewServer(socketPath)
	plugin := &DevicePlugin{
		server: server,
		log:    log,
	}

	v1beta1.RegisterDevicePluginServer(server.GRPC, plugin)

	plugin.log.Info().Msg("plugin initialized")
	return plugin
}

func (p *DevicePlugin) Run(ctx context.Context) error {
	// errCh is used to collect errors from goroutines and
	// handle them in Run().
	var (
		runErr error
		errCh  = make(chan error)
	)
	defer close(errCh)
	go func() {
		for err := range errCh {
			if err != nil {
				p.log.Err(err).Msg("error reported by goroutine")
				runErr = errors.Join(runErr, err)
			}
		}
	}()

	if err := p.server.GRPCServe(ctx, errCh); err != nil {
		return err
	}

	if err := p.server.GRPCReady(ctx); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("grpc server ready")

	kubeletAddr := fmt.Sprintf("unix://%s", v1beta1.KubeletSocket)
	if err := plugins.RegisterWithKubelet(ctx, socketName, resourceName, kubeletAddr); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("plugin registered with kubelet")

	restart := func() error {
		if err := p.server.GRPCServe(ctx, errCh); err != nil {
			return err
		}

		return plugins.RegisterWithKubelet(ctx, socketName, resourceName, kubeletAddr)
	}

	closeHandler, err := plugins.RegisterWithRestartHandler(ctx, restart, socketPath, p.log, errCh)
	if err != nil {
		return err
	}
	defer closeHandler()
	p.log.Info().Str("addr", socketPath).Msg("restart handler configured")

	<-ctx.Done()
	return runErr
}
