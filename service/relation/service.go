package relation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	chatdao "github.com/chenjie199234/im/dao/chat"
	relationdao "github.com/chenjie199234/im/dao/relation"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"
	"github.com/chenjie199234/im/util"

	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/log/trace"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/util/common"
	"github.com/chenjie199234/Corelib/util/graceful"
)

// Service subservice for relation business
type Service struct {
	stop *graceful.Graceful

	relationDao *relationdao.Dao
	chatDao     *chatdao.Dao
}

// Start -
func Start() *Service {
	return &Service{
		stop: graceful.New(),

		relationDao: relationdao.NewDao(nil, config.GetRedis("im_redis"), config.GetMongo("im_mongo")),
		chatDao:     chatdao.NewDao(nil, config.GetRedis("im_redis"), config.GetMongo("im_mongo")),
	}
}
func (s *Service) UpdateSelfName(ctx context.Context, req *api.UpdateSelfNameReq) (*api.UpdateSelfNameResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	//rate check
	if e := s.relationDao.RedisUpdateUserRelationRate(ctx, userid); e != nil {
		log.Error(ctx, "[UpdateSelfName] rate check failed", log.String("user_id", userid), log.String("new_name", req.NewName), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.relationDao.SetUserName(ctx, userid, req.NewName); e != nil {
		log.Error(ctx, "[UpdateSelfName] dao op failed",
			log.String("user_id", userid),
			log.String("new_name", req.NewName),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.UpdateSelfNameResp{}, nil
}
func (s *Service) UpdateGroupName(ctx context.Context, req *api.UpdateGroupNameReq) (*api.UpdateGroupNameResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	if info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator); e != nil {
		log.Error(ctx, "[UpdateGroupName] get operator's group info failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty != 2 {
		return nil, ecode.ErrPermission
	}
	//rate check
	if e := s.relationDao.RedisUpdateGroupNameRate(ctx, req.GroupId); e != nil {
		log.Error(ctx, "[UpdateGroupName] rate check failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.String("new_name", req.NewName),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.relationDao.SetGroupName(ctx, req.GroupId, req.NewName); e != nil {
		log.Error(ctx, "[UpdateGroupName] dao op failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.String("new_name", req.NewName),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.UpdateGroupNameResp{}, nil
}
func (s *Service) MakeFriend(ctx context.Context, req *api.MakeFriendReq) (*api.MakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	requester := md["Token-User"]
	if requester == req.UserId {
		return nil, ecode.ErrReq
	}
	//check accepter already set self's name
	if _, e := s.relationDao.GetUserName(ctx, req.UserId); e != nil {
		log.Error(ctx, "[MakeFriend] check accepter's name failed", log.String("accepter", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check requester already set self's name
	requestername, e := s.relationDao.GetUserName(ctx, requester)
	if e != nil {
		log.Error(ctx, "[MakeFriend] check requester's name failed", log.String("requester", requester), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in accepter's view
	if _, e := s.relationDao.GetUserRelation(ctx, req.UserId, requester, "user"); e != nil && e != ecode.ErrNotFriends {
		log.Error(ctx, "[MakeFriend] check current relation in accepter's view failed",
			log.String("requester", requester),
			log.String("accepter", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in requester's view
		if _, e = s.relationDao.GetUserRelation(ctx, requester, req.UserId, "user"); e != nil && e != ecode.ErrNotFriends {
			log.Error(ctx, "[MakeFriend] check current relation in accepter's view failed",
				log.String("requester", requester),
				log.String("accepter", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			//already be friends,this is unnecessary
			return &api.MakeFriendResp{}, nil
		} else if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check accepte's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, req.UserId, requester, "user"); e != nil {
				log.Error(ctx, "[MakeFriend] check accepter's current relations count failed", log.String("accepter", req.UserId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrTargetTooManyRelations
			}
		}
	} else if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check requester's currnent relations count
		if count, e := s.relationDao.CountUserRelations(ctx, requester, req.UserId, "user"); e != nil {
			log.Error(ctx, "[MakeFriend] check requester's current relations count failed", log.String("requester", requester), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
		//check accepter's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, req.UserId, requester, "user"); e != nil {
			log.Error(ctx, "[MakeFriend] check accepter's current relations count failed", log.String("accepter", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}
	if e := s.relationDao.RedisAddMakeFriendRequest(ctx, requester, requestername, req.UserId); e != nil {
		log.Error(ctx, "[MakeFriend] redis op failed", log.String("requester", requester), log.String("accepter", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.MakeFriendResp{}, nil
}
func (s *Service) AcceptMakeFriend(ctx context.Context, req *api.AcceptMakeFriendReq) (*api.AcceptMakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	accepter := md["Token-User"]
	if accepter == req.UserId {
		return nil, ecode.ErrReq
	}
	if e := s.relationDao.RedisRefreshUserRequest(ctx, accepter, req.UserId, "user"); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] redis op failed", log.String("accepter", accepter), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check accepter's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, accepter, req.UserId, "user"); e != nil {
			log.Error(ctx, "[AcceptMakeFriend] check accepter's current relations count failed", log.String("accepter", accepter), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
		//check friend's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, req.UserId, accepter, "user"); e != nil {
			log.Error(ctx, "[AcceptMakeFriend] check friend's current relations count failed", log.String("friend_id", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if username, friendname, e := s.relationDao.AcceptMakeFriend(ctx, accepter, req.UserId); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] dao op failed", log.String("accepter", accepter), log.String("friend_id", req.UserId), log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else {
		//push to self
		util.AddWork(func() {
			s.pushuser(trace.CloneSpan(ctx), accepter, "userRelationAdd", req.UserId, "user", friendname)
			s.stop.DoneOne()
		})
		//push to friend
		util.AddWork(func() {
			s.pushuser(trace.CloneSpan(ctx), req.UserId, "userRelationAdd", accepter, "user", username)
			s.stop.DoneOne()
		})
	}
	if e := s.relationDao.RedisDelUserRequest(ctx, accepter, req.UserId, "user"); e != nil {
		log.Error(ctx, "[AcceptMakeFriend] redis op failed", log.String("accepter", accepter), log.String("friend_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.AcceptMakeFriendResp{}, nil
}
func (s *Service) RefuseMakeFriend(ctx context.Context, req *api.RefuseMakeFriendReq) (*api.RefuseMakeFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser := md["Token-User"]
	if refuser == req.UserId {
		return nil, ecode.ErrReq
	}
	if e := s.relationDao.RedisDelUserRequest(ctx, refuser, req.UserId, "user"); e != nil {
		log.Error(ctx, "[RefuseMakeFriend] redis op failed", log.String("refuser", refuser), log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ErrReq
	}
	return &api.RefuseMakeFriendResp{}, nil
}
func (s *Service) GroupInvite(ctx context.Context, req *api.GroupInviteReq) (*api.GroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	inviter := md["Token-User"]
	if inviter == req.UserId {
		return nil, ecode.ErrReq
	}
	//check inviter permission
	if info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, inviter); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[GroupInvite] get inviter's group info failed", log.String("inviter", inviter), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	//check user already set self's name
	if _, e := s.relationDao.GetUserName(ctx, req.UserId); e != nil {
		log.Error(ctx, "[GroupInvite] check user's name failed", log.String("user_id", req.UserId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in user's view
	if _, e := s.relationDao.GetUserRelation(ctx, req.UserId, req.GroupId, "group"); e != nil && e != ecode.ErrNotInGroup {
		log.Error(ctx, "[GroupInvite] check current relation in user's view failed",
			log.String("user_id", req.UserId),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in group's view
		if _, e = s.relationDao.GetGroupMember(ctx, req.GroupId, req.UserId); e != nil && e != ecode.ErrGroupMemberNotExist {
			log.Error(ctx, "[GroupInvite] check current relation in group's view failed",
				log.String("group_id", req.GroupId),
				log.String("user_id", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			return &api.GroupInviteResp{}, nil
		} else if max := config.AC.Service.MaxGroupMember; max != 0 {
			//check group's current members count
			if count, e := s.relationDao.CountGroupMembers(ctx, req.GroupId, req.UserId); e != nil {
				log.Error(ctx, "[GroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
	} else {
		if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check user's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, req.UserId, req.GroupId, "group"); e != nil {
				log.Error(ctx, "[GroupInvite] check user's current relations count failed", log.String("user_id", req.UserId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrTargetTooManyRelations
			}
		}
		if max := config.AC.Service.MaxGroupMember; max != 0 {
			//check group's current relations count
			if count, e := s.relationDao.CountGroupMembers(ctx, req.GroupId, req.UserId); e != nil {
				log.Error(ctx, "[GroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
	}
	//check group name
	groupname, e := s.relationDao.GetGroupName(ctx, req.GroupId)
	if e != nil {
		log.Error(ctx, "[GroupInvite] get group's name failed", log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.relationDao.RedisAddGroupInviteRequest(ctx, req.GroupId, groupname, req.UserId); e != nil {
		log.Error(ctx, "[GroupInvite] redis op failed",
			log.String("inviter", inviter),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GroupInviteResp{}, nil
}
func (s *Service) AcceptGroupInvite(ctx context.Context, req *api.AcceptGroupInviteReq) (*api.AcceptGroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	accepter := md["Token-User"]
	if e := s.relationDao.RedisRefreshUserRequest(ctx, accepter, req.GroupId, "group"); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] redis op failed", log.String("accepter", accepter), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check accepter's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, accepter, req.GroupId, "group"); e != nil {
			log.Error(ctx, "[AcceptGroupInvite] check accepter's current relations count failed", log.String("accepter", accepter), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrSelfTooManyRelations
		}
	}
	if max := config.AC.Service.MaxGroupMember; max != 0 {
		//check group's current members count
		if count, e := s.relationDao.CountGroupMembers(ctx, req.GroupId, accepter); e != nil {
			log.Error(ctx, "[AcceptGroupInvite] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrGroupTooManyMembers
		}
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if username, groupname, e := s.relationDao.AcceptGroupInvite(ctx, accepter, req.GroupId); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] dao op failed", log.String("accepter", accepter), log.String("group_id", req.GroupId), log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else {
		//push to self
		util.AddWork(func() {
			s.pushuser(trace.CloneSpan(ctx), accepter, "userRelationAdd", req.GroupId, "group", groupname)
			s.stop.DoneOne()
		})
		//push to group members
		go func() {
			s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupJoin", accepter, username, 0)
			s.stop.DoneOne()
		}()
	}
	if e := s.relationDao.RedisDelUserRequest(ctx, accepter, req.GroupId, "group"); e != nil {
		log.Error(ctx, "[AcceptGroupInvite] redis op failed", log.String("accepter", accepter), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.AcceptGroupInviteResp{}, nil
}
func (s *Service) RefuseGroupInvite(ctx context.Context, req *api.RefuseGroupInviteReq) (*api.RefuseGroupInviteResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser := md["Token-User"]
	if e := s.relationDao.RedisDelUserRequest(ctx, refuser, req.GroupId, "group"); e != nil {
		log.Error(ctx, "[RefuseGroupInvite] redis op failed",
			log.String("refuser", refuser),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ErrReq
	}
	return &api.RefuseGroupInviteResp{}, nil
}
func (s *Service) GroupApply(ctx context.Context, req *api.GroupApplyReq) (*api.GroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	requester := md["Token-User"]
	//check requester already set self's name
	requestername, e := s.relationDao.GetUserName(ctx, requester)
	if e != nil {
		log.Error(ctx, "[GroupApply] check requester's name failed", log.String("requester", requester), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in group's view
	if _, e := s.relationDao.GetGroupMember(ctx, req.GroupId, requester); e != nil && e != ecode.ErrGroupMemberNotExist {
		log.Error(ctx, "[GroupApply] check current relation in group's view failed",
			log.String("group_id", req.GroupId),
			log.String("requester", requester),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == nil {
		//check current relation in requester's view
		if _, e = s.relationDao.GetUserRelation(ctx, requester, req.GroupId, "group"); e != nil && e != ecode.ErrNotInGroup {
			log.Error(ctx, "[GroupApply] check current relation in requester's view failed",
				log.String("requester", requester),
				log.String("group_id", req.GroupId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == nil {
			return &api.GroupApplyResp{}, nil
		} else if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check requester's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, requester, req.GroupId, "group"); e != nil {
				log.Error(ctx, "[GroupApply] check requester's current relations count failed",
					log.String("requester", requester),
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
			if count, e := s.relationDao.CountGroupMembers(ctx, req.GroupId, requester); e != nil {
				log.Error(ctx, "[GroupApply] check group's current relations count failed", log.String("group_id", req.GroupId), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrGroupTooManyMembers
			}
		}
		if max := config.AC.Service.MaxUserRelation; max != 0 {
			//check requester's current relations count
			if count, e := s.relationDao.CountUserRelations(ctx, requester, req.GroupId, "group"); e != nil {
				log.Error(ctx, "[GroupApply] check requester's current relations count failed", log.String("requester", requester), log.CError(e))
				return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
			} else if count >= uint64(max) {
				return nil, ecode.ErrSelfTooManyRelations
			}
		}
	}
	if e := s.relationDao.RedisAddGroupApplyRequest(ctx, requester, requestername, req.GroupId); e != nil {
		log.Error(ctx, "[GroupApply] redis op failed",
			log.String("requester", requester),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GroupApplyResp{}, nil
}
func (s *Service) AcceptGroupApply(ctx context.Context, req *api.AcceptGroupApplyReq) (*api.AcceptGroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	accepter := md["Token-User"]
	if accepter == req.UserId {
		return nil, ecode.ErrReq
	}
	//check permission
	if info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, accepter); e != nil {
		log.Error(ctx, "[AcceptGroupApply] get accepter's group info failed",
			log.String("accepter", accepter),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	if max := config.AC.Service.MaxGroupMember; max != 0 {
		//check group's current members count
		if count, e := s.relationDao.CountGroupMembers(ctx, req.GroupId, req.UserId); e != nil {
			log.Error(ctx, "[AcceptGroupApply] check group's current members count failed", log.String("group_id", req.GroupId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrGroupTooManyMembers
		}
	}
	if max := config.AC.Service.MaxUserRelation; max != 0 {
		//check requester's current relations count
		if count, e := s.relationDao.CountUserRelations(ctx, req.UserId, req.GroupId, "group"); e != nil {
			log.Error(ctx, "[AcceptGroupApply] check requester's current relations count failed", log.String("requester", req.UserId), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if count >= uint64(max) {
			return nil, ecode.ErrTargetTooManyRelations
		}
	}
	if e := s.relationDao.RedisRefreshGroupRequest(ctx, req.GroupId, req.UserId); e != nil {
		log.Error(ctx, "[AcceptGroupApply] redis op failed",
			log.String("accepter", accepter),
			log.String("group_id", req.GroupId),
			log.String("requester", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if username, groupname, e := s.relationDao.AcceptGroupApply(ctx, req.GroupId, req.UserId); e != nil {
		log.Error(ctx, "[AcceptGroupApply] dao op failed",
			log.String("accepter", accepter),
			log.String("group_id", req.GroupId),
			log.String("requester", req.UserId),
			log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else {
		//push to requester
		util.AddWork(func() {
			s.pushuser(trace.CloneSpan(ctx), req.UserId, "userRelationAdd", req.GroupId, "group", groupname)
			s.stop.DoneOne()
		})
		//push to group members
		go func() {
			s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupJoin", req.UserId, username, 0)
			s.stop.DoneOne()
		}()
	}
	if e := s.relationDao.RedisDelGroupRequest(ctx, req.GroupId, req.UserId); e != nil {
		log.Error(ctx, "[AcceptGroupApply] redis op failed",
			log.String("accepter", accepter),
			log.String("group_id", req.GroupId),
			log.String("requester", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.AcceptGroupApplyResp{}, nil
}
func (s *Service) RefuseGroupApply(ctx context.Context, req *api.RefuseGroupApplyReq) (*api.RefuseGroupApplyResp, error) {
	md := metadata.GetMetadata(ctx)
	refuser := md["Token-User"]
	//check refuser permission
	if info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, refuser); e != nil {
		log.Error(ctx, "[RefuseGroupApply] get refuser's group info failed",
			log.String("refuser", refuser),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	if e := s.relationDao.RedisDelGroupRequest(ctx, req.GroupId, req.UserId); e != nil {
		log.Error(ctx, "[RefuseGroupApply] redis op failed",
			log.String("refuser", refuser),
			log.String("group_id", req.GroupId),
			log.String("requester", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.RefuseGroupApplyResp{}, nil
}
func (s *Service) DelFriend(ctx context.Context, req *api.DelFriendReq) (*api.DelFriendResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	if operator == req.UserId {
		return nil, ecode.ErrReq
	}
	//check current relation in friend's view
	if _, e := s.relationDao.GetUserRelation(ctx, req.UserId, operator, "user"); e != nil && e != ecode.ErrNotFriends {
		log.Error(ctx, "[DelFriend] check current relation in friend's view failed",
			log.String("operator", operator),
			log.String("friend_id", req.UserId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrNotFriends {
		//check current relation in operator's view
		if _, e := s.relationDao.GetUserRelation(ctx, operator, req.UserId, "user"); e != nil && e != ecode.ErrNotFriends {
			log.Error(ctx, "[DelFriend] check current relation in operator's view failed",
				log.String("operator", operator),
				log.String("friend_id", req.UserId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == ecode.ErrNotFriends {
			return &api.DelFriendResp{}, nil
		}
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if e := s.relationDao.DelFriend(ctx, operator, req.UserId); e != nil {
		log.Error(ctx, "[DelFriend] dao op failed", log.String("operator", operator), log.String("friend_id", req.UserId), log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push to self
	util.AddWork(func() {
		s.pushuser(trace.CloneSpan(ctx), operator, "userRelationDel", req.UserId, "user", "")
		s.stop.DoneOne()
	})
	//push to target
	util.AddWork(func() {
		s.pushuser(trace.CloneSpan(ctx), req.UserId, "userRelationDel", operator, "user", "")
		s.stop.DoneOne()
	})
	return &api.DelFriendResp{}, nil
}
func (s *Service) LeaveGroup(ctx context.Context, req *api.LeaveGroupReq) (*api.LeaveGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	//check current relation in group's view
	if _, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator); e != nil && e != ecode.ErrGroupMemberNotExist {
		log.Error(ctx, "[LeaveGroup] check current relation in group's view failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrGroupMemberNotExist {
		//check current relation in operator's view
		if _, e = s.relationDao.GetUserRelation(ctx, operator, req.GroupId, "group"); e != nil && e != ecode.ErrNotInGroup {
			log.Error(ctx, "[LeaveGroup] check current relation in operator's view failed",
				log.String("operator", operator),
				log.String("group_id", req.GroupId),
				log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		} else if e == ecode.ErrNotInGroup {
			return &api.LeaveGroupResp{}, nil
		}
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if e := s.relationDao.LeaveGroup(ctx, operator, req.GroupId); e != nil {
		log.Error(ctx, "[LeaveGroup] dao op failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push to self
	util.AddWork(func() {
		s.pushuser(trace.CloneSpan(ctx), operator, "userRelationDel", req.GroupId, "group", "")
		s.stop.DoneOne()
	})
	//push to group members
	go func() {
		s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupLeave", operator, "", 0)
		s.stop.DoneOne()
	}()
	return &api.LeaveGroupResp{}, nil
}
func (s *Service) KickGroup(ctx context.Context, req *api.KickGroupReq) (*api.KickGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	if operator == req.UserId {
		return nil, ecode.ErrReq
	}
	var userinfo *model.RelationTarget
	//check current relation in user's view
	if _, e := s.relationDao.GetUserRelation(ctx, req.UserId, req.GroupId, "group"); e != nil && e != ecode.ErrNotInGroup {
		log.Error(ctx, "[KickGroup] check current relation in user's view failed",
			log.String("user_id", req.UserId),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if e == ecode.ErrNotInGroup {
		//check current relation in group's view
		if userinfo, e = s.relationDao.GetGroupMember(ctx, req.GroupId, req.UserId); e != nil && e != ecode.ErrGroupMemberNotExist {
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
	if info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[KickGroup] get operator's group info failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	} else if info.Duty <= userinfo.Duty {
		return nil, ecode.ErrPermission
	}
	if e := s.stop.Add(2); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	if e := s.relationDao.KickGroup(ctx, req.GroupId, req.UserId); e != nil {
		log.Error(ctx, "[KickGroup] dao op failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.CError(e))
		s.stop.DoneOne()
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push to be kicked member
	util.AddWork(func() {
		s.pushuser(trace.CloneSpan(ctx), req.UserId, "userRelationDel", req.GroupId, "group", "")
		s.stop.DoneOne()
	})
	//push to group members
	go func() {
		s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupKick", req.UserId, "", 0)
		s.stop.DoneOne()
	}()
	return &api.KickGroupResp{}, nil
}
func (s *Service) Relations(ctx context.Context, req *api.RelationsReq) (*api.RelationsResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	//check rate
	if e := s.relationDao.RedisGetUserRelationsRate(ctx, userid); e != nil {
		log.Error(ctx, "[Relations] rate check failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	relations, e := s.relationDao.GetUserRelations(ctx, userid)
	if e != nil {
		log.Error(ctx, "[Relations] dao op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	indexes, e := s.chatDao.RedisGetIndexAll(ctx, userid)
	if e != nil {
		log.Error(ctx, "[Relations] dao op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	resp := &api.RelationsResp{
		Update:    true,
		Relations: make([]*api.RelationInfo, 0, len(relations)),
	}
	missed := make([]string, 0, 10)
	for _, relation := range relations {
		if relation.Target == "" {
			continue
		}
		chatkey := util.FormChatKey(userid, relation.Target, relation.TargetType)
		if _, ok := indexes[chatkey]; !ok {
			missed = append(missed, chatkey)
		}
	}
	if len(missed) > 0 {
		lker := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		wg.Add(len(missed))
		for _, v := range missed {
			chatkey := v
			util.AddWork(func() {
				defer wg.Done()
				index, err := s.chatDao.GetIndex(ctx, userid, chatkey)
				if err != nil {
					log.Error(ctx, "[Relations] dao op failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(err))
					e = err
					return
				}
				lker.Lock()
				indexes[chatkey] = &model.IMIndex{
					MsgIndex:    index.MsgIndex,
					RecallIndex: index.RecallIndex,
					AckIndex:    index.AckIndex,
				}
				lker.Unlock()
			})
		}
		wg.Wait()
		if e != nil {
			return nil, e
		}
	}
	for _, relation := range relations {
		if relation.Target == "" {
			//self
			resp.Relations = append(resp.Relations, &api.RelationInfo{
				Target:      relation.Target,
				TargetType:  relation.TargetType,
				Name:        relation.Name,
				MsgIndex:    0,
				RecallIndex: 0,
				AckIndex:    0,
			})
		} else {
			chatkey := util.FormChatKey(userid, relation.Target, relation.TargetType)
			index := indexes[chatkey]
			resp.Relations = append(resp.Relations, &api.RelationInfo{
				Target:      relation.Target,
				TargetType:  relation.TargetType,
				Name:        relation.Name,
				MsgIndex:    index.MsgIndex,
				RecallIndex: index.RecallIndex,
				AckIndex:    index.AckIndex,
			})
		}
	}
	strs := make([]string, 0, len(relations))
	for _, v := range resp.Relations {
		str := v.TargetType +
			"_" +
			v.Target +
			"_" +
			v.Name +
			"_" +
			strconv.FormatUint(uint64(v.MsgIndex), 10) +
			"_" +
			strconv.FormatUint(uint64(v.RecallIndex), 10) +
			"_" +
			strconv.FormatUint(uint64(v.AckIndex), 10)
		strs = append(strs, str)
	}
	sort.Strings(strs)
	hashstr := sha256.Sum256(common.STB(strings.Join(strs, ",")))
	if hex.EncodeToString(hashstr[:]) == req.CurrentHash {
		resp.Update = false
		resp.Relations = nil
	}
	//do clean
	go func() {
		needdel := make([]string, 0, 10)
		for chatkey := range indexes {
			find := false
			for _, relation := range relations {
				if util.FormChatKey(userid, relation.Target, relation.TargetType) == chatkey {
					find = true
					break
				}
			}
			if !find {
				needdel = append(needdel, chatkey)
			}
		}
		if len(needdel) == 0 {
			return
		}
		ctx := trace.CloneSpan(ctx)
		for _, v := range needdel {
			chatkey := v
			util.AddWork(func() {
				if e := s.chatDao.RedisDelIndex(ctx, userid, chatkey); e != nil {
					log.Error(ctx, "[Relations] del redis index failed", log.String("user_id", userid), log.String("chat_key", chatkey), log.CError(e))
				}
			})
		}
	}()
	return resp, nil
}
func (s *Service) GroupMembers(ctx context.Context, req *api.GroupMembersReq) (*api.GroupMembersResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	//check rate
	if e := s.relationDao.RedisGetGroupMembersRate(ctx, userid, req.GroupId); e != nil {
		log.Error(ctx, "[GroupMembers] rate check failed", log.String("user_id", userid), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in user's view
	if _, e := s.relationDao.GetUserRelation(ctx, userid, req.GroupId, "group"); e != nil {
		log.Error(ctx, "[GroupMembers] check current relation in user's view failed",
			log.String("user_id", userid),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//check current relation in group's view
	if _, e := s.relationDao.GetGroupMember(ctx, req.GroupId, userid); e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[GroupMembers] check current relation in group's view failed",
			log.String("group_id", req.GroupId),
			log.String("user_id", userid),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	members, e := s.relationDao.GetGroupMembers(ctx, req.GroupId)
	if e != nil {
		log.Error(ctx, "[GroupMembers] dao op failed",
			log.String("user_id", userid),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, e
	}
	resp := &api.GroupMembersResp{
		Update:  true,
		Members: make([]*api.GroupMemberInfo, 0, len(members)),
	}
	strs := make([]string, 0, len(members))
	for _, member := range members {
		resp.Members = append(resp.Members, &api.GroupMemberInfo{
			Member: member.Target,
			Name:   member.Name,
			Duty:   uint32(member.Duty),
		})
		strs = append(strs, member.Target+"_"+strconv.Itoa(int(member.Duty))+"_"+member.Name)
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
	userid := md["Token-User"]
	if req.TargetType == "user" && req.Target == userid {
		return nil, ecode.ErrReq
	}
	//rate check
	if e := s.relationDao.RedisUpdateUserRelationRate(ctx, userid); e != nil {
		log.Error(ctx, "[UpdateUserRelationName] rate check failed",
			log.String("user_id", userid),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.String("new_name", req.NewName),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if e := s.relationDao.UpdateUserRelationName(ctx, userid, req.Target, req.TargetType, req.NewName); e != nil {
		log.Error(ctx, "[UpdateUserRelationName] dao op failed",
			log.String("user_id", userid),
			log.String("target", req.Target),
			log.String("target_type", req.TargetType),
			log.String("new_name", req.NewName),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.UpdateUserRelationNameResp{}, nil
}
func (s *Service) UpdateNameInGroup(ctx context.Context, req *api.UpdateNameInGroupReq) (*api.UpdateNameInGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	if e := s.stop.Add(1); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	//rate check
	if e := s.relationDao.RedisUpdateUserNameInGroupRate(ctx, userid, req.GroupId); e != nil {
		log.Error(ctx, "[UpdateNameInGroup] rate check failed",
			log.String("user_id", userid),
			log.String("group_id", req.GroupId),
			log.String("new_name", req.NewName),
			log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	now, e := s.relationDao.UpdateNameInGroup(ctx, userid, req.GroupId, req.NewName)
	if e != nil {
		log.Error(ctx, "[UpdateNameInGroup] dao op failed",
			log.String("user_id", userid),
			log.String("group_id", req.GroupId),
			log.String("new_name", req.NewName),
			log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push to group members
	go func() {
		s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupJoin", userid, now.Name, now.Duty)
		s.stop.DoneOne()
	}()
	return nil, nil
}
func (s *Service) UpdateDutyInGroup(ctx context.Context, req *api.UpdateDutyInGroupReq) (*api.UpdateDutyInGroupResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	//check user permission
	userinfo, e := s.relationDao.GetGroupMember(ctx, req.GroupId, req.UserId)
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
	operatorinfo, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator)
	if e != nil {
		if e == ecode.ErrGroupMemberNotExist {
			e = ecode.ErrNotInGroup
		}
		log.Error(ctx, "[UpdateDutyInGroup] get operator's group info failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if operatorinfo.Duty == 0 || operatorinfo.Duty <= uint8(req.NewDuty) || operatorinfo.Duty <= userinfo.Duty {
		return nil, ecode.ErrPermission
	}
	if e := s.stop.Add(1); e != nil {
		if e == graceful.ErrClosing {
			return nil, ecode.ErrServerClosing
		}
		return nil, ecode.ErrBusy
	}
	//check rate
	if e := s.relationDao.RedisUpdateUserDutyInGroupRate(ctx, req.GroupId); e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] rate check failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.Uint64("new_duty", uint64(req.NewDuty)),
			log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	now, e := s.relationDao.UpdateDutyInGroup(ctx, req.UserId, req.GroupId, uint8(req.NewDuty))
	if e != nil {
		log.Error(ctx, "[UpdateDutyInGroup] dao op failed",
			log.String("operator", operator),
			log.String("group_id", req.GroupId),
			log.String("user_id", req.UserId),
			log.Uint64("new_duty", uint64(req.NewDuty)),
			log.CError(e))
		s.stop.DoneOne()
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	//push to group members
	go func() {
		s.pushgroup(trace.CloneSpan(ctx), req.GroupId, "groupJoin", req.UserId, now.Name, now.Duty)
		s.stop.DoneOne()
	}()
	return nil, nil
}
func (s *Service) GetSelfRequestsCount(ctx context.Context, req *api.GetSelfRequestsCountReq) (*api.GetSelfRequestsCountResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	count, e := s.relationDao.RedisCountUserRequests(ctx, userid)
	if e != nil {
		log.Error(ctx, "[GetSelfRequestsCount] redis op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GetSelfRequestsCountResp{Count: uint32(count)}, nil
}
func (s *Service) GetSelfRequests(ctx context.Context, req *api.GetSelfRequestsReq) (*api.GetSelfRequestsResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	requests, e := s.relationDao.RedisGetUserRequests(ctx, userid, req.Cursor, req.Direction, 10)
	if e != nil {
		log.Error(ctx, "[GetSelfRequests] redis op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GetSelfRequestsResp{Requesters: requests}, nil
}
func (s *Service) GetGroupRequestsCount(ctx context.Context, req *api.GetGroupRequestsCountReq) (*api.GetGroupRequestsCountResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	//check operator permission
	info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator)
	if e != nil {
		log.Error(ctx, "[GetGroupRequestsCount] get operator's group info failed", log.String("operator", operator), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	count, e := s.relationDao.RedisCountGroupRequests(ctx, req.GroupId)
	if e != nil {
		log.Error(ctx, "[GetGroupRequestsCount] redis op failed", log.String("operator", operator), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GetGroupRequestsCountResp{Count: uint32(count)}, nil
}
func (s *Service) GetGroupRequests(ctx context.Context, req *api.GetGroupRequestsReq) (*api.GetGroupRequestsResp, error) {
	md := metadata.GetMetadata(ctx)
	operator := md["Token-User"]
	//check operator permission
	info, e := s.relationDao.GetGroupMember(ctx, req.GroupId, operator)
	if e != nil {
		log.Error(ctx, "[GetGroupRequests] get operator's group info failed", log.String("operator", operator), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if info.Duty == 0 {
		return nil, ecode.ErrPermission
	}
	requests, e := s.relationDao.RedisGetGroupRequests(ctx, req.GroupId, req.Cursor, req.Direction, 10)
	if e != nil {
		log.Error(ctx, "[GetGroupRequests] redis op failed", log.String("operator", operator), log.String("group_id", req.GroupId), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.GetGroupRequestsResp{Requesters: requests}, nil
}

// Stop -
func (s *Service) Stop() {
	s.stop.Close(nil, nil)
}
