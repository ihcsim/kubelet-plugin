package pflex

import (
	"context"
	"errors"
	"net"
	"time"
)

var grpcReadyTimeoutDuration = 600 * time.Second

func (p *DevicePlugin) grpcServe(ctx context.Context, errCh chan<- error) error {
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		p.gserver.GracefulStop()
	}()

	go func() {
		defer l.Close()
		if err := p.gserver.Serve(l); err != nil {
			errCh <- err
			return
		}
	}()

	return nil
}

func (p *DevicePlugin) grpcReady(ctx context.Context) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			if !errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
		default:
			if p.gserver.GetServiceInfo() != nil {
				break LOOP
			}
			continue
		}
	}

	return nil
}
