package crand

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ihcsim/kubelet-plugin/pkg/plugins"
	"github.com/rs/zerolog/log"
)

func (p *DevicePlugin) discoverDevices() ([]*plugins.Device, error) {
	var (
		// fullSet keeps track of all available devices. it's used to remove any stale
		// cache entries later.
		fullSet = map[string]*plugins.Device{}

		// changeSet identifies which devices have changed since the last discovery.
		changeSet = []*plugins.Device{}
	)

	f, err := os.Open(hostDevicePath)
	if err != nil {
		return nil, err
	}

	entries, err := f.ReadDir(0)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// skip any non-crand devices
		if !strings.Contains(entry.Name(), "crand") {
			continue
		}

		id := entry.Name()
		device := &plugins.Device{
			ID:     id,
			Health: plugins.Healthy,
		}

		fullSet[id] = device
		lastSeenState, exists := p.cache[id]

		// no change in device's state
		if exists && lastSeenState.Health == device.Health {
			continue
		}

		log := p.log.With().
			Str("device", id).
			Str("health", device.Health.String()).
			Str("path", filepath.Join(hostDevicePath, entry.Name())).
			Logger()

		if !exists {
			log.Info().Msg("found new device")
		} else {
			if lastSeenState.Health != device.Health {
				log.Info().
					Str("before", lastSeenState.Health.String()).
					Time("last seen", time.Unix(lastSeenState.LastSeenTimestamp, 0)).
					Msg("device health changed")
			}
		}

		p.cache[id] = &plugins.DeviceState{
			Device:            device,
			LastSeenTimestamp: time.Now().Unix(),
		}
		changeSet = append(changeSet, device)
	}

	// remove devices that are no longer present from cache
	for id := range p.cache {
		if _, exists := fullSet[id]; !exists {
			log.Info().Msg("removed device from cache")
			delete(p.cache, id)
		}
	}

	return changeSet, nil
}
