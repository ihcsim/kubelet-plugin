package kvm

import (
	"os"
	"path/filepath"
	"time"

	"github.com/ihcsim/kubelet-plugin/pkg/plugins"
)

func (p *DevicePlugin) discoverDevices() (bool, error) {
	id := resourceName
	device := &plugins.Device{
		ID:     id,
		Health: plugins.Healthy,
	}
	hasChanged := false

	// always update cache
	defer func() {
		p.cache = &plugins.DeviceState{
			LastSeenTimestamp: time.Now().Unix(),
			Device:            device,
		}
	}()

	if _, err := os.Open(hostDevicePath); err != nil {
		device.Health = plugins.Unhealthy
		hasChanged = true
		return hasChanged, err
	}

	log := p.log.With().
		Str("device", id).
		Str("health", device.Health.String()).
		Str("path", filepath.Join(hostDevicePath, device.ID)).
		Logger()

	// add new device's state to cache
	if p.cache == nil {
		log.Info().Msg("found new device")
		hasChanged = true
		return hasChanged, nil
	}

	lastSeenState := p.cache
	if lastSeenState.Health == device.Health {
		log.Info().
			Str("before", lastSeenState.Health.String()).
			Time("last seen", time.Unix(lastSeenState.LastSeenTimestamp, 0)).
			Msg("device health changed")
		hasChanged = true
	}

	return hasChanged, nil
}
