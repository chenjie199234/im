package xweb

import (
	"crypto/tls"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/model"
	"github.com/chenjie199234/im/service"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/util/ctime"
	"github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/web/mids"
)

var s *web.WebServer

// StartWebServer -
func StartWebServer() {
	c := config.GetWebServerConfig()
	var tlsc *tls.Config
	if len(c.Certs) > 0 {
		certificates := make([]tls.Certificate, 0, len(c.Certs))
		for cert, key := range c.Certs {
			temp, e := tls.LoadX509KeyPair(cert, key)
			if e != nil {
				log.Error(nil, "[xweb] load cert failed:", log.String("cert", cert), log.String("key", key), log.CError(e))
				return
			}
			certificates = append(certificates, temp)
		}
		tlsc = &tls.Config{Certificates: certificates}
	}
	var e error
	if s, e = web.NewWebServer(c.ServerConfig, model.Project, model.Group, model.Name, tlsc); e != nil {
		log.Error(nil, "[xweb] new server failed", log.CError(e))
		return
	}
	UpdateHandlerTimeout(config.AC.HandlerTimeout)
	UpdateWebPathRewrite(config.AC.WebPathRewrite)

	r := s.NewRouter()

	//this place can register global midwares
	//r.Use(globalmidwares)

	//you just need to register your service here
	api.RegisterStatusWebServer(r, service.SvcStatus, mids.AllMids())
	api.RegisterRelationWebServer(r, service.SvcRelation, mids.AllMids())
	api.RegisterChatWebServer(r, service.SvcChat, mids.AllMids())
	//example
	//api.RegisterExampleWebServer(r, service.SvcExample, mids.AllMids())

	s.SetRouter(r)
	if e = s.StartWebServer(":8000"); e != nil && e != web.ErrServerClosed {
		log.Error(nil, "[xweb] start server failed", log.CError(e))
		return
	}
	log.Info(nil, "[xweb] server closed")
}

// UpdateHandlerTimeout -
// first key path,second key method,value timeout duration
func UpdateHandlerTimeout(timeout map[string]map[string]ctime.Duration) {
	if s != nil {
		s.UpdateHandlerTimeout(timeout)
	}
}

// UpdateWebPathRewrite -
// first key method,second key origin url,value rewrite url
func UpdateWebPathRewrite(rewrite map[string]map[string]string) {
	if s != nil {
		s.UpdateHandlerRewrite(rewrite)
	}
}

// StopWebServer force - false(graceful),true(not graceful)
func StopWebServer(force bool) {
	if s != nil {
		s.StopWebServer(force)
	}
}

