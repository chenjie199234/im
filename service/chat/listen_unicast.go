package chat

import (
	"bytes"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/util/common"
)

type PushMsg struct {
	Type   string //send or recall
	Send   *Send
	Recall *Recall
}
type Send struct {
	Sender    string
	Msg       string
	Extra     string
	MsgIndex  uint64
	Timestamp uint64
}
type Recall struct {
	MsgIndex    uint64
	RecallIndex uint64
}

func (s *Service) ListenUnicast() (stop func()) {
	rdb := config.GetRedis("mq_redis")
	stop, _ = rdb.SubUnicast(s.rawname, 64, func(data []byte) {
		if index := bytes.Index(data, []byte{'_'}); index != 24 {
			log.Error(nil, "[ListenUnicast] failed", log.String("data", common.BTS(data)), log.CError(ecode.ErrCacheDataBroken))
			return
		}
		receiver := common.BTS(data[:24])
		if p := s.rawInstance.GetPeer(receiver); p != nil {
			w := (*writer)(p.GetData())
			w.add(data[25:])
		}
	})
	return stop
}
