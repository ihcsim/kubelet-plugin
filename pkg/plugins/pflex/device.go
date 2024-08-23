package pflex

import (
	"context"
	"time"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const hostDevicePath = "/dev"

var watchIntervalDuration = 5 * time.Second

func (p *DevicePlugin) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	p.log.Debug().Msg("Calling DevicePlugin.ListAndWatch()")
	tick := time.Tick(watchIntervalDuration)
	for range tick {
		_, changeSet, err := p.discoverDevices()
		if err != nil {
			return err
		}

		if len(changeSet) == 0 {
			continue
		}

		resp := &v1beta1.ListAndWatchResponse{}
		for _, change := range changeSet {
			device := &v1beta1.Device{
				ID:     change.ID,
				Health: change.Health.String(),
			}
			resp.Devices = append(resp.Devices, device)
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (p *DevicePlugin) Allocate(ctx context.Context, r *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse, error) {
	p.log.Debug().Msg("Calling DevicePlugin.Allocate()")
	return nil, nil
}

func (p *DevicePlugin) GetPreferredAllocation(ctx context.Context, r *v1beta1.PreferredAllocationRequest) (*v1beta1.PreferredAllocationResponse, error) {
	p.log.Debug().Msg("Calling DevicePlugin.GetPreferredAllocation()")
	return nil, nil
}

func (p *DevicePlugin) PreStartContainer(ctx context.Context, r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	p.log.Debug().Msg("Calling DevicePlugin.PrestartContainer()")
	return nil, nil
}

func (p *DevicePlugin) GetDevicePluginOptions(context.Context, *v1beta1.Empty) (*v1beta1.DevicePluginOptions, error) {
	p.log.Debug().Msg("Calling DevicePlugin.GetDevicePluginOptions()")
	return &v1beta1.DevicePluginOptions{}, nil
}

// Device represents a device managed by this plugin.
type Device struct {
	ID     string
	Health DeviceHealth
}

// DeviceState maintains the last seen of a device at a given timestamp.
type DeviceState struct {
	lastSeenTimestamp int64
	*Device
}
