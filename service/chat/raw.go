package chat

import (
	"context"
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/container/list"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/mids"
	"github.com/chenjie199234/Corelib/stream"
	"github.com/chenjie199234/Corelib/util/common"
)

type writer struct {
	p      *stream.Peer
	notice chan *struct{}
	list   *list.List[[]byte]
	count  uint64
}

func (w *writer) add(data []byte) {
	w.list.Push(data)
	count := atomic.AddUint64(&w.count, 1)
	if count == 1 {
		select {
		case <-w.notice:
		default:
		}
	} else if count > 512 {
		log.Error(nil, "[RawWriter] too many stacking msgs need to send",
			log.String("user_id", w.p.GetUniqueID()),
			log.String("remote_addr", w.p.GetRemoteAddr()),
			log.String("real_ip", w.p.GetRealPeerIP()))
		w.p.Close(false)
	} else if count > 128 {
		log.Warn(nil, "[RawWriter] too many stacking msgs need to send",
			log.String("user_id", w.p.GetUniqueID()),
			log.String("remote_addr", w.p.GetRemoteAddr()),
			log.String("real_ip", w.p.GetRealPeerIP()))
	}
}
func (w *writer) get() []byte {
	data, _ := w.list.Pop(nil)
	if data != nil {
		atomic.AddUint64(&w.count, math.MaxUint64)
	}
	return data
}
func (w *writer) run() {
	for {
		data := w.get()
		if data == nil {
			//empty wait for new data
			if _, ok := <-w.notice; !ok {
				//offline
				return
			}
			continue
		}
		if e := w.p.SendMessage(nil, data, nil, nil); e != nil {
			if e == stream.ErrConnClosed {
				//offline
				return
			}
			log.Error(nil, "[RawWriter] failed",
				log.String("user_id", w.p.GetUniqueID()),
				log.String("remote_addr", w.p.GetRemoteAddr()),
				log.String("real_ip", w.p.GetRealPeerIP()),
				log.String("msg", common.BTS(data)),
				log.CError(e))
		}
	}
}
func (s *Service) RawVerify(ctx context.Context, peerVerifyData []byte) (response []byte, uniqueid string, success bool) {
	t := mids.VerifyToken(ctx, common.BTS(peerVerifyData))
	if t == nil {
		log.Error(ctx, "[RawVerify] token verify failed", log.String("token", string(peerVerifyData)))
		return nil, "", false
	}
	if p := s.rawInstance.GetPeer(t.UserID); p != nil {
		//this user already connected
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
	w := &writer{
		p:      p,
		notice: make(chan *struct{}, 1),
		list:   list.NewList[[]byte](),
	}
	p.SetData(unsafe.Pointer(w))
	go w.run()
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
	w := (*writer)(p.GetData())
	close(w.notice)
	if e := s.chatDao.RedisDelSession(p, p.GetUniqueID(), p.GetRemoteAddr()); e != nil {
		log.Error(p, "[RawOffline] redis op failed",
			log.String("user_id", p.GetUniqueID()),
			log.String("remote_addr", p.GetRemoteAddr()),
			log.String("real_ip", p.GetRealPeerIP()),
			log.CError(e))
	}
}
