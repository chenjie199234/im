package match

import (
	"context"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/config"
	matchdao "github.com/chenjie199234/im/dao/match"
	relationdao "github.com/chenjie199234/im/dao/relation"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	// "github.com/chenjie199234/Corelib/cgrpc"
	// "github.com/chenjie199234/Corelib/crpc"
	// "github.com/chenjie199234/Corelib/log"
	// "github.com/chenjie199234/Corelib/web"
	"github.com/chenjie199234/Corelib/log"
	"github.com/chenjie199234/Corelib/metadata"
	"github.com/chenjie199234/Corelib/util/graceful"
)

// Service subservice for match business
type Service struct {
	stop *graceful.Graceful

	matchDao    *matchdao.Dao
	relationDao *relationdao.Dao
}

// Start -
func Start() *Service {
	s := &Service{
		stop: graceful.New(),

		matchDao:    matchdao.NewDao(nil, config.GetRedis("match_redis"), nil),
		relationDao: relationdao.NewDao(nil, config.GetRedis("im_redis"), config.GetMongo("im_mongo")),
	}
	s.job()
	return s
}
func (s *Service) Status(ctx context.Context, req *api.StatusReq) (*api.StatusResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	status, e := s.matchDao.RedisGetUserMatchStatus(ctx, userid)
	if e != nil {
		log.Error(ctx, "[Status] redis op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	if status == nil {
		return &api.StatusResp{}, nil
	}
	return status, nil
}
func (s *Service) Do(ctx context.Context, req *api.DoReq) (*api.DoResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	if config.AC.Service.MaxUserRelation != 0 {
		count, e := s.relationDao.CountUserRelations(ctx, userid, "", "")
		if e != nil {
			log.Error(ctx, "[Do] check user's current relations count failed", log.String("user_id", userid), log.CError(e))
			return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
		}
		if count >= uint64(config.AC.Service.MaxUserRelation)+1 {
			return nil, ecode.ErrSelfTooManyRelations
		}
	}

	var activities map[string][2]uint8
	var cities map[string][]string
	switch req.Lan {
	case "zh":
		activities = model.ActivitiesCN
		cities = model.CitiesCN
	case "en":
		fallthrough
	case "ru":
		fallthrough
	case "de":
		fallthrough
	case "fr":
		fallthrough
	case "ar":
		fallthrough
	case "hi":
		fallthrough
	case "it":
		fallthrough
	case "ja":
		fallthrough
	case "ko":
		fallthrough
	default:
		return nil, ecode.ErrBan
	}
	if allregions, ok := cities[req.City]; !ok {
		log.Error(ctx, "[Do] unknown city", log.String("user_id", userid), log.String("city", req.City))
		return nil, ecode.ErrReq
	} else {
		for _, region := range req.Regions {
			find := false
			for _, v := range allregions {
				if v == region {
					find = true
					break
				}
			}
			if !find {
				log.Error(ctx, "[Do] unknown region", log.String("user_id", userid), log.String("region", region))
				return nil, ecode.ErrReq
			}
		}
	}
	for _, activity := range req.Activities {
		if _, ok := activities[activity]; !ok {
			log.Error(ctx, "[Do] unknown activity", log.String("user_id", userid), log.String("activity", activity))
			return nil, ecode.ErrReq
		}
	}
	if e := s.matchDao.RedisStartMatch(ctx, userid, req.Lan, req.City, req.Regions, req.Activities); e != nil {
		log.Error(ctx, "[Do] redis op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.DoResp{}, nil
}
func (s *Service) Cancel(ctx context.Context, req *api.CancelReq) (*api.CancelResp, error) {
	md := metadata.GetMetadata(ctx)
	userid := md["Token-User"]
	if e := s.matchDao.RedisCancelMatch(ctx, userid); e != nil {
		log.Error(ctx, "[Cancel] redis op failed", log.String("user_id", userid), log.CError(e))
		return nil, ecode.ReturnEcode(e, ecode.ErrSystem)
	}
	return &api.CancelResp{}, nil
}
func (s *Service) Activities(ctx context.Context, req *api.ActivitiesReq) (*api.ActivitiesResp, error) {
	var activities map[string][2]uint8
	switch req.Lan {
	case "zh":
		activities = model.ActivitiesCN
	case "en":
		fallthrough
	case "ru":
		fallthrough
	case "de":
		fallthrough
	case "fr":
		fallthrough
	case "ar":
		fallthrough
	case "hi":
		fallthrough
	case "it":
		fallthrough
	case "ja":
		fallthrough
	case "ko":
		fallthrough
	default:
		return nil, ecode.ErrBan
	}
	resp := &api.ActivitiesResp{
		Infos: make([]*api.ActivityInfo, 0, len(model.ActivitiesCN)),
	}
	for activity, peoplenum := range activities {
		resp.Infos = append(resp.Infos, &api.ActivityInfo{
			Name: activity,
			Min:  uint32(peoplenum[0]),
			Max:  uint32(peoplenum[1]),
		})
	}
	return resp, nil
}
func (s *Service) Cities(ctx context.Context, req *api.CitiesReq) (*api.CitiesResp, error) {
	var cities map[string][]string
	switch req.Lan {
	case "zh":
		cities = model.CitiesCN
	case "en":
		fallthrough
	case "ru":
		fallthrough
	case "de":
		fallthrough
	case "fr":
		fallthrough
	case "ar":
		fallthrough
	case "hi":
		fallthrough
	case "it":
		fallthrough
	case "ja":
		fallthrough
	case "ko":
		fallthrough
	default:
		return nil, ecode.ErrBan
	}
	resp := &api.CitiesResp{
		Infos: make([]*api.CityInfo, 0, len(model.CitiesCN)),
	}
	for city, regions := range cities {
		resp.Infos = append(resp.Infos, &api.CityInfo{
			Name:    city,
			Regions: regions,
		})
	}
	return resp, nil
}

// Stop -
func (s *Service) Stop() {
	s.stop.Close(nil, nil)
}
