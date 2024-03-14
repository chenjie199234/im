package chat

import (
	"context"

	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/mids"
	"github.com/chenjie199234/Corelib/stream"
	"github.com/chenjie199234/Corelib/util/common"
)

func (s *Service) RawVerify(ctx context.Context, peerVerifyData []byte) (response []byte, uniqueid string, success bool) {
	t := mids.VerifyToken(ctx, common.BTS(peerVerifyData))
	if t == nil {
		log.Error(ctx, "[RawVerify] token verify failed", log.String("token", string(peerVerifyData)))
		return nil, "", false
	}
	if p := s.rawInstance.GetPeer(t.UserID); p != nil {
		// this user already connected
		//this new connection will kick the old connection
		p.Close(true)
	}
	return nil, t.UserID, true
}
func (s *Service) RawOnline(p *stream.Peer) (success bool) {
	//send first ping to get the netlag from pong as soon as possible
	if e := p.SendPing(); e != nil {
		log.Error(p, "[RawOnline] send first ping failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		return false
	}
	if e := s.chatDao.RedisSetSession(p, p.GetUniqueID(), p.GetRemoteAddr(), p.GetRealPeerIP(), s.rawname); e != nil {
		log.Error(p, "[RawOnline] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		return false
	}
	return true
}
func (s *Service) RawPingPong(p *stream.Peer) {
	if e := s.chatDao.RedisUpdateSession(p, p.GetUniqueID(), p.GetRemoteAddr(), p.GetNetlag()); e != nil {
		log.Error(p, "[RawPingPong] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
		if e == ecode.ErrSession {
			//be kicked or session expired
			p.Close(false)
		}
	}
}
func (s *Service) RawUser(p *stream.Peer, userdata []byte) {
	//client will not send message to server
	p.Close(false)
}
func (s *Service) RawOffline(p *stream.Peer) {
	if e := s.chatDao.RedisDelSession(p, p.GetUniqueID(), p.GetRemoteAddr()); e != nil {
		log.Error(p, "[RawOffline] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
	}
}
