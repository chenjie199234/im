package relation

import (
	"context"
	"unsafe"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/log/trace"
	cmongo "github.com/chenjie199234/Corelib/mongo"
	cmysql "github.com/chenjie199234/Corelib/mysql"
	credis "github.com/chenjie199234/Corelib/redis"
	"github.com/chenjie199234/Corelib/util/oneshot"
	gredis "github.com/redis/go-redis/v9"
)

// Dao this is a data operation layer to operate relation service's data
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

// ------------------------------------------user-----------------------------------
func (d *Dao) UpdateUserRelationName(ctx context.Context, userid, target, targetType, newname string) error {
	now, e := d.MongoUpdateUserRelationName(ctx, userid, target, targetType, newname)
	if e != nil {
		log.Error(ctx, "[dao.UpdateUserRelationName] db op failed",
			log.String("user_id", userid),
			log.String("target", target),
			log.String("target_type", targetType),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	if e = d.RedisAddUserRelation(ctx, userid, target, targetType, now.Name); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.UpdateUserRelationName] redis op failed",
				log.String("user_id", userid),
				log.String("target", target),
				log.String("target_type", targetType),
				log.String("name", newname),
				log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) UpdateNameInGroup(ctx context.Context, userid, groupid, newname string) error {
	now, e := d.MongoUpdateNameInGroup(ctx, userid, groupid, newname)
	if e != nil {
		log.Error(ctx, "[dao.UpdateNameInGroup] db op failed",
			log.String("user_id", userid),
			log.String("group_id", groupid),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	if e = d.RedisAddGroupMember(ctx, groupid, userid, now.Name, now.Duty); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.UpdateNameInGroup] redis op failed",
				log.String("user_id", userid),
				log.String("group_id", groupid),
				log.String("name", newname),
				log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) AcceptMakeFriend(ctx context.Context, userid, friendid string) error {
	username, friendname, e := d.MongoAcceptMakeFriend(ctx, userid, friendid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptMakeFriend] db op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
		return e
	}
	// friend's relation should be add first,because the user will retry
	if e := d.RedisAddUserRelation(ctx, friendid, userid, "user", username); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, friendid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptMakeFriend] redis op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddUserRelation(ctx, userid, friendid, "user", friendname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptMakeFriend] redis op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) AcceptGroupInvite(ctx context.Context, userid, groupid string) error {
	username, groupname, e := d.MongoAcceptGroupInvite(ctx, userid, groupid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptGroupInvite] db op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
		return e
	}
	// group's relation should be add first,because the user will retry
	if e := d.RedisAddGroupMember(ctx, groupid, userid, username, 0); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupInvite] redis op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddUserRelation(ctx, userid, groupid, "group", groupname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupInvite] redis op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) DelFriend(ctx context.Context, userid, friendid string) error {
	if e := d.MongoDelFriend(ctx, userid, friendid); e != nil {
		log.Error(ctx, "[dao.DelFriend] db op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
		return e
	}
	// friend's relation should be deleted first,because the user will retry
	if e := d.RedisDelUserRelation(ctx, friendid, userid, "user"); e != nil {
		log.Error(ctx, "[dao.DelFriend] redis op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
		return e
	}
	if e := d.RedisDelUserRelation(ctx, userid, friendid, "user"); e != nil {
		log.Error(ctx, "[dao.DelFriend] redis op failed", log.String("user_id", userid), log.String("friend_id", friendid), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) LeaveGroup(ctx context.Context, userid, groupid string) error {
	if e := d.MongoLeaveGroup(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] db op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
		return e
	}
	// group's relation should be deleted first,because the user will retry
	if e := d.RedisDelGroupMember(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] redis op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
		return e
	}
	if e := d.RedisDelUserRelation(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] redis op failed", log.String("user_id", userid), log.String("group_id", groupid), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) SetUserName(ctx context.Context, userid, newname string) error {
	if e := d.MongoSetUserName(ctx, userid, newname); e != nil {
		log.Error(ctx, "[dao.SetUserName] db op failed", log.String("user_id", userid), log.CError(e))
		return e
	}
	if e := d.RedisAddUserRelation(ctx, userid, "", "user", newname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.SetUserName] redis op failed", log.String("user_id", userid), log.String("name", newname))
			return e
		}
	}
	return nil
}
func (d *Dao) GetUserName(ctx context.Context, userid string) (string, error) {
	if relation, e := d.RedisGetUserRelation(ctx, userid, "", "user"); e == nil {
		return relation.Name, nil
	} else if e == ecode.ErrUserNotExist {
		return "", e
	} else if e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserName] redis op failed", log.String("user_id", userid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid, func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserName] db op failed", log.String("user_id", userid), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(trace.CloneSpan(ctx), userid, relations); e != nil {
				log.Error(ctx, "[dao.GetUserName] update redis failed", log.String("user_id", userid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return "", e
	}
	relations := *(*[]*model.RelationTarget)(unsafeRelations)
	for _, relation := range relations {
		if relation.Target == "" {
			return relation.Name, nil
		}
	}
	//this is impossible
	return "", ecode.ErrUserNotExist
}

func (d *Dao) GetUserRelations(ctx context.Context, userid string) ([]*model.RelationTarget, error) {
	if relations, e := d.RedisGetUserRelations(ctx, userid); e == nil || e == ecode.ErrUserNotExist {
		return relations, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserRelations] redis op failed", log.String("user_id", userid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid, func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserRelations] db op failed", log.String("user_id", userid), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(trace.CloneSpan(ctx), userid, relations); e != nil {
				log.Error(ctx, "[dao.GetUserRelations] update redis failed", log.String("user_id", userid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return nil, e
	}
	return *(*[]*model.RelationTarget)(unsafeRelations), nil
}
func (d *Dao) CountUserRelations(ctx context.Context, userid, exceptTarget, exceptTargetType string) (uint64, error) {
	if count, e := d.RedisCountUserRelations(ctx, userid, exceptTarget, exceptTargetType); e == nil || e == ecode.ErrUserNotExist {
		return count, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.CountUserRelations] redis op failed", log.String("user_id", userid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid, func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.CountUserRelations] db op failed", log.String("user_id", userid), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(trace.CloneSpan(ctx), userid, relations); e != nil {
				log.Error(ctx, "[dao.CountUserRelations] update redis failed", log.String("user_id", userid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return 0, e
	}
	relations := *(*[]*model.RelationTarget)(unsafeRelations)
	if exceptTarget != "" {
		for _, relation := range relations {
			if relation.Target == exceptTarget && relation.TargetType == exceptTargetType {
				return uint64(len(relations)) - 2, nil
			}
		}
	}
	return uint64(len(relations)) - 1, nil
}
func (d *Dao) GetUserRelation(ctx context.Context, userid, target, targetType string) (*model.RelationTarget, error) {
	if targetType != "user" && targetType != "group" {
		return nil, ecode.ErrReq
	}
	if relation, e := d.RedisGetUserRelation(ctx, userid, target, targetType); e == nil || e == ecode.ErrUserNotExist || e == ecode.ErrNotFriends || e == ecode.ErrNotInGroup {
		return relation, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserRelation] redis op failed",
			log.String("user_id", userid),
			log.String("target", target),
			log.String("target_type", targetType),
			log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid, func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserRelation] db op failed", log.String("user_id", userid), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(trace.CloneSpan(ctx), userid, relations); e != nil {
				log.Error(ctx, "[dao.GetUserRelation] update redis failed", log.String("user_id", userid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return nil, e
	}
	relations := *(*[]*model.RelationTarget)(unsafeRelations)
	for _, relation := range relations {
		if relation.Target == target && relation.TargetType == targetType {
			return relation, nil
		}
	}
	if targetType == "user" {
		return nil, ecode.ErrNotFriends
	}
	return nil, ecode.ErrNotInGroup
}

// ------------------------------------------group-----------------------------------
func (d *Dao) UpdateDutyInGroup(ctx context.Context, userid, groupid string, newduty uint8) error {
	now, e := d.MongoUpdateDutyInGroup(ctx, userid, groupid, newduty)
	if e != nil {
		log.Error(ctx, "[dao.UpdateDutyInGroup] db op failed",
			log.String("user_id", userid),
			log.String("group_id", groupid),
			log.Uint64("duty", uint64(newduty)),
			log.CError(e))
		return e
	}
	if e = d.RedisAddGroupMember(ctx, groupid, userid, now.Name, now.Duty); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.UpdateDutyInGroup] redis op failed",
				log.String("user_id", userid),
				log.String("group_id", groupid),
				log.Uint64("duty", uint64(newduty)),
				log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) AcceptGroupApply(ctx context.Context, groupid, userid string) error {
	username, groupname, e := d.MongoAcceptGroupApply(ctx, groupid, userid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptGroupApply] db op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
		return e
	}
	// user's relation should be deleted first,because the group's admin will retry
	if e := d.RedisAddUserRelation(ctx, userid, groupid, "group", groupname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupApply] redis op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddGroupMember(ctx, groupid, userid, username, 0); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupApply] redis op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) KickGroup(ctx context.Context, groupid, userid string) error {
	if e := d.MongoKickGroup(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[dao.KickGroup] db op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
		return e
	}
	// user's relation should be deleted first,because the group's admin will retry
	if e := d.RedisDelUserRelation(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[dao.KickGroup] redis op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
		return e
	}
	if e := d.RedisDelGroupMember(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[dao.KickGroup] redis op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) SetGroupName(ctx context.Context, groupid, name string) error {
	if e := d.MongoSetGroupName(ctx, groupid, name); e != nil {
		log.Error(ctx, "[dao.SetGroupName] db op failed", log.String("group_id", groupid), log.String("name", name), log.CError(e))
		return e
	}
	if e := d.RedisAddGroupMember(ctx, groupid, "", name, 1); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.SetGroupName] db op failed", log.String("group_id", groupid), log.String("name", name))
			return e
		}
	}
	return nil
}
func (d *Dao) GetGroupName(ctx context.Context, groupid string) (string, error) {
	if member, e := d.RedisGetGroupMember(ctx, groupid, ""); e == nil {
		return member.Name, nil
	} else if e == ecode.ErrGroupNotExist {
		return "", e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupName] redis op failed", log.String("group_id", groupid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid, func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupName] db op failed", log.String("group_id", groupid), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(trace.CloneSpan(ctx), groupid, members); e != nil {
				log.Error(ctx, "[dao.GetGroupName] update redis failed", log.String("group_id", groupid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return "", e
	}
	members := *(*[]*model.RelationTarget)(unsafeMembers)
	for _, member := range members {
		if member.Target == "" {
			return member.Name, nil
		}
	}
	//this is impossible
	return "", ecode.ErrGroupNotExist
}
func (d *Dao) GetGroupMembers(ctx context.Context, groupid string) ([]*model.RelationTarget, error) {
	if members, e := d.RedisGetGroupMembers(ctx, groupid); e == nil || e == ecode.ErrGroupNotExist {
		return members, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupMembers] redis op failed", log.String("group_id", groupid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid, func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupMembers] db op failed", log.String("group_id", groupid), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(trace.CloneSpan(ctx), groupid, members); e != nil {
				log.Error(ctx, "[dao.GetGroupMembers] update redis failed", log.String("group_id", groupid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return nil, e
	}
	return *(*[]*model.RelationTarget)(unsafeMembers), nil
}
func (d *Dao) CountGroupMembers(ctx context.Context, groupid, exceptMember string) (uint64, error) {
	if count, e := d.RedisCountGroupMembers(ctx, groupid, exceptMember); e == nil || e == ecode.ErrGroupNotExist {
		return count, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.CountGroupMembers] redis op failed", log.String("group_id", groupid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid, func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.CountGroupMembers] db op failed", log.String("group_id", groupid), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(trace.CloneSpan(ctx), groupid, members); e != nil {
				log.Error(ctx, "[dao.CountGroupMembers] update redis failed", log.String("group_id", groupid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return 0, e
	}
	members := *(*[]*model.RelationTarget)(unsafeMembers)
	if exceptMember != "" {
		for _, member := range members {
			if member.Target == exceptMember {
				return uint64(len(members)) - 2, nil
			}
		}
	}
	return uint64(len(members)) - 1, nil
}
func (d *Dao) GetGroupMember(ctx context.Context, groupid, userid string) (*model.RelationTarget, error) {
	if member, e := d.RedisGetGroupMember(ctx, groupid, userid); e == nil || e == ecode.ErrGroupNotExist || e == ecode.ErrGroupMemberNotExist {
		return member, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupMember] redis op failed", log.String("group_id", groupid), log.String("user_id", userid), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid, func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupMember] db op failed", log.String("group_id", groupid), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(trace.CloneSpan(ctx), groupid, members); e != nil {
				log.Error(ctx, "[dao.GetGroupMember] update redis failed", log.String("group_id", groupid), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return nil, e
	}
	members := *(*[]*model.RelationTarget)(unsafeMembers)
	for _, member := range members {
		if member.Target == userid {
			return member, nil
		}
	}
	return nil, ecode.ErrGroupMemberNotExist
}
