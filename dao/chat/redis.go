package chat

import (
	"context"
	"strconv"
	"strings"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/util/common"
	gredis "github.com/redis/go-redis/v9"
)

const defaultExpire = 604800

var setindex *gredis.Script

func init() {
	setindex = gredis.NewScript(`local num=0
for i=2,#ARGV,2 do
	local tmp=redis.call("HGET",KEYS[1],ARGV[i])
	local args={}
	if(not tmp or tonumber(tmp)<tonumber(ARGV[i+1])) then
		table.insert(args,ARGV[i])
		table.insert(args,ARGV[i+1])
		num=num+1
	end
	if(#args>0) then
		redis.call("HSET",KEYS[1],unpack(args))
	end
end
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
}

//------------------------------------index-----------------------------------

// if MsgIndex or RecallIndex or AckIndex is 0 means this kind of index will not be setted
func (d *Dao) RedisSetIndex(ctx context.Context, userid, chatkey string, MsgIndex, RecallIndex, AckIndex uint32) error {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_exist"
	field2 := chatkey + "_msg"
	field3 := chatkey + "_recall"
	field4 := chatkey + "_ack"
	args := make([]interface{}, 0, 8)
	args = append(args, defaultExpire)
	if MsgIndex == 0 && RecallIndex == 0 && AckIndex == 0 {
		args = append(args, field1, 0)
	} else {
		args = append(args, field1, 1)
	}
	if MsgIndex != 0 {
		args = append(args, field2, MsgIndex)
	}
	if RecallIndex != 0 {
		args = append(args, field3, RecallIndex)
	}
	if AckIndex != 0 {
		args = append(args, field4, AckIndex)
	}
	return setindex.Run(ctx, d.imredis, []string{key}, args...).Err()
}

// return MsgIndex,RecallIndex,AckIndex
func (d *Dao) RedisGetIndex(ctx context.Context, userid, chatkey string) (*model.IMIndex, error) {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_exist"
	field2 := chatkey + "_msg"
	field3 := chatkey + "_recall"
	field4 := chatkey + "_ack"
	rs, e := d.imredis.HMGet(ctx, key, field1, field2, field3, field4).Result()
	if e != nil {
		return nil, e
	}
	if rs[0] == nil {
		return nil, gredis.Nil
	}
	var MsgIndex, RecallIndex, AckIndex uint64
	if rs[1] != nil {
		if MsgIndex, e = strconv.ParseUint(rs[0].(string), 10, 32); e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
	}
	if rs[2] != nil {
		if RecallIndex, e = strconv.ParseUint(rs[1].(string), 10, 32); e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
	}
	if rs[3] != nil {
		if AckIndex, e = strconv.ParseUint(rs[2].(string), 10, 32); e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
	}
	return &model.IMIndex{MsgIndex: uint32(MsgIndex), RecallIndex: uint32(RecallIndex), AckIndex: uint32(AckIndex)}, nil
}

// return: key:chatkey,value:index
func (d *Dao) RedisGetIndexAll(ctx context.Context, userid string) (map[string]*model.IMIndex, error) {
	key := "im_user_{" + userid + "}"
	tmp, e := d.imredis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	r := make(map[string]*model.IMIndex, len(tmp)/4)
	for k, v := range tmp {
		var chatkey string
		if strings.HasSuffix(k, "_exist") {
			chatkey = k[:len(k)-6]
			if _, ok := r[chatkey]; !ok {
				r[chatkey] = &model.IMIndex{}
			}
		} else if strings.HasSuffix(k, "_msg") {
			index, e := strconv.ParseUint(v, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			chatkey = k[:len(k)-4]
			if v, ok := r[chatkey]; !ok {
				r[chatkey] = &model.IMIndex{MsgIndex: uint32(index)}
			} else {
				v.MsgIndex = uint32(index)
			}
		} else if strings.HasSuffix(k, "_recall") {
			index, e := strconv.ParseUint(v, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			chatkey = k[:len(k)-7]
			if v, ok := r[chatkey]; !ok {
				r[chatkey] = &model.IMIndex{RecallIndex: uint32(index)}
			} else {
				v.RecallIndex = uint32(index)
			}
		} else if strings.HasSuffix(k, "_ack") {
			index, e := strconv.ParseUint(v, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			chatkey = k[:len(k)-4]
			if v, ok := r[chatkey]; !ok {
				r[chatkey] = &model.IMIndex{AckIndex: uint32(index)}
			} else {
				v.AckIndex = uint32(index)
			}
		}
	}
	return r, nil
}
func (d *Dao) RedisDelIndex(ctx context.Context, userid, chatkey string) error {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_exist"
	field2 := chatkey + "_msg"
	field3 := chatkey + "_recall"
	field4 := chatkey + "_ack"
	return d.imredis.HDel(ctx, key, field1, field2, field3, field4).Err()
}

// -------------------------------------------gate redis-------------------------------------
func (d *Dao) GetSession(ctx context.Context, userid string) (*model.IMSession, error) {
	key := "raw_user_{" + userid + "}"
	rs, e := d.gateredis.HGetAll(ctx, key).Result()
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
func (d *Dao) Unicast(ctx context.Context, rawname, userid string, data []byte) error {
	return d.gateredis.PubUnicast(ctx, rawname, 32, userid, userid+"_"+common.BTS(data))
}
