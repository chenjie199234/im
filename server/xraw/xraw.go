package xraw

import (
	"crypto/tls"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/util"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/stream"
)

var s *stream.Instance

func StartRawServer() {
	c := config.GetRawServerConfig()
	var tlsc *tls.Config
	if len(c.Certs) > 0 {
		certificates := make([]tls.Certificate, 0, len(c.Certs))
		for cert, key := range c.Certs {
			temp, e := tls.LoadX509KeyPair(cert, key)
			if e != nil {
				log.Error(nil, "[xraw] load cert failed:", log.String("cert", cert), log.String("key", key), log.CError(e))
				return
			}
			certificates = append(certificates, temp)
		}
		tlsc = &tls.Config{Certificates: certificates}
	}
	s, _ = stream.NewInstance(&stream.InstanceConfig{
		TcpC:               &stream.TcpConfig{ConnectTimeout: c.ConnectTimeout.StdDuration()},
		HeartprobeInterval: c.HeartProbe.StdDuration(),
		GroupNum:           c.GroupNum,
		VerifyFunc:         util.RawVerify,
		OnlineFunc:         util.RawOnline,
		PingPongFunc:       util.RawPingPong,
		UserdataFunc:       util.RawUser,
		OfflineFunc:        util.RawOffline,
	})

	util.SetRawInstance(s)

	if e := s.StartServer(":8080", tlsc); e != nil && e != stream.ErrServerClosed {
		log.Error(nil, "[xraw] start server failed", log.CError(e))
		return
	}
	log.Info(nil, "[xraw] server closed")
}

func StopRawServer() {
	if s != nil {
		s.Stop()
	}
}
