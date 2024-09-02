package plugins

import (
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"
)

var grpcReadyTimeoutDuration = 600 * time.Second

type Server struct {
	GRPC       *grpc.Server
	socketPath string
}

func NewServer(socketPath string) *Server {
	return &Server{
		GRPC:       grpc.NewServer(),
		socketPath: socketPath,
	}
}

func (s *Server) GRPCServe(ctx context.Context, errCh chan<- error) error {
	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		s.GRPC.GracefulStop()
	}()

	go func() {
		defer l.Close()
		if err := s.GRPC.Serve(l); err != nil {
			errCh <- err
			return
		}
	}()

	return nil
}

func (s *Server) GRPCReady(ctx context.Context) error {
	ready, cancel := context.WithTimeout(ctx, grpcReadyTimeoutDuration)
	defer cancel()

LOOP:
	for {
		select {
		case <-ready.Done():
			if !errors.Is(ready.Err(), context.Canceled) {
				return ctx.Err()
			}
		default:
			if s.GRPC.GetServiceInfo() != nil {
				break LOOP
			}
			continue
		}
	}

	return nil
}
