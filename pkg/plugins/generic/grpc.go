package generic

import "net"

func (p *DevicePlugin) serve() error {
	l, err := net.Listen("unix", p.socket)
	if err != nil {
		return err
	}

	p.log.Info().Str("addr", l.Addr().String()).Str("protocol", l.Addr().Network()).Msg("grpc server started")
	return p.gserver.Serve(l)
}
