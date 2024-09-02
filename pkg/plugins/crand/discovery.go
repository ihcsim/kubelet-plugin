package crand

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func (p *DevicePlugin) discoverDevices() (map[string]*Device, []*Device, error) {
	var (
		fullSet   = map[string]*Device{}
		changeSet = []*Device{}
	)

	f, err := os.Open(hostDevicePath)
	if err != nil {
		return nil, nil, err
	}

	entries, err := f.ReadDir(0)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.Contains(entry.Name(), "crand") {
			id := entry.Name()
			device := &Device{
				ID:     id,
				Health: Healthy,
			}

			fullSet[id] = device
			lastSeenState, exists := p.cache[id]

			log := p.log.With().
				Str("device", id).
				Str("health", device.Health.String()).
				Str("path", filepath.Join(hostDevicePath, entry.Name())).
				Logger()

			if exists && lastSeenState.Health == device.Health {
				continue
			}
			changeSet = append(changeSet, device)

			// add new device's state to cache
			if !exists {
				log.Info().Msg("found new device")
				p.cache[id] = &DeviceState{
					Device:            device,
					lastSeenTimestamp: time.Now().Unix(),
				}
				continue
			}

			// update existing device's state in cache, if changed
			if lastSeenState.Health != device.Health {
				log.Info().
					Str("before", lastSeenState.Health.String()).
					Time("last seen", time.Unix(lastSeenState.lastSeenTimestamp, 0)).
					Msg("device health changed")
				p.cache[id] = &DeviceState{
					lastSeenTimestamp: time.Now().Unix(),
					Device:            device,
				}
				changeSet = append(changeSet, device)
			}
		}
	}

	// remove devices that are no longer present from cache
	for id := range p.cache {
		if _, exists := fullSet[id]; !exists {
			log.Info().Msg("removed device from cache")
			delete(p.cache, id)
		}
	}

	return fullSet, changeSet, nil
}
