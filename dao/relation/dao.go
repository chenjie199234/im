package relation

import (
	"context"
	"unsafe"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/log"
	cmongo "github.com/chenjie199234/Corelib/mongo"
	cmysql "github.com/chenjie199234/Corelib/mysql"
	credis "github.com/chenjie199234/Corelib/redis"
	"github.com/chenjie199234/Corelib/util/oneshot"
	gredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Dao this is a data operation layer to operate relation service's data
type Dao struct {
	mysql *cmysql.Client
	redis *credis.Client
	mongo *cmongo.Client
}

// NewDao Dao is only a data operation layer
// don't write business logic in this package
// business logic should be written in service package
func NewDao(mysql *cmysql.Client, redis *credis.Client, mongo *cmongo.Client) *Dao {
	return &Dao{
		mysql: mysql,
		redis: redis,
		mongo: mongo,
	}
}

// ------------------------------------------user-----------------------------------
func (d *Dao) UpdateUserRelationName(ctx context.Context, userid, target primitive.ObjectID, targetType, newname string) error {
	now, e := d.MongoUpdateUserRelationName(ctx, userid, target, targetType, newname)
	if e != nil {
		log.Error(ctx, "[dao.UpdateUserRelationName] db op failed",
			log.String("user_id", userid.Hex()),
			log.String("target", target.Hex()),
			log.String("target_type", targetType),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	if e = d.RedisRefreshUserRelation(ctx, userid, target, targetType, now.Name); e != nil {
		log.Error(ctx, "[dao.UpdateUserRelationName] redis op failed",
			log.String("user_id", userid.Hex()),
			log.String("target", target.Hex()),
			log.String("target_type", targetType),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) UpdateNameInGroup(ctx context.Context, userid, groupid primitive.ObjectID, newname string) error {
	now, e := d.MongoUpdateNameInGroup(ctx, userid, groupid, newname)
	if e != nil {
		log.Error(ctx, "[dao.UpdateNameInGroup] db op failed",
			log.String("user_id", userid.Hex()),
			log.String("group_id", groupid.Hex()),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	if e = d.RedisRefreshGroupMember(ctx, groupid, userid, now.Name, now.Duty); e != nil {
		log.Error(ctx, "[dao.UpdateNameInGroup] redis op failed",
			log.String("user_id", userid.Hex()),
			log.String("group_id", groupid.Hex()),
			log.String("name", newname),
			log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) AcceptMakeFriend(ctx context.Context, userid, friendid primitive.ObjectID) error {
	username, friendname, e := d.MongoAcceptMakeFriend(ctx, userid, friendid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptMakeFriend] db op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
		return e
	}
	// friend's relation should be add first,because the user will retry
	if e := d.RedisAddUserRelation(ctx, friendid, userid, "user", username); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, friendid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptMakeFriend] redis op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddUserRelation(ctx, userid, friendid, "user", friendname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptMakeFriend] redis op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) AcceptGroupInvite(ctx context.Context, userid, groupid primitive.ObjectID) error {
	username, groupname, e := d.MongoAcceptGroupInvite(ctx, userid, groupid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptGroupInvite] db op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
		return e
	}
	// group's relation should be add first,because the user will retry
	if e := d.RedisAddGroupMember(ctx, groupid, userid, username, 0); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupInvite] redis op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddUserRelation(ctx, userid, groupid, "group", groupname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupInvite] redis op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) DelFriend(ctx context.Context, userid, friendid primitive.ObjectID) error {
	if e := d.MongoDelFriend(ctx, userid, friendid); e != nil {
		log.Error(ctx, "[dao.DelFriend] db op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
		return e
	}
	// friend's relation should be deleted first,because the user will retry
	if e := d.RedisDelUserRelation(ctx, friendid, userid, "user"); e != nil {
		log.Error(ctx, "[dao.DelFriend] redis op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
		return e
	}
	if e := d.RedisDelUserRelation(ctx, userid, friendid, "user"); e != nil {
		log.Error(ctx, "[dao.DelFriend] redis op failed", log.String("user_id", userid.Hex()), log.String("friend_id", friendid.Hex()), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) LeaveGroup(ctx context.Context, userid, groupid primitive.ObjectID) error {
	if e := d.MongoLeaveGroup(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] db op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
		return e
	}
	// group's relation should be deleted first,because the user will retry
	if e := d.RedisDelGroupMember(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] redis op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
		return e
	}
	if e := d.RedisDelUserRelation(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[dao.LeaveGroup] redis op failed", log.String("user_id", userid.Hex()), log.String("group_id", groupid.Hex()), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) SetUserName(ctx context.Context, userid primitive.ObjectID, name string) error {
	if e := d.MongoSetUserName(ctx, userid, name); e != nil {
		log.Error(ctx, "[dao.SetUserName] db op failed", log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	if e := d.RedisDelUserName(context.Background(), userid); e != nil {
		log.Error(ctx, "[dao.SetUserName] redis op failed", log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) GetUserName(ctx context.Context, userid primitive.ObjectID) (string, error) {
	if userid.IsZero() {
		return "", ecode.ErrReq
	}
	if username, e := d.RedisGetUserName(ctx, userid); e == nil || e == ecode.ErrUserNotExist {
		return username, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserName] redis op failed", log.String("user_id", userid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeUserName, e := oneshot.Do("GetUserName_"+userid.Hex(), func() (unsafe.Pointer, error) {
		username, e := d.MongoGetUserName(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserName] db op failed", log.String("user_id", userid.Hex()), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserName(context.Background(), userid, username); e != nil {
				log.Error(ctx, "[dao.GetUserName] update redis failed", log.String("user_id", userid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&username), e
	})
	if e != nil {
		return "", e
	}
	return *(*string)(unsafeUserName), nil
}

func (d *Dao) GetUserRelations(ctx context.Context, userid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if userid.IsZero() {
		return nil, ecode.ErrReq
	}
	if relations, e := d.RedisGetUserRelations(ctx, userid); e == nil {
		return relations, nil
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserRelations] redis op failed", log.String("user_id", userid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid.Hex(), func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserRelations] db op failed", log.String("user_id", userid.Hex()), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(ctx, userid, relations); e != nil {
				log.Error(ctx, "[dao.GetUserRelations] update redis failed", log.String("user_id", userid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return nil, e
	}
	return *(*[]*model.RelationTarget)(unsafeRelations), nil
}
func (d *Dao) CountUserRelations(ctx context.Context, userid primitive.ObjectID, exceptTarget primitive.ObjectID, exceptTargetType string) (uint64, error) {
	if userid.IsZero() || (!exceptTarget.IsZero() && exceptTargetType != "user" && exceptTargetType != "group") {
		return 0, ecode.ErrReq
	}
	if count, e := d.RedisCountUserRelations(ctx, userid, exceptTarget, exceptTargetType); e == nil || e == ecode.ErrUserNotExist {
		return count, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.CountUserRelations] redis op failed", log.String("user_id", userid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid.Hex(), func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.CountUserRelations] db op failed", log.String("user_id", userid.Hex()), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(ctx, userid, relations); e != nil {
				log.Error(ctx, "[dao.CountUserRelations] update redis failed", log.String("user_id", userid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&relations), nil
	})
	if e != nil {
		return 0, e
	}
	relations := *(*[]*model.RelationTarget)(unsafeRelations)
	if !exceptTarget.IsZero() {
		for _, relation := range relations {
			if relation.Target == exceptTarget && relation.TargetType == exceptTargetType {
				return uint64(len(relations)) - 2, nil
			}
		}
	}
	return uint64(len(relations)) - 1, nil
}
func (d *Dao) GetUserRelation(ctx context.Context, userid, target primitive.ObjectID, targetType string) (*model.RelationTarget, error) {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") {
		return nil, ecode.ErrReq
	}
	if relation, e := d.RedisGetUserRelation(ctx, userid, target, targetType); e == nil || e == ecode.ErrUserNotExist || e == ecode.ErrNotFriends || e == ecode.ErrNotInGroup {
		return relation, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetUserRelation] redis op failed",
			log.String("user_id", userid.Hex()),
			log.String("target", target.Hex()),
			log.String("target_type", targetType),
			log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeRelations, e := oneshot.Do("GetUserRelations_"+userid.Hex(), func() (unsafe.Pointer, error) {
		relations, e := d.MongoGetUserRelations(ctx, userid)
		if e != nil {
			log.Error(ctx, "[dao.GetUserRelation] db op failed", log.String("user_id", userid.Hex()), log.CError(e))
			if e != ecode.ErrUserNotExist {
				return nil, e
			}
			//if the error is ErrUserNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetUserRelations(ctx, userid, relations); e != nil {
				log.Error(ctx, "[dao.GetUserRelation] update redis failed", log.String("user_id", userid.Hex()), log.CError(e))
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
func (d *Dao) UpdateDutyInGroup(ctx context.Context, userid, groupid primitive.ObjectID, newduty uint8) error {
	now, e := d.MongoUpdateDutyInGroup(ctx, userid, groupid, newduty)
	if e != nil {
		log.Error(ctx, "[dao.UpdateDutyInGroup] db op failed",
			log.String("user_id", userid.Hex()),
			log.String("group_id", groupid.Hex()),
			log.Uint64("duty", uint64(newduty)),
			log.CError(e))
		return e
	}
	if e = d.RedisRefreshGroupMember(ctx, groupid, userid, now.Name, now.Duty); e != nil {
		log.Error(ctx, "[dao.UpdateDutyInGroup] redis op failed",
			log.String("user_id", userid.Hex()),
			log.String("group_id", groupid.Hex()),
			log.Uint64("duty", uint64(newduty)),
			log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) AcceptGroupApply(ctx context.Context, groupid, userid primitive.ObjectID) error {
	username, groupname, e := d.MongoAcceptGroupApply(ctx, groupid, userid)
	if e != nil {
		log.Error(ctx, "[dao.AcceptGroupApply] db op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	// user's relation should be deleted first,because the group's admin will retry
	if e := d.RedisAddUserRelation(ctx, userid, groupid, "group", groupname); e != nil && e != gredis.Nil {
		if e == ecode.ErrUserNotExist {
			e = d.RedisDelUserRelations(ctx, userid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupApply] redis op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
			return e
		}
	}
	if e := d.RedisAddGroupMember(ctx, groupid, userid, username, 0); e != nil && e != gredis.Nil {
		if e == ecode.ErrGroupNotExist {
			e = d.RedisDelGroupMembers(ctx, groupid)
		}
		if e != nil {
			log.Error(ctx, "[dao.AcceptGroupApply] redis op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
			return e
		}
	}
	return nil
}
func (d *Dao) KickGroup(ctx context.Context, groupid, userid primitive.ObjectID) error {
	if e := d.MongoKickGroup(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[dao.KickGroup] db op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	// user's relation should be deleted first,because the group's admin will retry
	if e := d.RedisDelUserRelation(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[dao.KickGroup] redis op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	if e := d.RedisDelGroupMember(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[dao.KickGroup] redis op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) SetGroupName(ctx context.Context, groupid primitive.ObjectID, name string) error {
	if e := d.MongoSetGroupName(ctx, groupid, name); e != nil {
		log.Error(ctx, "[dao.SetGroupName] db op failed", log.String("group_id", groupid.Hex()), log.String("name", name), log.CError(e))
		return e
	}
	if e := d.RedisDelGroupName(ctx, groupid); e != nil {
		log.Error(ctx, "[dao.SetGroupName] redis op failed", log.String("group_id", groupid.Hex()), log.String("name", name), log.CError(e))
		return e
	}
	return nil
}
func (d *Dao) GetGroupName(ctx context.Context, groupid primitive.ObjectID) (string, error) {
	if groupid.IsZero() {
		return "", ecode.ErrReq
	}
	if groupname, e := d.RedisGetGroupName(ctx, groupid); e == nil || e == ecode.ErrGroupNotExist {
		return groupname, nil
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupName] redis op failed", log.String("group_id", groupid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeGroupName, e := oneshot.Do("GetGroupName_"+groupid.Hex(), func() (unsafe.Pointer, error) {
		groupname, e := d.MongoGetGroupName(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupName] db op failed", log.String("group_id", groupid.Hex()), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupName(context.Background(), groupid, groupname); e != nil {
				log.Error(ctx, "[dao.GetGroupName] update redis failed", log.String("group_id", groupid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&groupname), e
	})
	if e != nil {
		return "", e
	}
	return *(*string)(unsafeGroupName), e
}
func (d *Dao) GetGroupMembers(ctx context.Context, groupid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	if members, e := d.RedisGetGroupMembers(ctx, groupid); e == nil {
		return members, nil
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupMembers] redis op failed", log.String("group_id", groupid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid.Hex(), func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupMembers] db op failed", log.String("group_id", groupid.Hex()), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(ctx, groupid, members); e != nil {
				log.Error(ctx, "[dao.GetGroupMembers] update redis failed", log.String("group_id", groupid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return nil, e
	}
	return *(*[]*model.RelationTarget)(unsafeMembers), nil
}
func (d *Dao) CountGroupMembers(ctx context.Context, groupid primitive.ObjectID, exceptMember primitive.ObjectID) (uint64, error) {
	if groupid.IsZero() {
		return 0, ecode.ErrReq
	}
	if count, e := d.RedisCountGroupMembers(ctx, groupid, exceptMember); e == nil || e == ecode.ErrGroupNotExist {
		return count, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.CountGroupMembers] redis op failed", log.String("group_id", groupid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid.Hex(), func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.CountGroupMembers] db op failed", log.String("group_id", groupid.Hex()), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(ctx, groupid, members); e != nil {
				log.Error(ctx, "[dao.CountGroupMembers] update redis failed", log.String("group_id", groupid.Hex()), log.CError(e))
			}
		}()
		return unsafe.Pointer(&members), nil
	})
	if e != nil {
		return 0, e
	}
	members := *(*[]*model.RelationTarget)(unsafeMembers)
	if !exceptMember.IsZero() {
		for _, member := range members {
			if member.Target == exceptMember {
				return uint64(len(members)) - 2, nil
			}
		}
	}
	return uint64(len(members)) - 1, nil
}
func (d *Dao) GetGroupMember(ctx context.Context, groupid, userid primitive.ObjectID) (*model.RelationTarget, error) {
	if groupid.IsZero() || userid.IsZero() {
		return nil, ecode.ErrReq
	}
	if member, e := d.RedisGetGroupMember(ctx, groupid, userid); e == nil || e == ecode.ErrGroupNotExist || e == ecode.ErrNotInGroup {
		return member, e
	} else if e != nil && e != gredis.Nil {
		log.Error(ctx, "[dao.GetGroupMember] redis op failed", log.String("group_id", groupid.Hex()), log.String("user_id", userid.Hex()), log.CError(e))
	}
	//redis error or redis not exist,we need to query db
	unsafeMembers, e := oneshot.Do("GetGroupMembers_"+groupid.Hex(), func() (unsafe.Pointer, error) {
		members, e := d.MongoGetGroupMembers(ctx, groupid)
		if e != nil {
			log.Error(ctx, "[dao.GetGroupMember] db op failed", log.String("group_id", groupid.Hex()), log.CError(e))
			if e != ecode.ErrGroupNotExist {
				return nil, e
			}
			//if the error is ErrGroupNotExist,set the empty value in redis below
		}
		go func() {
			if e := d.RedisSetGroupMembers(ctx, groupid, members); e != nil {
				log.Error(ctx, "[dao.GetGroupMember] update redis failed", log.String("group_id", groupid.Hex()), log.CError(e))
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
