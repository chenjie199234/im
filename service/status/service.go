package status

import (
	"context"
	"time"

	// "github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/api"
	statusdao "github.com/chenjie199234/im/dao/status"
	// "github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/monitor"
	"github.com/chenjie199234/Corelib/util/graceful"
	"github.com/chenjie199234/Corelib/util/host"
	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/log"
	// "github.com/chenjie199234/Corelib/web"
)

// Service subservice for status business
type Service struct {
	stop *graceful.Graceful

	statusDao *statusdao.Dao
}

// Start -
func Start() *Service {
	return &Service{
		stop: graceful.New(),

		//statusDao: statusdao.NewDao(config.GetMysql("status_mysql"), config.GetRedis("status_redis"), config.GetMongo("status_mongo")),
		statusDao: statusdao.NewDao(nil, nil, nil),
	}
}

// Ping -
func (s *Service) Ping(ctx context.Context, in *api.Pingreq) (*api.Pingresp, error) {
	//if _, ok := ctx.(*crpc.Context); ok {
	//        log.Info("this is a crpc call")
	//}
	//if _, ok := ctx.(*cgrpc.Context); ok {
	//        log.Info("this is a cgrpc call")
	//}
	//if _, ok := ctx.(*web.Context); ok {
	//        log.Info("this is a web call")
	//}
	totalmem, lastmem, maxmem := monitor.GetMEM()
	lastcpu, maxcpu, avgcpu := monitor.GetCPU()
	return &api.Pingresp{
		ClientTimestamp: in.Timestamp,
		ServerTimestamp: time.Now().UnixNano(),
		TotalMem:        totalmem,
		CurMemUsage:     lastmem,
		MaxMemUsage:     maxmem,
		CpuNum:          monitor.CPUNum,
		CurCpuUsage:     lastcpu,
		AvgCpuUsage:     avgcpu,
		MaxCpuUsage:     maxcpu,
		Host:            host.Hostname,
		Ip:              host.Hostip,
	}, nil
}

// Stop -
func (s *Service) Stop() {
	s.stop.Close(nil, nil)
}