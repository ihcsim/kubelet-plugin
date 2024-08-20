package generic

import (
	"context"
	"fmt"
	"net"
)

var serverNotReadyWithinTimeout = fmt.Errorf("server not ready within timeout")

func (p *DevicePlugin) serve() error {
	l, err := net.Listen("unix", p.socket)
	if err != nil {
		return err
	}

	p.log.Info().Str("addr", fmt.Sprintf("%s://%s", l.Addr().Network(), l.Addr().String())).
		Msg("starting grpc server")
	return p.gserver.Serve(l)
}

func (p *DevicePlugin) gracefulStop() {
	p.gserver.GracefulStop()
}

func (p *DevicePlugin) grpcReady(ctx context.Context) error {
	var ready bool
	for {
		select {
		case <-ctx.Done():
			if !ready {
				return serverNotReadyWithinTimeout
			}
		default:
			if p.gserver.GetServiceInfo() == nil {
				continue
			}
			ready = true
			return nil
		}
	}

	return nil
}
