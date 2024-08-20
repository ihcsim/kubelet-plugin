package generic

import (
	"context"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func (p *DevicePlugin) ListAndWatch(empty *v1beta1.Empty, server v1beta1.DevicePlugin_ListAndWatchServer) error {
	return nil
}

func (p *DevicePlugin) Allocate(ctx context.Context, r *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse, error) {
	return nil, nil
}

func (p *DevicePlugin) GetPreferredAllocation(ctx context.Context, r *v1beta1.PreferredAllocationRequest) (*v1beta1.PreferredAllocationResponse, error) {
	return nil, nil
}

func (p *DevicePlugin) PreStartContainer(ctx context.Context, r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	return nil, nil
}

func (p *DevicePlugin) GetDevicePluginOptions(context.Context, *v1beta1.Empty) (*v1beta1.DevicePluginOptions, error) {
	return nil, nil
}
