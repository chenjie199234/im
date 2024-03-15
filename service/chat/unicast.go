package chat

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/util/common"
)

type PushMsg struct {
	Type    string      //msg or recall
	Content interface{} // *Msg or *Recall
	// Msg     *Msg
	// Recall  *Recall
}
type Msg struct {
	Sender    string
	Msg       string
	Extra     string
	MsgIndex  uint32
	Timestamp uint32 //unit seconds
}
type Recall struct {
	MsgIndex    uint32
	RecallIndex uint32
}

var rdb = config.GetRedis("mq_redis")

func (s *Service) listenUnicast() (stop func()) {
	stop, _ = rdb.SubUnicast(s.rawname, 32, func(data []byte) {
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
func (s *Service) unicast(ctx context.Context, userid string, msg *PushMsg) error {
	data, _ := json.Marshal(msg)
	if e := rdb.PubUnicast(ctx, s.rawname, 32, userid, userid+"_"+common.BTS(data)); e != nil {
		log.Error(ctx, "[Unicast] failed", log.String("user_id", userid), log.Any("msg", msg), log.CError(e))
		return e
	}
	return nil
}
