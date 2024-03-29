package chat

import (
	"context"
	"time"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	chatdao "github.com/chenjie199234/im/dao/chat"
	relationDao "github.com/chenjie199234/im/dao/relation"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"
	"github.com/chenjie199234/im/util"

	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/log"
	// "github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/log/trace"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/util/graceful"
)

// Service subservice for chat business
type Service struct {
	stop *graceful.Graceful

	chatDao     *chatdao.Dao
	relationDao *relationDao.Dao
}

// Start -
func Start() *Service {
	s := &Service{
		stop: graceful.New(),

		chatDao:     chatdao.NewDao(nil, config.GetRedis("im_redis"), config.GetMongo("im_mongo")),
		relationDao: relationDao.NewDao(nil, config.GetRedis("im_redis"), config.GetMongo("im_mongo")),
	}
	return s
}

func (s *Service) Send(ctx context.Context, req *api.SendReq) (*api.SendResp, error) {
	md := metadata.GetMetadata(ctx)
	sender := md["Token-User"]
	//check the relation in target's view
	if req.TargetType == "user" {
		if _, e := s.relationDao.GetUserRelation(ctx, req.Target, sender, "user"); e != nil {
			log.Error(ctx, "[Send] check relation failed",
				log.String("sender", sender),
				log.String("target", req.Target),
				log.String("target_type", req.TargetType),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		}
	} else if _, e := s.relationDao.GetGroupMember(ctx, req.Target, sender); e != nil {
		log.Error(ctx, "[Send] check relation failed",
			log.String("sender", sender),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.stop.Add(1); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	//check the rate limit
	if e := s.chatDao.RedisSendRecallRate(ctx, sender); e != nil {
		log.Error(ctx, "[Send] rate check failed", log.String("sender", sender), log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	msg := &model.MsgInfo{
		ChatKey: util.FormChatKey(sender, req.Target, req.TargetType),
		Sender:  sender,
		Msg:     req.Msg,
		Extra:   req.Extra,
	}
	if e := s.chatDao.MongoSend(ctx, msg); e != nil {
		log.Error(ctx, "[Send] db op failed", log.String("sender", sender), log.String("chat_key", msg.ChatKey), log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push
	go func() {
		s.pushsend(trace.CloneSpan(ctx), msg, req.Target, req.TargetType)
		s.stop.DoneOne()
	}()
	return &api.SendResp{
		MsgIndex:  msg.MsgIndex,
		Timestamp: uint32(msg.ID.Timestamp().Unix()),
	}, nil
}
func (s *Service) Recall(ctx context.Context, req *api.RecallReq) (*api.RecallResp, error) {
	md := metadata.GetMetadata(ctx)
	recaller := md["Token-User"]
	//check the relation in self's view
	if _, e := s.relationDao.GetUserRelation(ctx, recaller, req.Target, req.TargetType); e != nil {
		log.Error(ctx, "[Recall] check relation failed",
			log.String("recaller", recaller),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.stop.Add(1); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	//check the rate limit
	if e := s.chatDao.RedisSendRecallRate(ctx, recaller); e != nil {
		log.Error(ctx, "[Recall] rate check failed", log.String("recaller", recaller), log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	chatkey := util.FormChatKey(recaller, req.Target, req.TargetType)
	recallindex, e := s.chatDao.MongoRecall(ctx, recaller, chatkey, req.MsgIndex)
	if e != nil {
		log.Error(ctx, "[Recall] db op failed",
			log.String("recaller", recaller),
			log.String("chat_key", chatkey),
			log.Uint64("msg_index", uint64(req.MsgIndex)),
			log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push
	go func() {
		s.pushrecall(trace.CloneSpan(ctx), recallindex, req.MsgIndex, chatkey, recaller, req.Target, req.TargetType)
		s.stop.DoneOne()
	}()
	return &api.RecallResp{RecallIndex: recallindex}, nil
}
func (s *Service) Ack(ctx context.Context, req *api.AckReq) (*api.AckResp, error) {
	md := metadata.GetMetadata(ctx)
	acker := md["Token-User"]
	//check the relation in self's view
	if _, e := s.relationDao.GetUserRelation(ctx, acker, req.Target, req.TargetType); e != nil {
		log.Error(ctx, "[Ack] check relation failed",
			log.String("acker", acker),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	chatkey := util.FormChatKey(acker, req.Target, req.TargetType)
	//check the max msg index
	if index, e := s.chatDao.GetIndex(ctx, acker, chatkey); e != nil {
		log.Error(ctx, "[Ack] check index failed",
			log.String("acker", acker),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if index.MsgIndex < req.MsgIndex {
		return nil, ecode.ErrMsgNotExist
	} else if index.AckIndex >= req.MsgIndex {
		//already acked
		return &api.AckResp{}, nil
	}
	if e := s.chatDao.Ack(ctx, acker, chatkey, req.MsgIndex); e != nil {
		log.Error(ctx, "[Ack] dao op failed",
			log.String("acker", acker),
			log.String("chat_key", chatkey),
			log.Uint64("msg_index", uint64(req.MsgIndex)),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//TODO push if need
	return &api.AckResp{}, nil
}
func (s *Service) Pull(ctx context.Context, req *api.PullReq) (*api.PullResp, error) {
	md := metadata.GetMetadata(ctx)
	puller := md["Token-User"]
	//check the relation in self's view
	if _, e := s.relationDao.GetUserRelation(ctx, puller, req.Target, req.TargetType); e != nil {
		log.Error(ctx, "[Pull] check relation failed",
			log.String("puller", puller),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	chatkey := util.FormChatKey(puller, req.Target, req.TargetType)
	//check the rate limit
	if e := s.chatDao.RedisPullRate(ctx, puller, chatkey); e != nil {
		log.Error(ctx, "[Pull] rate check failed", log.String("puller", puller), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	resp := &api.PullResp{
		Msgs: make([]*api.MsgInfo, 0, req.Count),
	}
	mintimestamp := time.Now().Unix() - 30*24*60*60
	if req.StartMsgIndex != 0 {
		msgs, e := s.chatDao.MongoGetMsgs(ctx, chatkey, req.Direction, req.StartMsgIndex, req.Count, uint32(mintimestamp))
		if e != nil {
			log.Error(ctx, "[Pull] db op failed", log.String("puller", puller), log.String("chat_key", chatkey), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		}
		for _, msg := range msgs {
			resp.Msgs = append(resp.Msgs, &api.MsgInfo{
				MsgIndex:    msg.MsgIndex,
				RecallIndex: msg.RecallIndex,
				Msg:         msg.Msg,
				Extra:       msg.Extra,
				Timestamp:   uint32(msg.ID.Timestamp().Unix()),
				Sender:      msg.Sender,
			})
		}
	}
	if req.StartRecallIndex != 0 {
		recalls, e := s.chatDao.MongoGetRecalls(ctx, chatkey, req.Direction, req.StartRecallIndex, req.Count, uint32(mintimestamp))
		if e != nil {
			log.Error(ctx, "[Pull] db op failed", log.String("puller", puller), log.String("chat_key", chatkey), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		}
		resp.Recalls = recalls
	}
	return resp, nil
}

// Stop -
func (s *Service) Stop() {
	s.stop.Close(nil, nil)
}
