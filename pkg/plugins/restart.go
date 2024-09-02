package plugins

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// RegisterWithRestartHandler watches for the kubelet socket file and re-registers
// the plugin with the kubelet during a restart.
// See https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#handling-kubelet-restarts
func RegisterWithRestartHandler(
	ctx context.Context,
	restart func() error,
	socketPath string,
	log *zerolog.Logger,
	errCh chan<- error) (func(), error) {
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
					log.Info().Msg("kubelet restarted, re-registering plugin with kubelet")
					if err := restart(); err != nil {
						errCh <- err
						break LOOP
					}
					log.Info().Msg("re-registration completed successfully")
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
