package kvm

import (
	"context"
	"time"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	containerDevicePath = "/dev/kvm"
	hostDevicePath      = "/dev/kvm"
	devicePermissions   = "rw"
)

var watchIntervalDuration = 10 * time.Second

func (p *DevicePlugin) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	p.log.Debug().Msg("calling DevicePlugin.ListAndWatch()")
	tick := time.NewTicker(watchIntervalDuration)
	defer tick.Stop()

	for range tick.C {
		hasChanged, err := p.discoverDevices()
		if err != nil {
			return err
		}

		if !hasChanged {
			continue
		}

		changed := p.cache
		resp := &v1beta1.ListAndWatchResponse{
			Devices: []*v1beta1.Device{
				&v1beta1.Device{
					ID:     changed.ID,
					Health: changed.Health.String(),
				},
			},
		}

		p.log.Debug().Any("changeSet", changed).Msg("sending ListAndWatch response")
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
		car := &v1beta1.ContainerAllocateResponse{}
		for _, id := range allocateRequest.DevicesIDs {
			p.log.Info().Str("name", id).Msg("allocating CDI device")
			car.CDIDevices = []*v1beta1.CDIDevice{
				{
					Name: id,
				},
			}
			car.Devices = []*v1beta1.DeviceSpec{
				{
					ContainerPath: containerDevicePath,
					HostPath:      hostDevicePath,
					Permissions:   devicePermissions,
				},
			}
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
