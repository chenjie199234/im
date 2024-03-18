package chat

import (
	"context"
	"unsafe"

	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/log"
	cmongo "github.com/chenjie199234/Corelib/mongo"
	cmysql "github.com/chenjie199234/Corelib/mysql"
	credis "github.com/chenjie199234/Corelib/redis"
	"github.com/chenjie199234/Corelib/util/oneshot"
	gredis "github.com/redis/go-redis/v9"
)

// Dao this is a data operation layer to operate chat service's data
type Dao struct {
	mysql              *cmysql.Client
	imredis, gateredis *credis.Client
	mongo              *cmongo.Client
}

// NewDao Dao is only a data operation layer
// don't write business logic in this package
// business logic should be written in service package
func NewDao(mysql *cmysql.Client, imredis, gateredis *credis.Client, mongo *cmongo.Client) *Dao {
	return &Dao{
		mysql:     mysql,
		imredis:   imredis,
		gateredis: gateredis,
		mongo:     mongo,
	}
}
func (d *Dao) GetIndex(ctx context.Context, userid, chatkey string) (*model.IMIndex, error) {
	if index, e := d.RedisGetIndex(ctx, userid, chatkey); e == nil {
		return index, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetIndex] redis op failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeIndex, e := oneshot.Do("GetIndex_msg_recall_"+chatkey, func() (unsafe.Pointer, error) {
		msg, e := d.MongoGetMaxMsgIndex(ctx, chatkey)
		if e != nil {
			log.Error(ctx, "[dao.GetIndex] db op failed", log.String("chat_key", chatkey), log.CError(e))
			return nil, e
		}
		recall, e := d.MongoGetMaxRecallIndex(ctx, chatkey)
		if e != nil {
			log.Error(ctx, "[dao.GetIndex] db op failed", log.String("chat_key", chatkey), log.CError(e))
			return nil, e
		}
		go func() {
			if e := d.RedisSetIndex(ctx, userid, chatkey, msg, recall, 0); e != nil {
				log.Error(ctx, "[dao.GetIndex] update redis failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
			}
		}()
		r := &model.IMIndex{MsgIndex: msg, RecallIndex: recall}
		return unsafe.Pointer(r), nil
	})
	if e != nil {
		return nil, e
	}
	unsafeAck, e := oneshot.Do("GetIndex_ack_"+chatkey+"_"+userid, func() (unsafe.Pointer, error) {
		ack, e := d.MongoGetAck(ctx, userid, chatkey)
		if e != nil {
			log.Error(ctx, "[dao.GetIndex] db op failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
			return nil, e
		}
		//when ack is 0,don't need to update the redis,it is already updated in the last oneshot
		if ack != 0 {
			go func() {
				if e := d.RedisSetIndex(ctx, userid, chatkey, 0, 0, ack); e != nil {
					log.Error(ctx, "[dao.GetIndex] update redis failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
				}
			}()
		}
		return unsafe.Pointer(&ack), nil
	})
	if e != nil {
		return nil, e
	}
	r := (*model.IMIndex)(unsafeIndex)
	r.AckIndex = *(*uint32)(unsafeAck)
	return r, nil
}
func (d *Dao) Ack(ctx context.Context, userid, chatkey string, msgindex uint32) (e error) {
	if msgindex, e = d.MongoAck(ctx, userid, chatkey, msgindex); e != nil {
		log.Error(ctx, "[dao.Ack] db op failed",
			log.String("user_id", userid),
			log.String("chat_key", chatkey),
			log.Uint64("msg_index", uint64(msgindex)),
			log.CError(e))
		return e
	}
	if e = d.RedisSetIndex(ctx, userid, chatkey, 0, 0, msgindex); e != nil {
		log.Error(ctx, "[dao.Ack] redis op failed",
			log.String("user_id", userid),
			log.String("chat_key", chatkey),
			log.Uint64("msg_index", uint64(msgindex)),
			log.CError(e))
	}
	return
}
