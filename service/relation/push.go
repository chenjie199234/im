package relation

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/util"

	"github.com/chenjie199234/Corelib/log"
)

// action can be userRelationAdd,userRelationDel
func (s *Service) pushuser(ctx context.Context, receiver, action, target, targetType, name string) {
	if action != "userRelationAdd" && action != "userRelationDel" {
		return
	}
	tmp := &util.MQ{
		Type: action,
		Content: &util.UserAction{
			Target:     target,
			TargetType: targetType,
			Name:       name,
		},
	}
	if action == "userRelationAdd" {
		chatkey := util.FormChatKey(receiver, target, targetType)
		index, e := s.chatDao.GetIndex(ctx, receiver, chatkey)
		if e != nil {
			log.Error(ctx, "[PushUser] get chat index failed", log.String("user_id", receiver), log.String("chat_key", chatkey), log.CError(e))
			return
		}
		ua := tmp.Content.(*util.UserAction)
		ua.MsgIndex = index.MsgIndex
		ua.RecallIndex = index.RecallIndex
		ua.AckIndex = index.AckIndex
	}
	pushmsg, _ := json.Marshal(tmp)
	if e := util.Unicast(ctx, receiver, pushmsg); e != nil && e != ecode.ErrSession {
		log.Error(ctx, "[PushUser] push mq failed",
			log.String("receiver", receiver),
			log.String("action", action),
			log.String("target", target),
			log.String("target_type", targetType),
			log.String("name", name),
			log.CError(e))
	}
}

// action can be groupJoin,groupLeave,groupKick
func (s *Service) pushgroup(ctx context.Context, groupid, action, actionmember, name string, duty uint8) {
	if action != "groupJoin" && action != "groupLeave" && action != "groupKick" {
		return
	}
	members, e := s.relationDao.GetGroupMembers(ctx, groupid)
	if e != nil {
		log.Error(ctx, "[PushGroup] get group members failed", log.String("group_id", groupid), log.CError(e))
		return
	}
	wg := &sync.WaitGroup{}
	pushmsg, _ := json.Marshal(&util.MQ{
		Type: action,
		Content: &util.GroupAction{
			Member: actionmember,
			Name:   name,
			Duty:   duty,
		},
	})
	for _, v := range members {
		member := v
		if member.Target == "" {
			continue
		}
		wg.Add(1)
		util.AddWork(func() {
			if e := util.Unicast(ctx, member.Target, pushmsg); e != nil && e != ecode.ErrSession {
				log.Error(ctx, "[PushGroup] push mq failed",
					log.String("group_id", groupid),
					log.String("receive_member", member.Target),
					log.String("action", action),
					log.String("action_member", actionmember),
					log.String("name", name),
					log.Uint64("duty", uint64(duty)),
					log.CError(e))
			}
			wg.Done()
		})
	}
	wg.Wait()
}
