package crand

import "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

type DeviceHealth string

func (h DeviceHealth) String() string {
	return string(h)
}

const (
	Healthy   DeviceHealth = v1beta1.Healthy
	Unhealthy DeviceHealth = v1beta1.Unhealthy
)
