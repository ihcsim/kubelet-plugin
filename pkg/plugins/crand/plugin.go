package crand

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	socketName   = "crand.sock"
	resourceName = "github.com.ihcsim/crand"
)

var (
	_ v1beta1.DevicePluginServer = &DevicePlugin{}

	socketPath = v1beta1.DevicePluginPath + socketName
)

type DevicePlugin struct {
	// cache is used to store the last-seen state of devices on the host.
	// it's updated by the discoverDevices() method.
	cache map[string]*DeviceState

	gserver *grpc.Server
	log     *zerolog.Logger
}

func NewPlugin(log *zerolog.Logger) *DevicePlugin {
	gserver := grpc.NewServer()
	plugin := &DevicePlugin{
		cache:   map[string]*DeviceState{},
		gserver: gserver,
		log:     log,
	}
	v1beta1.RegisterDevicePluginServer(gserver, plugin)

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

	if err := p.grpcServe(ctx, errCh); err != nil {
		return err
	}

	ready, cancel := context.WithTimeout(ctx, grpcReadyTimeoutDuration)
	defer cancel()
	if err := p.grpcReady(ready); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("grpc server ready")

	kubeletAddr := fmt.Sprintf("unix://%s", v1beta1.KubeletSocket)
	if err := p.registerKubelet(ctx, kubeletAddr); err != nil {
		return err
	}
	p.log.Info().Str("addr", socketPath).Msg("plugin registered with kubelet")

	closeHandler, err := p.restartHandler(ctx, kubeletAddr, errCh)
	if err != nil {
		return err
	}
	defer closeHandler()
	p.log.Info().Str("addr", socketPath).Msg("restart handler configured")

	<-ctx.Done()
	return runErr
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

func (p *DevicePlugin) restartHandler(ctx context.Context, kubeletAddr string, errCh chan<- error) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		watcher.Close()
	}

	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()

	go func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				if event.Name == socketPath && event.Has(fsnotify.Remove) {
					p.log.Info().Msg("kubelet restarted, re-registering plugin with kubelet")
					if err := p.grpcServe(ctx, errCh); err != nil {
						errCh <- err
						break LOOP
					}

					if err := p.registerKubelet(ctx, kubeletAddr); err != nil {
						errCh <- err
						break LOOP
					}
					p.log.Info().Msg("re-registration completed successfully")
				}
				<-tick.C
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}

				if err != nil {
					errCh <- err
					break LOOP
				}
				<-tick.C
			}
		}
	}()

	return cleanup, watcher.Add(v1beta1.DevicePluginPath)
}
