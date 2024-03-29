package util

import (
	"bytes"
	"context"
	"strconv"
	"time"
	"unsafe"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/mids"
	"github.com/chenjie199234/Corelib/stream"
	"github.com/chenjie199234/Corelib/util/common"
	gredis "github.com/redis/go-redis/v9"
)

type MQ struct {
	Type    string      //msg or recall or userRelationAdd or userRelationDel or groupJoin or groupLeave or groupKick
	Content interface{} // *Msg or *Recall or *UserAction or *GroupAction
}
type Msg struct {
	Sender    string `json:"sender"`
	Msg       string `json:"msg"`
	Extra     string `json:"extra"`
	MsgIndex  uint32 `json:"msg_index"`
	Timestamp uint32 `json:"timestamp"` //unit seconds
}
type Recall struct {
	MsgIndex    uint32 `json:"msg_index"`
	RecallIndex uint32 `json:"recall_index"`
}
type UserAction struct {
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	Name        string `json:"name,omitempty"`
	MsgIndex    uint32 `json:"msg_index,omitempty"`
	RecallIndex uint32 `json:"recall_index,omitempty"`
	AckIndex    uint32 `json:"ack_index,omitempty"`
}
type GroupAction struct {
	Member string `json:"member"`
	Name   string `json:"name,omitempty"`
	Duty   uint8  `json:"duty,omitempty"`
}

var setsession *gredis.Script
var updatesession *gredis.Script
var delsession *gredis.Script

func init() {
	setsession = gredis.NewScript(`redis.call("DEL",KEYS[1])
local num=redis.call("HSET",KEYS[1],unpack(ARGV,2))
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
	updatesession = gredis.NewScript(`local tmp=redis.call("HGET",KEYS[1],"remote_addr")
if(not tmp or tmp~=ARGV[2]) then
	return nil
end
redis.call("HSET",KEYS[1],unpack(ARGV,3))
redis.call("EXPIRE",KEYS[1],ARGV[1])
return 0`)
	delsession = gredis.NewScript(`local tmp=redis.call("HGET",KEYS[1],"remote_addr")
if(tmp and tmp==ARGV[1]) then
	return redis.call("DEL",KEYS[1])
end
return 0`)
}

var RawName = strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + common.MakeRandCode(10)
var rawInstance *stream.Instance
var rdb = config.GetRedis("gate_redis")
var online func(context.Context) error

func SetOnlineFunc(f func(ctx context.Context) error) {
	online = f
}
func SetRawInstance(rawinstance *stream.Instance) {
	rawInstance = rawinstance
	rdb.SubUnicast(RawName, 32, func(data []byte, last bool) {
		if index := bytes.Index(data, []byte{'_'}); index != 24 {
			log.Error(nil, "[ListenUnicast] failed", log.String("data", common.BTS(data)), log.CError(ecode.ErrCacheDataBroken))
			return
		}
		receiver := common.BTS(data[:24])
		if p := rawInstance.GetPeer(receiver); p != nil {
			writer := *(*chan []byte)(p.GetData())
			select {
			case writer <- data[25:]:
				if len(writer) > 128 {
					log.Warn(nil, "[ListenUnicast] too many stacking msgs need to send",
						log.String("user_id", p.GetUniqueID()),
						log.String("remote_addr", p.GetRemoteAddr()),
						log.String("real_ip", p.GetRealPeerIP()))
				}
			default:
				log.Error(nil, "[ListenUnicast] too many stacking msgs need to send",
					log.String("user_id", p.GetUniqueID()),
					log.String("remote_addr", p.GetRemoteAddr()),
					log.String("real_ip", p.GetRealPeerIP()))
				p.Close(false)
			}
		}
	})
}
func RawVerify(ctx context.Context, peerVerifyData []byte) (response []byte, uniqueid string, success bool) {
	t := mids.VerifyToken(ctx, common.BTS(peerVerifyData))
	if t == nil {
		log.Error(ctx, "[RawVerify] token verify failed", log.String("token", string(peerVerifyData)))
		return nil, "", false
	}
	//check rate 1 minute do 10 times
	status, e := rdb.RateLimit(ctx, map[string][2]uint64{"user_login_{" + t.UserID + "}": {10, 59}})
	if e != nil {
		log.Error(ctx, "[RawVerify] rate check failed", log.String("user_id", t.UserID), log.CError(e))
		return nil, "", false
	}
	if !status {
		log.Error(ctx, "[RawVerify] login too frequently", log.String("user_id", t.UserID))
		return nil, "", false
	}
	if p := rawInstance.GetPeer(t.UserID); p != nil {
		//this user already connected
		//this new connection will kick the old connection
		p.Close(true)
	}
	return nil, t.UserID, true
}
func RawOnline(ctx context.Context, p *stream.Peer) (success bool) {
	//send first ping to get the netlag from pong as soon as possible
	if e := p.SendPing(); e != nil {
		log.Error(p, "[RawOnline] send first ping failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		return false
	}
	expire := uint64(config.GetRawServerConfig().HeartProbe.StdDuration().Seconds()) * 3
	key := "raw_user_{" + p.GetUniqueID() + "}"
	if e := setsession.Run(p, rdb, []string{key}, expire, "remote_addr", p.GetRemoteAddr(), "real_ip", p.GetRealPeerIP(), "gate", RawName, "netlag", 0).Err(); e != nil {
		log.Error(p, "[RawOnline] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		return false
	}
	w := make(chan []byte, 512)
	p.SetData(unsafe.Pointer(&w))
	//TODO call relation service's Online function
	online(metadata.AddMetadata(ctx, "Token-User", p.GetUniqueID()))
	go func() {
		//writer goroutine
		for {
			data := <-w
			if len(data) == 0 {
				//offline
				return
			}
			if e := p.SendMessage(nil, data, nil, nil); e != nil {
				if e == stream.ErrConnClosed {
					//offline
					return
				}
				log.Error(nil, "[RawWriter] failed",
					log.String("user_id", p.GetUniqueID()),
					log.String("remote_addr", p.GetRemoteAddr()),
					log.String("real_ip", p.GetRealPeerIP()),
					log.String("msg", common.BTS(data)),
					log.CError(e))
			}
		}
	}()
	return true
}
func RawPingPong(p *stream.Peer) {
	expire := uint64(config.GetRawServerConfig().HeartProbe.StdDuration().Seconds()) * 3
	key := "raw_user_{" + p.GetUniqueID() + "}"
	if e := updatesession.Run(p, rdb, []string{key}, expire, p.GetRemoteAddr(), "netlag", p.GetNetlag()).Err(); e != nil {
		log.Error(p, "[RawPingPong] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		if e == gredis.Nil {
			//be kicked or session expired
			p.Close(false)
		}
	}
}
func RawUser(p *stream.Peer, userdata []byte) {
	//client will not send message to server
	p.Close(false)
}
func RawOffline(p *stream.Peer) {
	w := *(*chan []byte)(p.GetData())
	//wake up the writer goroutine
	select {
	case w <- nil:
	default:
	}
	key := "raw_user_{" + p.GetUniqueID() + "}"
	if e := delsession.Run(p, rdb, []string{key}, p.GetRemoteAddr()).Err(); e != nil {
		log.Error(p, "[RawOffline] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
	}
}
