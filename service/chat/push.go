package chat

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"
	"github.com/chenjie199234/im/util"

	"github.com/chenjie199234/Corelib/log"
)

func (s *Service) pushsend(ctx context.Context, msg *model.MsgInfo, target, targetType string) {
	pushmsg, _ := json.Marshal(&util.MQ{
		Type: "msg",
		Content: &util.Msg{
			Sender:    msg.Sender,
			Msg:       msg.Msg,
			Extra:     msg.Extra,
			MsgIndex:  msg.MsgIndex,
			Timestamp: uint32(msg.ID.Timestamp().Unix()),
		},
	})
	single := func(userid string) {
		if userid != msg.Sender {
			if e := util.Unicast(ctx, userid, pushmsg); e != nil && e != ecode.ErrSession {
				log.Error(ctx, "[PushSend] push mq failed", log.String("user_id", userid), log.String("chat_key", msg.ChatKey), log.CError(e))
			}
		}
		var e error
		if userid == msg.Sender {
			e = s.chatDao.RedisSetIndex(ctx, userid, msg.ChatKey, msg.MsgIndex, math.MaxUint32, msg.MsgIndex)
		} else {
			e = s.chatDao.RedisSetIndex(ctx, userid, msg.ChatKey, msg.MsgIndex, math.MaxUint32, math.MaxUint32)
		}
		if e != nil {
			time.Sleep(time.Millisecond * 10)
			if e = s.chatDao.RedisDelIndex(ctx, userid, msg.ChatKey); e != nil {
				log.Error(ctx, "[PushSend] update index failed", log.String("user_id", userid), log.String("chat_key", msg.ChatKey), log.CError(e))
			}
		}
	}
	if targetType == "user" {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		util.AddWork(func() { single(msg.Sender); wg.Done() })
		util.AddWork(func() { single(target); wg.Done() })
		wg.Wait()
	} else if members, e := s.relationDao.GetGroupMembers(ctx, target); e != nil {
		log.Error(ctx, "[PushSend] get group members failed", log.String("group_id", target), log.CError(e))
		return
	} else {
		wg := &sync.WaitGroup{}
		wg.Add(len(members))
		for _, v := range members {
			member := v
			if member.Target == "" {
				continue
			}
			util.AddWork(func() { single(member.Target); wg.Done() })
		}
		wg.Wait()
	}
}

func (s *Service) pushrecall(ctx context.Context, recallindex, msgindex uint32, chatkey, recaller, target, targetType string) {
	pushmsg, _ := json.Marshal(&util.MQ{
		Type: "recall",
		Content: &util.Recall{
			MsgIndex:    msgindex,
			RecallIndex: recallindex,
		},
	})
	single := func(userid string) {
		if userid != recaller {
			if e := util.Unicast(ctx, userid, pushmsg); e != nil && e != ecode.ErrSession {
				log.Error(ctx, "[PushRecall] push mq failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
			}
		}
		if e := s.chatDao.RedisSetIndex(ctx, userid, chatkey, math.MaxUint32, recallindex, math.MaxUint32); e != nil {
			time.Sleep(time.Millisecond * 10)
			if e = s.chatDao.RedisDelIndex(ctx, userid, chatkey); e != nil {
				log.Error(ctx, "[PushRecall] redis op failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
			}
		}
	}
	if targetType == "user" {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		util.AddWork(func() { single(recaller); wg.Done() })
		util.AddWork(func() { single(target); wg.Done() })
		wg.Wait()
	} else if members, e := s.relationDao.GetGroupMembers(ctx, target); e != nil {
		log.Error(ctx, "[PushRecall] get group members failed", log.String("group_id", target), log.CError(e))
		return
	} else {
		wg := &sync.WaitGroup{}
		wg.Add(len(members))
		for _, v := range members {
			member := v
			if member.Target == "" {
				continue
			}
			util.AddWork(func() { single(member.Target); wg.Done() })
		}
		wg.Wait()
	}
}
