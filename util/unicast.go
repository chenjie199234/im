package util

import (
	"context"
	"strconv"

	"github.com/chenjie199234/im/config"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/util/common"
	gredis "github.com/redis/go-redis/v9"
)

var gateredis = config.GetRedis("gate_redis")

func GetSession(ctx context.Context, userid string) (*model.IMSession, error) {
	key := "raw_user_{" + userid + "}"
	rs, e := gateredis.HGetAll(ctx, key).Result()
	if e == gredis.Nil {
		e = ecode.ErrSession
	}
	if e != nil {
		return nil, e
	}
	remoteaddr, ok := rs["remote_addr"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	realip, ok := rs["real_ip"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	gate, ok := rs["gate"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	str, ok := rs["netlag"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	netlag, e := strconv.ParseUint(str, 10, 64)
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	return &model.IMSession{RemoteAddr: remoteaddr, RealIP: realip, Gate: gate, Netlag: netlag}, nil
}

// data can't be nil or empty
func Unicast(ctx context.Context, userid string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	session, e := GetSession(ctx, userid)
	if e != nil {
		return e
	}
	return gateredis.PubUnicast(ctx, session.Gate, 32, userid, userid+"_"+common.BTS(data))
}
