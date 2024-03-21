package chat

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

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

// if MsgIndex or RecallIndex or AckIndex is math.MaxUint32 means this kind of index will not be setted
func (d *Dao) RedisSetIndex(ctx context.Context, userid, chatkey string, MsgIndex, RecallIndex, AckIndex uint32) error {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_msg"
	field2 := chatkey + "_recall"
	field3 := chatkey + "_ack"
	args := make([]interface{}, 0, 7)
	args = append(args, defaultExpire)
	if MsgIndex != math.MaxUint32 {
		args = append(args, field1, MsgIndex)
	}
	if RecallIndex != math.MaxUint32 {
		args = append(args, field2, RecallIndex)
	}
	if AckIndex != math.MaxUint32 {
		args = append(args, field3, AckIndex)
	}
	return setindex.Run(ctx, d.redis, []string{key}, args...).Err()
}

// return MsgIndex,RecallIndex,AckIndex
func (d *Dao) RedisGetIndex(ctx context.Context, userid, chatkey string) (*model.IMIndex, error) {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_msg"
	field2 := chatkey + "_recall"
	field3 := chatkey + "_ack"
	rs, e := d.redis.HMGet(ctx, key, field1, field2, field3).Result()
	if e != nil {
		return nil, e
	}
	if rs[0] == nil || rs[1] == nil || rs[2] == nil {
		return nil, gredis.Nil
	}
	MsgIndex, e := strconv.ParseUint(rs[0].(string), 10, 32)
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	RecallIndex, e := strconv.ParseUint(rs[1].(string), 10, 32)
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	AckIndex, e := strconv.ParseUint(rs[2].(string), 10, 32)
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	return &model.IMIndex{MsgIndex: uint32(MsgIndex), RecallIndex: uint32(RecallIndex), AckIndex: uint32(AckIndex)}, nil
}

// return: key:chatkey,value:index
func (d *Dao) RedisGetIndexAll(ctx context.Context, userid string) (map[string]*model.IMIndex, error) {
	key := "im_user_{" + userid + "}"
	tmp, e := d.redis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	r := make(map[string]*model.IMIndex, len(tmp)/3)
	for k := range tmp {
		var chatkey string
		if strings.HasSuffix(k, "_msg") {
			chatkey = k[:len(k)-4]
		} else if strings.HasSuffix(k, "_recall") {
			chatkey = k[:len(k)-7]
		} else if strings.HasSuffix(k, "_ack") {
			chatkey = k[:len(k)-4]
		} else {
			return nil, ecode.ErrCacheDataBroken
		}
		field1 := chatkey + "_msg"
		field2 := chatkey + "_recall"
		field3 := chatkey + "_ack"
		var value1, value2, value3 uint32
		if str, ok := tmp[field1]; ok {
			delete(tmp, field1)
			index, e := strconv.ParseUint(str, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			value1 = uint32(index)
		} else {
			delete(tmp, field2)
			delete(tmp, field3)
			continue
		}
		if str, ok := tmp[field2]; ok {
			delete(tmp, field2)
			index, e := strconv.ParseUint(str, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			value2 = uint32(index)
		} else {
			delete(tmp, field3)
			continue
		}
		if str, ok := tmp[field3]; ok {
			delete(tmp, field3)
			index, e := strconv.ParseUint(str, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			value3 = uint32(index)
		} else {
			continue
		}
		r[chatkey] = &model.IMIndex{
			MsgIndex:    value1,
			RecallIndex: value2,
			AckIndex:    value3,
		}
	}
	return r, nil
}
func (d *Dao) RedisDelIndex(ctx context.Context, userid, chatkey string) error {
	key := "im_user_{" + userid + "}"
	field1 := chatkey + "_msg"
	field2 := chatkey + "_recall"
	field3 := chatkey + "_ack"
	return d.redis.HDel(ctx, key, field1, field2, field3).Err()
}
