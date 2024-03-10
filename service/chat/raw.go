package chat

import (
	"context"

	"github.com/chenjie199234/Corelib/stream"
)

func (s *Service) RawVerify(ctx context.Context, peerVerifyData []byte, p *stream.Peer) (response []byte, success bool) {

}
func (s *Service) RawOnline(p *stream.Peer) (success bool) {

}
func (s *Service) RawPingPong(p *stream.Peer) {

}
func (s *Service) RawUser(p *stream.Peer, userdata []byte) {

}
func (s *Service) RawOffline(p *stream.Peer) {

}
