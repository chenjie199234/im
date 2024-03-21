package chat

import (
	"context"
	"time"

	"github.com/chenjie199234/im/ecode"
)

func (d *Dao) RedisSendRecallRate(ctx context.Context, userid string) error {
	//1 seconds do once
	status, e := d.imredis.SetNX(ctx, "send_recall_rate_{"+userid+"}", 1, time.Second).Result()
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
func (d *Dao) RedisPullRate(ctx context.Context, userid, chatkey string) error {
	//5 seconds do once for per chatkey
	//1 seconds do twice for all chatkey
	status, e := d.imredis.RateLimit(ctx, map[string][2]uint64{
		"pull_rate_{" + userid + "}_" + chatkey: {1, 5},
		"pull_rate_{" + userid + "}":            {2, 1},
	})
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
