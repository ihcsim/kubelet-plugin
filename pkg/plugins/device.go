package plugins

import "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

const (
	Healthy   DeviceHealth = v1beta1.Healthy
	Unhealthy DeviceHealth = v1beta1.Unhealthy
)

// Device represents a device managed by this plugin.
type Device struct {
	ID     string
	Health DeviceHealth
}

// DeviceState maintains the last seen of a device at a given timestamp.
type DeviceState struct {
	LastSeenTimestamp int64
	*Device
}

type DeviceHealth string

func (h DeviceHealth) String() string {
	return string(h)
}
