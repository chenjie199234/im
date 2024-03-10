package chat

import (
	"context"
	"time"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	chatdao "github.com/chenjie199234/im/dao/chat"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"
	"github.com/chenjie199234/im/util"

	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/log"
	// "github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/util/graceful"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service subservice for chat business
type Service struct {
	stop *graceful.Graceful

	chatDao *chatdao.Dao
}

// Start -
func Start() *Service {
	return &Service{
		stop: graceful.New(),

		chatDao: chatdao.NewDao(config.GetMysql("chat_mysql"), config.GetRedis("chat_redis"), config.GetMongo("chat_mongo")),
	}
}

func (s *Service) Send(ctx context.Context, req *api.SendReq) (*api.SendResp, error) {
	md := metadata.GetMetadata(ctx)
	sender, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[Send] token format wrong", log.String("sender", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	target, e := primitive.ObjectIDFromHex(req.Target)
	if e != nil {
		log.Error(ctx, "[Send] target format wrong", log.String("target", req.Target), log.String("target_type", req.TargetType), log.CError(e))
		return nil, ecode.ErrReq
	}
	//TODO check the relation
	chatkey := util.FormChatKey(md["Token-User"], req.Target, req.TargetType)
	msg := &model.MsgInfo{
		ChatKey: chatkey,
		Sender:  sender,
		Msg:     req.Msg,
		Extra:   req.Extra,
	}
	if e := s.chatDao.MongoSend(ctx, msg); e != nil {
		log.Error(ctx, "[Send] db op failed", log.String("sender", md["Token-User"]), log.String("chat_key", chatkey), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//TODO push
	return &api.SendResp{
		MsgIndex:  msg.MsgIndex,
		Timestamp: uint64(msg.ID.Timestamp().Unix()),
	}, nil
}
func (s *Service) Recall(ctx context.Context, req *api.RecallReq) (*api.RecallResp, error) {
	md := metadata.GetMetadata(ctx)
	recaller, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[Recall] token format wrong", log.String("recaller", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	target, e := primitive.ObjectIDFromHex(req.Target)
	if e != nil {
		log.Error(ctx, "[Recall] target format wrong", log.String("target", req.Target), log.String("target_type", req.TargetType), log.CError(e))
		return nil, ecode.ErrReq
	}
	//TODO check relation
	chatkey := util.FormChatKey(md["Token-User"], req.Target, req.TargetType)
	recallindex, e := s.chatDao.MongoRecall(ctx, recaller, chatkey, req.MsgIndex)
	if e != nil {
		log.Error(ctx, "[Recall] db op failed", log.String("recaller", md["Token-User"]), log.String("chat_key", chatkey), log.Uint64("msg_index", req.MsgIndex), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//TODO push
	return &api.RecallResp{RecallIndex: recallindex}, nil
}
func (s *Service) Ack(ctx context.Context, req *api.AckReq) (*api.AckResp, error) {
	md := metadata.GetMetadata(ctx)
	acker, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[Ack] token format wrong", log.String("acker", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	target, e := primitive.ObjectIDFromHex(req.Target)
	if e != nil {
		log.Error(ctx, "[Ack] target format wrong", log.String("target", req.Target), log.String("target_type", req.TargetType), log.CError(e))
		return nil, ecode.ErrReq
	}
	chatkey := util.FormChatKey(md["Token-User"], req.Target, req.TargetType)
	if req.TargetType == "group" {
		//TODO check relation
	}
	if e := s.chatDao.MongoAck(ctx, acker, chatkey, req.MsgIndex); e != nil {
		log.Error(ctx, "[Ack] db op failed", log.String("acker", md["Token-User"]), log.String("chat_key", chatkey), log.Uint64("msg_index", req.MsgIndex), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//TODO push if need
	return &api.AckResp{}, nil
}
func (s *Service) Pull(ctx context.Context, req *api.PullReq) (*api.PullResp, error) {
	md := metadata.GetMetadata(ctx)
	puller, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[Pull] token format wrong", log.String("puller", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	target, e := primitive.ObjectIDFromHex(req.Target)
	if e != nil {
		log.Error(ctx, "[Pull] target format wrong", log.String("target", req.Target), log.String("target_type", req.TargetType), log.CError(e))
		return nil, ecode.ErrReq
	}
	//TODO check relation
	resp := &api.PullResp{
		Msgs: make([]*api.MsgInfo, 0, req.Count),
	}
	chatkey := util.FormChatKey(md["Token-User"], req.Target, req.TargetType)
	mintimestamp := time.Now().Unix() - 30*24*60*60
	if req.StartMsgIndex != 0 {
		msgs, e := s.chatDao.MongoGetMsgs(ctx, chatkey, req.Direction, req.StartMsgIndex, req.Count, uint32(mintimestamp))
		if e != nil {
			log.Error(ctx, "[Pull] db op failed", log.String("puller", md["Token-User"]), log.String("chat_key", chatkey), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		}
		for _, msg := range msgs {
			resp.Msgs = append(resp.Msgs, &api.MsgInfo{
				MsgIndex:    msg.MsgIndex,
				RecallIndex: msg.RecallIndex,
				Msg:         msg.Msg,
				Extra:       msg.Extra,
				Timestamp:   uint64(msg.ID.Timestamp().Unix()),
				Sender:      msg.Sender.Hex(),
			})
		}
	}
	if req.StartRecallIndex != 0 {
		recalls, e := s.chatDao.MongoGetRecalls(ctx, chatkey, req.Direction, req.StartRecallIndex, req.Count, uint32(mintimestamp))
		if e != nil {
			log.Error(ctx, "[Pull] db op failed", log.String("puller", md["Token-User"]), log.String("chat_key", chatkey), log.CError(e))
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
