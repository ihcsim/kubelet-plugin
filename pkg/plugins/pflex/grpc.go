package pflex

import (
	"context"
	"fmt"
	"net"
)

var serverNotReadyWithinTimeout = fmt.Errorf("server not ready within timeout")

func (p *DevicePlugin) grpcStartServe(ctx context.Context, errCh chan<- error) {
	go func() {
		l, err := net.Listen("unix", socket)
		if err != nil {
			errCh <- err
			return
		}

		p.grpcRegisterStop(ctx)
		errCh <- p.gserver.Serve(l)
	}()
}

func (p *DevicePlugin) grpcRegisterStop(ctx context.Context) {
	go func() {
		<-ctx.Done()
		p.gserver.GracefulStop()
	}()
}

func (p *DevicePlugin) grpcReady(ctx context.Context) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			return serverNotReadyWithinTimeout
		default:
			if p.gserver.GetServiceInfo() == nil {
				continue
			}
			break LOOP
		}
	}

	return nil
}
