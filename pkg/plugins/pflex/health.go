package pflex

type DeviceHealth string

func (h DeviceHealth) String() string {
	return string(h)
}

const (
	Healthy   DeviceHealth = "healthy"
	Unhealthy DeviceHealth = "unhealthy"
)
