package crand

import (
	"context"
	"time"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const hostDevicePath = "/dev"

var watchIntervalDuration = 10 * time.Second

func (p *DevicePlugin) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	p.log.Debug().Msg("calling DevicePlugin.ListAndWatch()")
	tick := time.NewTicker(watchIntervalDuration)
	defer tick.Stop()

	for range tick.C {
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

		p.log.Debug().Any("changeSet", changeSet).Msg("sending ListAndWatch response")
		if err := stream.Send(resp); err != nil {
			p.log.Err(err).Msg("failed to send ListAndWatch response")
			return err
		}
	}
	return nil
}

func (p *DevicePlugin) Allocate(ctx context.Context, r *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse, error) {
	p.log.Debug().Msg("calling DevicePlugin.Allocate()")
	resp := &v1beta1.AllocateResponse{}
	for _, allocateRequest := range r.ContainerRequests {
		// see cdi/crand.yaml for crand CDI configuration
		car := &v1beta1.ContainerAllocateResponse{}
		for _, id := range allocateRequest.DevicesIDs {
			p.log.Info().Str("name", id).Msg("allocating CDI device")
			car.CDIDevices = append(car.CDIDevices, &v1beta1.CDIDevice{
				Name: id,
			})
		}
		resp.ContainerResponses = append(resp.ContainerResponses, car)
	}
	return resp, nil
}

func (p *DevicePlugin) GetPreferredAllocation(ctx context.Context, r *v1beta1.PreferredAllocationRequest) (*v1beta1.PreferredAllocationResponse, error) {
	p.log.Debug().Msg("calling DevicePlugin.GetPreferredAllocation()")
	return &v1beta1.PreferredAllocationResponse{}, nil
}

func (p *DevicePlugin) PreStartContainer(ctx context.Context, r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	p.log.Debug().Msg("calling DevicePlugin.PrestartContainer()")
	return &v1beta1.PreStartContainerResponse{}, nil
}

func (p *DevicePlugin) GetDevicePluginOptions(context.Context, *v1beta1.Empty) (*v1beta1.DevicePluginOptions, error) {
	p.log.Debug().Msg("calling DevicePlugin.GetDevicePluginOptions()")
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
