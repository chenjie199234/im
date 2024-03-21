package relation

import (
	"context"
	"time"

	"github.com/chenjie199234/im/ecode"
)

func (d *Dao) RedisUpdateUserRelationRate(ctx context.Context, userid string) error {
	//1 seconds do once
	status, e := d.imredis.SetNX(ctx, "update_user_relation_rate_{"+userid+"}", 1, time.Second).Result()
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
func (d *Dao) RedisUpdateGroupRelationRate(ctx context.Context, groupid string) error {
	//1 seconds do once
	status, e := d.imredis.SetNX(ctx, "update_group_relation_rate_{"+groupid+"}", 1, time.Second).Result()
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
func (d *Dao) RedisGetUserRelationsRate(ctx context.Context, userid string) error {
	// 1 minute do 5 times
	status, e := d.imredis.RateLimit(ctx, map[string][2]uint64{"get_user_relations_rate_{" + userid + "}": {5, 60}})
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
func (d *Dao) RedisGetGroupMembersRate(ctx context.Context, userid, groupid string) error {
	//1 minute do 5 times
	status, e := d.imredis.RateLimit(ctx, map[string][2]uint64{"get_group_members_rate_{" + userid + "}_" + groupid: {5, 60}})
	if e != nil {
		return e
	}
	if !status {
		return ecode.ErrTooFast
	}
	return nil
}
