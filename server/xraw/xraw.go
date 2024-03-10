package xraw

import (
	"crypto/tls"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/service"

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
				log.Error(nil, "[xrawchat] load cert failed:", log.String("cert", cert), log.String("key", key), log.CError(e))
				return
			}
			certificates = append(certificates, temp)
		}
		tlsc = &tls.Config{Certificates: certificates}
	}
	s, _ = stream.NewInstance(&stream.InstanceConfig{
		TcpC:               &stream.TcpConfig{ConnectTimeout: c.ConnectTimeout.StdDuration()},
		HeartprobeInterval: c.HeartProbe.StdDuration(),
		GroupNum:           uint32(c.GroupNum),
		VerifyFunc:         service.SvcChat.RawVerify,
		OnlineFunc:         service.SvcChat.RawOnline,
		PingPongFunc:       service.SvcChat.RawPingPong,
		UserdataFunc:       service.SvcChat.RawUser,
		OfflineFunc:        service.SvcChat.RawOffline,
	})
	if e := s.StartServer(":8080", tlsc); e != nil && e != stream.ErrServerClosed {
		log.Error(nil, "[xrawchat] start server failed", log.CError(e))
		return
	}
	log.Info(nil, "[xrawchat] server closed")
}

func StopRawServer() {
	if s != nil {
		s.Stop()
	}
}
