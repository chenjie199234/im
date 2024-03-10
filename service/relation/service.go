package relation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	relationdao "github.com/chenjie199234/im/dao/relation"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/util/common"
	"github.com/chenjie199234/Corelib/util/graceful"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service subservice for relation business
type Service struct {
	stop *graceful.Graceful

	relationDao *relationdao.Dao
}

// Start -
func Start() *Service {
	return &Service{
		stop: graceful.New(),

		relationDao: relationdao.NewDao(config.GetMysql("relation_mysql"), config.GetRedis("relation_redis"), config.GetMongo("relation_mongo")),
	}
}
func (s *Service) MakeFriend(ctx context.Context, req *api.MakeFriendReq) (*api.MakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	requester, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[MakeFriend] token format wrong", log.String("requester", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	accepter, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[MakeFriend] accepter format wrong", log.String("accepter", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if requester.IsZero() || accepter.IsZero() || requester == accepter {
		return nil, ecode.ErrReq
	}

	//check accepter already set self's name
	if _, e = s.relationDao.GetUserName(ctx, accepter); e != nil {
		log.Error(ctx, "[MakeFriend] check accepter's name failed", log.String("accepter", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check requester already set self's name
	if _, e = s.relationDao.GetUserName(ctx, requester); e != nil {
		log.Error(ctx, "[MakeFriend] check requester's name failed", log.String("requester", md["Token-User"]), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in accepter's view
	if _, e = s.relationDao.GetUserRelation(ctx, accepter, requester, "user"); e != nil && e != ecode.ErrNotFriends {
		log.Error(ctx, "[MakeFriend] check current relation in accepter's view failed",
			log.String("requester", md["Token-User"]),
			log.String("accepter", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in requester's view
		if _, e = s.relationDao.GetUserRelation(ctx, requester, accepter, "user"); e != nil && e != ecode.ErrNotFriends {
			log.Error(ctx, "[MakeFriend] check current relation in accepter's view failed",
				log.String("requester", md["Token-User"]),
				log.String("accepter", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			//already be friends,this is unnecessary
			return &api.MakeFriendResp{}, nil
		} else if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check accepte's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, accepter, requester, "user"); e != nil {
				log.Error(ctx, "[MakeFriend] check accepter's current relations count failed", log.String("accepter", req.UserId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrTargetTooManyRelations
			}
		}
	} else if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check requester's currnent relations count
		if count, e := s.relationDao.CountUserRelations(ctx, requester, accepter, "user"); e != nil {
			log.Error(ctx, "[MakeFriend] check requester's current relations count failed", log.String("requester", md["Token-User"]), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
		//check accepter's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, accepter, requester, "user"); e != nil {
			log.Error(ctx, "[MakeFriend] check accepter's current relations count failed", log.String("accepter", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}

	if e := s.relationDao.RedisAddMakeFriendRequest(ctx, requester, accepter); e != nil {
		log.Error(ctx, "[MakeFriend] redis op failed", log.String("requester", md["Token-User"]), log.String("accepter", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.MakeFriendResp{}, nil
}
func (s *Service) AcceptMakeFriend(ctx context.Context, req *api.AcceptMakeFriendReq) (*api.AcceptMakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[AcceptMakeFriend] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	friendid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[AcceptMakeFriend] friendid format wrong", log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || friendid.IsZero() || userid == friendid {
		return nil, ecode.ErrReq
	}
	if e = s.relationDao.RedisRefreshUserRequest(ctx, userid, friendid, "user"); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] redis op failed", log.String("user_id", md["Token-User"]), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check user's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, userid, friendid, "user"); e != nil {
			log.Error(ctx, "[AcceptMakeFriend] check user's current relations count failed", log.String("user_id", md["Token-User"]), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
		//check friend's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, friendid, userid, "user"); e != nil {
			log.Error(ctx, "[AcceptMakeFriend] check friend's current relations count failed", log.String("friend_id", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}
	if e = s.relationDao.AcceptMakeFriend(ctx, userid, friendid); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] dao op failed", log.String("user_id", md["Token-User"]), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e = s.relationDao.RedisDelUserRequest(ctx, userid, friendid, "user"); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] redis op failed", log.String("user_id", md["Token-User"]), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.AcceptMakeFriendResp{}, nil
}
func (s *Service) RefuseMakeFriend(ctx context.Context, req *api.RefuseMakeFriendReq) (*api.RefuseMakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[RefuseMakeFriend] token format wrong", log.String("refuser", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	userid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[RefuseMakeFriend] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if refuser.IsZero() || userid.IsZero() || refuser == userid {
		return nil, ecode.ErrReq
	}
	if e = s.relationDao.RedisDelUserRequest(ctx, refuser, userid, "user"); e != nil {
		log.Error(ctx, "[RefuseMakeFriend] redis op failed", log.String("refuser", md["Token-User"]), log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	return &api.RefuseMakeFriendResp{}, nil
}
func (s *Service) GroupInvite(ctx context.Context, req *api.GroupInviteReq) (*api.GroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	inviter, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[GroupInvite] token format wrong", log.String("inviter", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	userid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[GroupInvite] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[GroupInvite] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if inviter.IsZero() || userid.IsZero() || groupid.IsZero() || inviter == userid {
		return nil, ecode.ErrReq
	}
	//check user already set self's name
	if _, e = s.relationDao.GetUserName(ctx, userid); e != nil {
		log.Error(ctx, "[GroupInvite] check user's name failed", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check inviter permission
	if info, e := s.relationDao.GetGroupMember(ctx, groupid, inviter); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[GroupInvite] get inviter's group info failed", log.String("inviter", md["Token-User"]), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	//check current relation in user's view
	if _, e = s.relationDao.GetUserRelation(ctx, userid, groupid, "group"); e != nil && e != ecode.ErrNotInGroup {
		log.Error(ctx, "[GroupInvite] check current relation in user's view failed",
			log.String("user_id", req.UserId),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in group's view
		if _, e = s.relationDao.GetGroupMember(ctx, groupid, userid); e != nil && e != ecode.ErrGroupMemberNotExist {
			log.Error(ctx, "[GroupInvite] check current relation in group's view failed",
				log.String("group_id", req.GroupId),
				log.String("user_id", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			return &api.GroupInviteResp{}, nil
		} else if max := config.AC.Service.MaxGroupMember; max != 0 {
			//check group's current members count
			if count, e := s.relationDao.CountGroupMembers(ctx, groupid, userid); e != nil {
				log.Error(ctx, "[GroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
	} else {
		if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check user's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, userid, groupid, "group"); e != nil {
				log.Error(ctx, "[GroupInvite] check user's current relations count failed", log.String("user_id", req.UserId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrTargetTooManyRelations
			}
		}
		if max := config.AC.Service.MaxGroupMember; max != 0 {
			//check group's current relations count
			if count, e := s.relationDao.CountGroupMembers(ctx, groupid, userid); e != nil {
				log.Error(ctx, "[GroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
	}
	if e = s.relationDao.RedisAddGroupInviteRequest(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[GroupInvite] redis op failed",
			log.String("inviter", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GroupInviteResp{}, nil
}
func (s *Service) AcceptGroupInvite(ctx context.Context, req *api.AcceptGroupInviteReq) (*api.AcceptGroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[AcceptGroupInvite] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[AcceptGroupInvite] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	if e = s.relationDao.RedisRefreshUserRequest(ctx, userid, groupid, "user"); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] redis op failed", log.String("user_id", md["Token-User"]), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check user's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, userid, groupid, "group"); e != nil {
			log.Error(ctx, "[AcceptGroupInvite] check user's current relations count failed", log.String("user_id", md["Token-User"]), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
	}
	if max := config.AC.Service.MaxGroupMember; max != 0 {
		//check group's current members count
		if count, e := s.relationDao.CountGroupMembers(ctx, groupid, userid); e != nil {
			log.Error(ctx, "[AcceptGroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrGroupTooManyMembers
		}
	}
	if e = s.relationDao.AcceptGroupInvite(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] dao op failed", log.String("user_id", md["Token-User"]), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e = s.relationDao.RedisDelUserRequest(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] redis op failed", log.String("user_id", md["Token-User"]), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.AcceptGroupInviteResp{}, nil
}
func (s *Service) RefuseGroupInvite(ctx context.Context, req *api.RefuseGroupInviteReq) (*api.RefuseGroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[RefuseGroupInvite] token format wrong", log.String("refuser", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[RefuseGroupInvite] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if refuser.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	if e = s.relationDao.RedisDelUserRequest(ctx, refuser, groupid, "group"); e != nil {
		log.Error(ctx, "[RefuseGroupInvite] redis op failed",
			log.String("refuser", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ErrReq
	}
	return &api.RefuseGroupInviteResp{}, nil
}
func (s *Service) GroupApply(ctx context.Context, req *api.GroupApplyReq) (*api.GroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[GroupApply] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[GroupApply] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	//check user already set self's name
	if _, e := s.relationDao.GetUserName(ctx, userid); e != nil {
		log.Error(ctx, "[GroupApply] check user's name failed", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in group's view
	if _, e = s.relationDao.GetGroupMember(ctx, groupid, userid); e != nil && e != ecode.ErrGroupMemberNotExist {
		log.Error(ctx, "[GroupApply] check current relation in group's view failed",
			log.String("group_id", req.GroupId),
			log.String("user_id", md["Token-User"]),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in user's view
		if _, e = s.relationDao.GetUserRelation(ctx, userid, groupid, "group"); e != nil && e != ecode.ErrNotInGroup {
			log.Error(ctx, "[GroupApply] check current relation in user's view failed",
				log.String("user_id", md["Token-User"]),
				log.String("group_id", req.GroupId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			return &api.GroupApplyResp{}, nil
		} else if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check user's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, userid, groupid, "group"); e != nil {
				log.Error(ctx, "[GroupApply] check user's current relations count failed",
					log.String("user_id", md["Token-User"]),
					log.String("group_id", req.GroupId),
					log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrSelfTooManyRelations
			}
		}
	} else {
		if max := config.AC.Service.MaxGroupMember; max != 0 {
			//check group's current members count
			if count, e := s.relationDao.CountGroupMembers(ctx, groupid, userid); e != nil {
				log.Error(ctx, "[GroupApply] check group's current relations count failed", log.String("group_id", groupid.Hex()), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
		if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check user's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, userid, groupid, "group"); e != nil {
				log.Error(ctx, "[GroupApply] check user's current relations count failed", log.String("user_id", md["Token-User"]), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrSelfTooManyRelations
			}
		}
	}
	if e = s.relationDao.RedisAddGroupApplyRequest(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[GroupApply] redis op failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GroupApplyResp{}, nil
}
func (s *Service) AcceptGroupApply(ctx context.Context, req *api.AcceptGroupApplyReq) (*api.AcceptGroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	operator, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[AcceptGroupApply] token format wrong", log.String("operator", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[AcceptGroupApply] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	userid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[AcceptGroupApply] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if operator.IsZero() || groupid.IsZero() || userid.IsZero() || operator == userid {
		return nil, ecode.ErrReq
	}
	//check permission
	if info, e := s.relationDao.GetGroupMember(ctx, groupid, operator); e != nil {
		log.Error(ctx, "[AcceptGroupApply] get operator's group info failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	if max := config.AC.Service.MaxGroupMember; max != 0 {
		//check group's current members count
		if count, e := s.relationDao.CountGroupMembers(ctx, groupid, userid); e != nil {
			log.Error(ctx, "[AcceptGroupApply] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrGroupTooManyMembers
		}
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check user's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, userid, groupid, "group"); e != nil {
			log.Error(ctx, "[AcceptGroupApply] check user's current relations count failed", log.String("user_id", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}
	if e = s.relationDao.RedisRefreshGroupRequest(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[AcceptGroupApply] redis op failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e = s.relationDao.AcceptGroupApply(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[AcceptGroupApply] dao op failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e = s.relationDao.RedisDelGroupRequest(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[AcceptGroupApply] redis op failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return nil, nil
}
func (s *Service) RefuseGroupApply(ctx context.Context, req *api.RefuseGroupApplyReq) (*api.RefuseGroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[RefuseGroupApply] token format wrong", log.String("refuser", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[RefuseGroupApply] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	userid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[RefuseGroupApply] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if refuser.IsZero() || groupid.IsZero() || userid.IsZero() || refuser == userid {
		return nil, ecode.ErrReq
	}
	//check refuser permission
	if info, e := s.relationDao.GetGroupMember(ctx, groupid, refuser); e != nil {
		log.Error(ctx, "[RefuseGroupApply] get refuser's group info failed",
			log.String("refuser", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	if e := s.relationDao.RedisDelGroupRequest(ctx, groupid, userid); e != nil {
		log.Error(ctx, "[RefuseGroupApply] redis op failed",
			log.String("refuser", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.RefuseGroupApplyResp{}, nil
}
func (s *Service) DelFriend(ctx context.Context, req *api.DelFriendReq) (*api.DelFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[DelFriend] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	friendid, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[DelFriend] friendid format wrong", log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || friendid.IsZero() || userid == friendid {
		return nil, ecode.ErrReq
	}
	//check current relation in friend's view
	if _, e = s.relationDao.GetUserRelation(ctx, friendid, userid, "user"); e != nil && e != ecode.ErrNotFriends {
		log.Error(ctx, "[DelFriend] check current relation in friend's view failed",
			log.String("user_id", md["Token-User"]),
			log.String("friend_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrNotFriends {
		//check current relation in user's view
		if _, e = s.relationDao.GetUserRelation(ctx, userid, friendid, "user"); e != nil && e != ecode.ErrNotFriends {
			log.Error(ctx, "[DelFriend] check current relation in user's view failed",
				log.String("user_id", md["Token-User"]),
				log.String("friend_id", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == ecode.ErrNotFriends {
			return &api.DelFriendResp{}, nil
		}
	}
	if e = s.relationDao.DelFriend(ctx, userid, friendid); e != nil {
		log.Error(ctx, "[DelFriend] dao op failed", log.String("user_id", md["Token-User"]), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.DelFriendResp{}, nil
}
func (s *Service) LeaveGroup(ctx context.Context, req *api.LeaveGroupReq) (*api.LeaveGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[LeaveGroup] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[LeaveGroup] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	//check current relation in group's view
	if _, e := s.relationDao.GetGroupMember(ctx, groupid, userid); e != nil && e != ecode.ErrGroupMemberNotExist {
		log.Error(ctx, "[LeaveGroup] check current relation in group's view failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrGroupMemberNotExist {
		//check current relation in user's view
		if _, e = s.relationDao.GetUserRelation(ctx, userid, groupid, "group"); e != nil && e != ecode.ErrNotInGroup {
			log.Error(ctx, "[LeaveGroup] check current relation in user's view failed",
				log.String("user_id", md["Token-User"]),
				log.String("group_id", req.GroupId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == ecode.ErrNotInGroup {
			return &api.LeaveGroupResp{}, nil
		}
	}
	if e = s.relationDao.LeaveGroup(ctx, userid, groupid); e != nil {
		log.Error(ctx, "[LeaveGroup] dao op failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.LeaveGroupResp{}, nil
}
func (s *Service) KickGroup(ctx context.Context, req *api.KickGroupReq) (*api.KickGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	operator, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[KickGroup] token format wrong", log.String("operator", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[KickGroup] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	member, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[KickGroup] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if operator.IsZero() || groupid.IsZero() || member.IsZero() || operator == member {
		return nil, ecode.ErrReq
	}
	var userinfo *model.RelationTarget
	//check current relation in user's view
	if _, e = s.relationDao.GetUserRelation(ctx, member, groupid, "group"); e != nil && e != ecode.ErrNotInGroup {
		log.Error(ctx, "[KickGroup] check current relation in user's view failed",
			log.String("user_id", req.UserId),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrNotInGroup {
		//check current relation in group's view
		if userinfo, e = s.relationDao.GetGroupMember(ctx, groupid, member); e != nil && e != ecode.ErrGroupMemberNotExist {
			log.Error(ctx, "[KickGroup] check current relation in group's view failed",
				log.String("group_id", req.GroupId),
				log.String("user_id", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == ecode.ErrGroupMemberNotExist {
			return &api.KickGroupResp{}, nil
		}
	}
	//check operator permission
	if info, e := s.relationDao.GetGroupMember(ctx, groupid, operator); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[KickGroup] get operator's group info failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty <= userinfo.Duty {
		return nil, ecode.ErrPermission
	}
	if e = s.relationDao.KickGroup(ctx, groupid, member); e != nil {
		log.Error(ctx, "[KickGroup] dao op failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.KickGroupResp{}, nil
}
func (s *Service) Relations(ctx context.Context, req *api.RelationsReq) (*api.RelationsResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[Relations] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	if userid.IsZero() {
		return nil, ecode.ErrReq
	}
	relations, e := s.relationDao.GetUserRelations(ctx, userid)
	if e != nil {
		log.Error(ctx, "[Relations] dao op failed", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	resp := &api.RelationsResp{
		Update:    true,
		Relations: make([]*api.RelationInfo, 0, len(relations)),
	}
	strs := make([]string, 0, len(relations))
	for _, relation := range relations {
		resp.Relations = append(resp.Relations, &api.RelationInfo{
			Target:     relation.Target.Hex(),
			TargetType: relation.TargetType,
			Name:       relation.Name,
			Duty:       uint32(relation.Duty),
		})
		strs = append(strs, relation.TargetType+"_"+relation.Target.Hex()+"_"+relation.Name)
	}
	sort.Strings(strs)
	hashstr := sha256.Sum256(common.STB(strings.Join(strs, ",")))
	if hex.EncodeToString(hashstr[:]) == req.CurrentHash {
		resp.Update = false
		resp.Relations = nil
	}
	return resp, nil
}
func (s *Service) GroupMembers(ctx context.Context, req *api.GroupMembersReq) (*api.GroupMembersResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[GroupMembers] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[GroupMembers] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	//check current relation in user's view
	if _, e = s.relationDao.GetUserRelation(ctx, userid, groupid, "group"); e != nil {
		log.Error(ctx, "[GroupMembers] check current relation in user's view failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in group's view
	if _, e = s.relationDao.GetGroupMember(ctx, groupid, userid); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[GroupMembers] check current relation in group's view failed",
			log.String("group_id", req.GroupId),
			log.String("user_id", md["Token-User"]),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	members, e := s.relationDao.GetGroupMembers(ctx, groupid)
	if e != nil {
		log.Error(ctx, "[GroupMembers] dao op failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, e
	}
	resp := &api.GroupMembersResp{
		Update:  true,
		Members: make([]*api.RelationInfo, 0, len(members)),
	}
	strs := make([]string, 0, len(members))
	for _, member := range members {
		resp.Members = append(resp.Members, &api.RelationInfo{
			Target:     member.Target.Hex(),
			TargetType: "user",
			Name:       member.Name,
			Duty:       uint32(member.Duty),
		})
		strs = append(strs, member.Target.Hex()+"_"+strconv.Itoa(int(member.Duty))+"_"+member.Name)
	}
	sort.Strings(strs)
	hashstr := sha256.Sum256(common.STB(strings.Join(strs, ",")))
	if hex.EncodeToString(hashstr[:]) == req.CurrentHash {
		resp.Update = false
		resp.Members = nil
	}
	return resp, nil
}
func (s *Service) UpdateUserRelationName(ctx context.Context, req *api.UpdateUserRelationNameReq) (*api.UpdateUserRelationNameResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[UpdateUserRelationName] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	target, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[UpdateUserRelationName] target format wrong", log.String("target", req.Target), log.String("target_type", req.TargetType), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || target.IsZero() {
		return nil, ecode.ErrReq
	}
	//check current relation and current name in user's view
	info, e := s.relationDao.GetUserRelation(ctx, userid, target, req.TargetType)
	if e != nil {
		log.Error(ctx, "[UpdateUserRelationName] get user's relation info failed",
			log.String("user_id", md["Token-User"]),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if info.Name == req.NewName {
		return &api.UpdateUserRelationNameResp{}, nil
	}
	if e = s.relationDao.UpdateUserRelationName(ctx, userid, target, req.TargetType, req.NewName); e != nil {
		log.Error(ctx, "[UpdateUserRelationName] dao op failed",
			log.String("user_id", md["Token-User"]),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.UpdateUserRelationNameResp{}, nil
}
func (s *Service) UpdateNameInGroup(ctx context.Context, req *api.UpdateNameInGroupReq) (*api.UpdateNameInGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	userid, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[UpdateNameInGroup] token format wrong", log.String("user_id", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[UpdateNameInGroup] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if userid.IsZero() || groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	//check current relation and current name in group's view
	info, e := s.relationDao.GetGroupMember(ctx, groupid, userid)
	if e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[UpdateNameInGroup] get user's group info failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if info.Name == req.NewName {
		return &api.UpdateNameInGroupResp{}, nil
	}
	if e = s.relationDao.UpdateNameInGroup(ctx, userid, groupid, req.NewName); e != nil {
		log.Error(ctx, "[UpdateNameInGroup] dao op failed",
			log.String("user_id", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return nil, nil
}
func (s *Service) UpdateDutyInGroup(ctx context.Context, req *api.UpdateDutyInGroupReq) (*api.UpdateDutyInGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	operator, e := primitive.ObjectIDFromHex(md["Token-User"])
	if e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] token format wrong", log.String("operator", md["Token-User"]), log.CError(e))
		return nil, ecode.ErrToken
	}
	groupid, e := primitive.ObjectIDFromHex(req.GroupId)
	if e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] groupid format wrong", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ErrReq
	}
	member, e := primitive.ObjectIDFromHex(req.UserId)
	if e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] userid format wrong", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	if operator.IsZero() || groupid.IsZero() || member.IsZero() || operator == member {
		return nil, ecode.ErrReq
	}
	//check user permission
	userinfo, e := s.relationDao.GetGroupMember(ctx, groupid, member)
	if e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] get user's group info failed",
			log.String("user_id", req.UserId),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if userinfo.Duty == uint8(req.NewDuty) {
		return &api.UpdateDutyInGroupResp{}, nil
	}
	//check operator permission
	operatorinfo, e := s.relationDao.GetGroupMember(ctx, groupid, operator)
	if e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[UpdateDutyInGroup] get operator's group info failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if operatorinfo.Duty == 0 || operatorinfo.Duty <= uint8(req.NewDuty) || operatorinfo.Duty <= userinfo.Duty {
		return nil, ecode.ErrPermission
	}
	if e = s.relationDao.UpdateDutyInGroup(ctx, member, groupid, uint8(req.NewDuty)); e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] dao op failed",
			log.String("operator", md["Token-User"]),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.Uint64("new_duty", uint64(req.NewDuty)),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return nil, nil
}

// Stop -
func (s *Service) Stop() {
	s.stop.Close(nil, nil)
}
