package relation

import (
	"context"
	"strconv"
	"strings"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	"github.com/chenjie199234/Corelib/util/common"
	gredis "github.com/redis/go-redis/v9"
)

const defaultExpire = 86400
const maxExpire = 604800

// in redis's SortedSet,member num <=128,use listpack to store
const defaultRequest = 32
const maxRequest = 128

var addRequestScript *gredis.Script
var refreshRequestScript *gredis.Script
var countRequestScript *gredis.Script
var getRequestScript *gredis.Script
var delRequestScript *gredis.Script

var setRelationScript *gredis.Script
var addRelationScript *gredis.Script

func init() {
	addRequestScript = gredis.NewScript(`local time=redis.call("TIME")
local dels=redis.call("ZRANGE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[3]*1000000,"BYSCORE")
if(dels and #dels>0) then
	redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[3]*1000000)
	redis.call("HDEL",KEYS[2],unpack(dels))
end
redis.call("ZREM",KEYS[1],ARGV[1])
if(redis.call("ZCARD",KEYS[1])>=tonumber(ARGV[4])) then
	return -1
end
local num=redis.call("ZADD",KEYS[1],time[1]*1000000+time[2],ARGV[1])
redis.call("HSET",KEYS[2],ARGV[1],ARGV[2])
redis.call("EXPIRE",KEYS[1],ARGV[3])
redis.call("EXPIRE",KEYS[2],ARGV[3])
return num`)
	refreshRequestScript = gredis.NewScript(`local time=redis.call("TIME")
if(redis.call("EXPIRE",KEYS[1],ARGV[2])==0) then
	return nil
end
redis.call("EXPIRE",KEYS[2],ARGV[2])
local dels=redis.call("ZRANGE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[2]*1000000,"BYSCORE")
if(dels and #dels>0) then
	redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[2]*1000000)
	redis.call("HDEL",KEYS[2],unpack(dels))
end
if(not redis.call("ZSCORE",KEYS[1],ARGV[1])) then
	return nil
end
local num=redis.call("ZADD",KEYS[1],time[1]*1000000+time[2],ARGV[1])
return num`)
	countRequestScript = gredis.NewScript(`local time=redis.call("TIME")
if(redis.call("EXPIRE",KEYS[1],ARGV[1])==0) then
	return 0
end
redis.call("EXPIRE",KEYS[2],ARGV[1])
local dels=redis.call("ZRANGE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[1]*1000000,"BYSCORE")
if(dels and #dels>0) then
	redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[1]*1000000)
	redis.call("HDEL",KEYS[2],unpack(dels))
end
return redis.call("ZCARD",KEYS[1])`)
	getRequestScript = gredis.NewScript(`local time=redis.call("TIME")
if(redis.call("EXPIRE",KEYS[1],ARGV[1])==0) then
	return {}
end
redis.call("EXPIRE",KEYS[2],ARGV[1])
local dels=redis.call("ZRANGE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[1]*1000000,"BYSCORE")
if(dels and #dels>0) then
	redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[1]*1000000)
	redis.call("HDEL",KEYS[2],unpack(dels))
end
local tmp
if(ARGV[4]=="REV") then
	tmp=redis.call("ZRANGE",KEYS[1],ARGV[2],ARGV[3],"BYSCORE","REV","LIMIT",0,ARGV[5],"WITHSCORES")
else
	tmp=redis.call("ZRANGE",KEYS[1],ARGV[2],ARGV[3],"BYSCORE","LIMIT",0,ARGV[4],"WITHSCORES")
end
if(not tmp or #tmp==0) then
	return {}
end
local fields={}
local result={}
for i=1,#tmp,2 do
	table.insert(fields,tmp[i])
	table.insert(result,tmp[i])
	table.insert(result,tmp[i+1])
	table.insert(result,"")
end
tmp=redis.call("HMGET",KEYS[2],unpack(fields))
for i=1,#tmp,1 do
	result[i*3]=tmp[i]
end
return result`)
	delRequestScript = gredis.NewScript(`local time=redis.call("TIME")
if(redis.call("EXPIRE",KEYS[1],ARGV[2])==0) then
	return 0
end
redis.call("EXPIRE",KEYS[2],ARGV[2])
local dels=redis.call("ZRANGE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[2]*1000000,"BYSCORE")
if(dels and #dels>0) then
	redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]*1000000+time[2]-ARGV[2]*1000000)
	redis.call("HDEL",KEYS[2],unpack(dels))
end
local num=redis.call("ZREM",KEYS[1],ARGV[1])
if(num>0) then
	redis.call("HDEL",KEYS[2],ARGV[1])
end
return num`)

	setRelationScript = gredis.NewScript(`redis.call("DEL",KEYS[1])
local num=redis.call("HSET",KEYS[1],unpack(ARGV,2))
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
	addRelationScript = gredis.NewScript(`if(redis.call("EXISTS",KEYS[1])==0)
then
	return nil
end
local tmp=redis.call("HGET",KEYS[1],ARGV[2])
if(not tmp or tmp=="")
then
	return -1
end
local num=redis.call("HSET",KEYS[1],ARGV[3],ARGV[4])
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
}

//-----------------------user request--------------------------------------

func (d *Dao) RedisAddMakeFriendRequest(ctx context.Context, requester, requestername, accepter string) error {
	key1 := "request_user_{" + accepter + "}"
	key2 := key1 + "_extra"
	field := "user_" + requester
	num, e := addRequestScript.Run(ctx, d.imredis, []string{key1, key2}, field, requestername, defaultExpire, defaultRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}
func (d *Dao) RedisAddGroupInviteRequest(ctx context.Context, groupid, groupname, accepter string) error {
	key1 := "request_user_{" + accepter + "}"
	key2 := key1 + "_extra"
	field := "group_" + groupid
	num, e := addRequestScript.Run(ctx, d.imredis, []string{key1, key2}, field, groupname, defaultExpire, defaultRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}
func (d *Dao) RedisGetUserRequests(ctx context.Context, userid string, cursor uint64, direction string, count uint8) ([]*api.RequesterInfo, error) {
	key1 := "request_user_{" + userid + "}"
	key2 := key1 + "_extra"
	args := make([]interface{}, 0, 5)
	if direction == "after" {
		args = append(args, defaultExpire, cursor, "+inf", count)
	} else if direction == "before" {
		args = append(args, defaultExpire, cursor, "-inf", "REV", count)
	} else {
		return nil, ecode.ErrReq
	}
	rs, e := getRequestScript.Run(ctx, d.imredis, []string{key1, key2}, args...).Slice()
	if e != nil {
		return nil, e
	}
	result := make([]*api.RequesterInfo, 0, len(rs))
	for i := 0; i < len(rs); i += 3 {
		pieces := strings.Split(rs[i].(string), "_")
		if len(pieces) != 2 || (pieces[0] != "user" && pieces[0] != "group") {
			return nil, ecode.ErrCacheDataBroken
		}
		score, e := strconv.ParseUint(rs[i+1].(string), 10, 64)
		if e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
		if rs[i+2].(string) == "" {
			return nil, ecode.ErrCacheDataBroken
		}
		result = append(result, &api.RequesterInfo{
			Requester:     pieces[1],
			RequesterType: pieces[0],
			Cursor:        score,
			Name:          rs[i+2].(string),
		})
	}
	return result, nil
}
func (d *Dao) RedisCountUserRequests(ctx context.Context, userid string) (uint64, error) {
	key1 := "request_user_{" + userid + "}"
	key2 := key1 + "_extra"
	return countRequestScript.Run(ctx, d.imredis, []string{key1, key2}, defaultExpire).Uint64()
}
func (d *Dao) RedisRefreshUserRequest(ctx context.Context, userid, target, targetType string) error {
	key1 := "request_user_{" + userid + "}"
	key2 := key1 + "_extra"
	field := targetType + "_" + target
	e := refreshRequestScript.Run(ctx, d.imredis, []string{key1, key2}, field, defaultExpire).Err()
	if e != nil && e == gredis.Nil {
		e = ecode.ErrRequestNotExist
	}
	return e
}
func (d *Dao) RedisDelUserRequest(ctx context.Context, userid, target, targetType string) error {
	key1 := "request_user_{" + userid + "}"
	key2 := key1 + "_extra"
	field := targetType + "_" + target
	return delRequestScript.Run(ctx, d.imredis, []string{key1, key2}, field, defaultExpire).Err()
}

//-----------------------group request--------------------------------------

func (d *Dao) RedisAddGroupApplyRequest(ctx context.Context, requester, requestername, groupid string) error {
	key1 := "request_group_{" + groupid + "}"
	key2 := key1 + "_extra"
	num, e := addRequestScript.Run(ctx, d.imredis, []string{key1, key2}, requester, requestername, defaultExpire, maxRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}
func (d *Dao) RedisGetGroupRequests(ctx context.Context, groupid string, cursor uint64, direction string, count uint8) ([]*api.RequesterInfo, error) {
	key1 := "request_group_{" + groupid + "}"
	key2 := key1 + "_extra"
	args := make([]interface{}, 0, 5)
	if direction == "after" {
		args = append(args, defaultExpire, cursor, "+inf", count)
	} else if direction == "before" {
		args = append(args, defaultExpire, cursor, "-inf", "REV", count)
	} else {
		return nil, ecode.ErrReq
	}
	rs, e := getRequestScript.Run(ctx, d.imredis, []string{key1, key2}, args...).Slice()
	if e != nil {
		return nil, e
	}
	result := make([]*api.RequesterInfo, 0, len(rs))
	for i := 0; i < len(rs); i += 3 {
		score, e := strconv.ParseUint(rs[i+1].(string), 10, 64)
		if e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
		if rs[i+2].(string) == "" {
			return nil, ecode.ErrCacheDataBroken
		}
		result = append(result, &api.RequesterInfo{
			Requester:     rs[i].(string),
			RequesterType: "user",
			Cursor:        score,
			Name:          rs[i+2].(string),
		})
	}
	return result, nil
}
func (d *Dao) RedisCountGroupRequests(ctx context.Context, groupid string) (uint64, error) {
	key1 := "request_group_{" + groupid + "}"
	key2 := key1 + "_extra"
	return countRequestScript.Run(ctx, d.imredis, []string{key1, key2}, defaultExpire).Uint64()
}
func (d *Dao) RedisRefreshGroupRequest(ctx context.Context, groupid, userid string) error {
	key1 := "request_group_{" + groupid + "}"
	key2 := key1 + "_extra"
	e := refreshRequestScript.Run(ctx, d.imredis, []string{key1, key2}, userid, defaultExpire).Err()
	if e != nil && e == gredis.Nil {
		e = ecode.ErrRequestNotExist
	}
	return e
}
func (d *Dao) RedisDelGroupRequest(ctx context.Context, groupid, userid string) error {
	key1 := "request_group_{" + groupid + "}"
	key2 := key1 + "_extra"
	return delRequestScript.Run(ctx, d.imredis, []string{key1, key2}, userid, defaultExpire).Err()
}

//-----------------------user--------------------------------------

// targets must be empty or contain user self(target == "")
func (d *Dao) RedisSetUserRelations(ctx context.Context, userid string, targets []*model.RelationTarget) error {
	key := "relation_user_{" + userid + "}"
	args := make([]interface{}, 0, len(targets)*2+3)
	args = append(args, maxExpire, "user_", "")

	var selfname string
	for _, v := range targets {
		if (v.TargetType != "user" && v.TargetType != "group") || v.Name == "" {
			return ecode.ErrReq
		}
		if v.Target == "" {
			selfname = v.Name
		}
		args = append(args, v.TargetType+"_"+v.Target, v.Name)
	}
	if len(targets) > 0 && selfname == "" {
		return ecode.ErrReq
	}
	return setRelationScript.Run(ctx, d.imredis, []string{key}, args...).Err()
}

func (d *Dao) RedisCountUserRelations(ctx context.Context, userid, exceptTarget, exceptTargetType string) (uint64, error) {
	key := "relation_user_{" + userid + "}"
	count, e := d.imredis.HLen(ctx, key).Uint64()
	if e != nil {
		return 0, e
	}
	if count == 0 {
		return 0, gredis.Nil
	} else if count == 1 {
		username, e := d.imredis.HGet(ctx, key, "user_").Result()
		if e != nil {
			return 0, e
		}
		if username == "" {
			return 0, ecode.ErrUserNotExist
		}
	} else if exceptTarget != "" {
		exist, e := d.imredis.HExists(ctx, key, exceptTargetType+"_"+exceptTarget).Result()
		if e != nil {
			return 0, e
		}
		if exist {
			return count - 2, nil
		}
	}
	return count - 1, nil
}

func (d *Dao) RedisAddUserRelation(ctx context.Context, userid, target, targetType, targetname string) error {
	if (targetType != "user" && targetType != "group") || targetname == "" {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid + "}"
	args := []interface{}{maxExpire, "user_", targetType + "_" + target, targetname}
	r, e := addRelationScript.Run(ctx, d.imredis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrUserNotExist
	}
	return nil
}

func (d *Dao) RedisGetUserRelations(ctx context.Context, userid string) ([]*model.RelationTarget, error) {
	key := "relation_user_{" + userid + "}"
	all, e := d.imredis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, gredis.Nil
	}
	r := make([]*model.RelationTarget, 0, len(all))
	for k, name := range all {
		if strings.HasPrefix(k, "user_") {
			if k[5:] == "" && name == "" {
				return nil, ecode.ErrUserNotExist
			}
			if name == "" {
				return nil, ecode.ErrCacheDataBroken
			}
			r = append(r, &model.RelationTarget{
				Target:     k[5:],
				TargetType: "user",
				Name:       name,
			})
		} else if strings.HasPrefix(k, "group_") {
			if k[6:] == "" || name == "" {
				return nil, ecode.ErrCacheDataBroken
			}
			r = append(r, &model.RelationTarget{
				Target:     k[6:],
				TargetType: "group",
				Name:       name,
			})
		} else {
			return nil, ecode.ErrCacheDataBroken
		}
	}
	return r, nil
}

func (d *Dao) RedisGetUserRelation(ctx context.Context, userid, target, targetType string) (*model.RelationTarget, error) {
	key := "relation_user_{" + userid + "}"
	rs, e := d.imredis.HMGet(ctx, key, "user_", targetType+"_"+target).Result()
	if e != nil {
		return nil, e
	}
	if rs[0] == nil {
		return nil, gredis.Nil
	}
	if rs[0].(string) == "" {
		return nil, ecode.ErrUserNotExist
	}
	if rs[1] == nil {
		if targetType == "user" {
			return nil, ecode.ErrNotFriends
		} else if targetType == "group" {
			return nil, ecode.ErrNotInGroup
		} else {
			return nil, ecode.ErrReq
		}
	}
	name := rs[1].(string)
	if target != "" && name == "" {
		return nil, ecode.ErrCacheDataBroken
	}
	return &model.RelationTarget{
		Target:     target,
		TargetType: targetType,
		Name:       name,
	}, nil
}

func (d *Dao) RedisDelUserRelations(ctx context.Context, userid string) error {
	key := "relation_user_{" + userid + "}"
	return d.imredis.Del(ctx, key).Err()
}

func (d *Dao) RedisDelUserRelation(ctx context.Context, userid, target, targetType string) error {
	key := "relation_user_{" + userid + "}"
	field := targetType + "_" + target
	return d.imredis.HDel(ctx, key, field).Err()
}

//-----------------------group--------------------------------------

// members must be empty or contain group self(target == "")
func (d *Dao) RedisSetGroupMembers(ctx context.Context, groupid string, users []*model.RelationTarget) error {
	key := "relation_group_{" + groupid + "}"
	args := make([]interface{}, 0, len(users)*2+3)
	args = append(args, maxExpire, "", "")

	var groupname string
	for _, v := range users {
		if v.TargetType != "user" || v.Name == "" {
			return ecode.ErrReq
		}
		if v.Target == "" {
			groupname = v.Name
		}
		args = append(args, v.Target, strconv.Itoa(int(v.Duty))+"_"+v.Name)
	}
	if len(users) > 0 && groupname == "" {
		return ecode.ErrReq
	}
	return setRelationScript.Run(ctx, d.imredis, []string{key}, args...).Err()
}

func (d *Dao) RedisCountGroupMembers(ctx context.Context, groupid, exceptMember string) (uint64, error) {
	key := "relation_group_{" + groupid + "}"
	count, e := d.imredis.HLen(ctx, key).Uint64()
	if e != nil {
		return 0, e
	}
	if count == 0 {
		return 0, gredis.Nil
	} else if count == 1 {
		groupname, e := d.imredis.HGet(ctx, key, "").Result()
		if e != nil {
			return 0, e
		}
		if groupname == "" {
			return 0, ecode.ErrGroupNotExist
		}
	} else if exceptMember != "" {
		exist, e := d.imredis.HExists(ctx, key, exceptMember).Result()
		if e != nil {
			return 0, e
		}
		if exist {
			return count - 2, nil
		}
	}
	return count - 1, nil
}

func (d *Dao) RedisAddGroupMember(ctx context.Context, groupid, userid, username string, userduty uint8) error {
	if username == "" {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid + "}"
	args := []interface{}{maxExpire, "", userid, strconv.Itoa(int(userduty)) + "_" + username}
	r, e := addRelationScript.Run(ctx, d.imredis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrGroupNotExist
	}
	return nil
}

func (d *Dao) RedisGetGroupMembers(ctx context.Context, groupid string) ([]*model.RelationTarget, error) {
	key := "relation_group_{" + groupid + "}"
	all, e := d.imredis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, gredis.Nil
	}
	r := make([]*model.RelationTarget, 0, len(all))
	for k, v := range all {
		if k == "" && v == "" {
			return nil, ecode.ErrGroupNotExist
		}
		index := strings.Index(v, "_")
		if index <= 0 {
			return nil, ecode.ErrCacheDataBroken
		}
		duty, e := strconv.Atoi(v[:index])
		if e != nil || duty < 0 || duty >= 256 {
			return nil, ecode.ErrCacheDataBroken
		}
		r = append(r, &model.RelationTarget{
			Target:     k,
			TargetType: "user",
			Name:       v[index+1:],
			Duty:       uint8(duty),
		})
	}
	return r, nil
}

func (d *Dao) RedisGetGroupMember(ctx context.Context, groupid, userid string) (*model.RelationTarget, error) {
	key := "relation_group_{" + groupid + "}"
	rs, e := d.imredis.HMGet(ctx, key, "", userid).Result()
	if e != nil {
		return nil, e
	}
	if rs[0] == nil {
		return nil, gredis.Nil
	}
	if rs[0].(string) == "" {
		return nil, ecode.ErrGroupNotExist
	}
	if rs[1] == nil {
		return nil, ecode.ErrGroupMemberNotExist
	}
	str := rs[1].(string)
	index := strings.Index(str, "_")
	if index <= 0 {
		return nil, ecode.ErrCacheDataBroken
	}
	name := str[index+1:]
	if userid != "" && name == "" {
		return nil, ecode.ErrCacheDataBroken
	}
	duty, e := strconv.Atoi(str[:index])
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	return &model.RelationTarget{
		Target:     userid,
		TargetType: "user",
		Name:       name,
		Duty:       uint8(duty),
	}, nil
}

func (d *Dao) RedisDelGroupMembers(ctx context.Context, groupid string) error {
	key := "relation_group_{" + groupid + "}"
	return d.imredis.Del(ctx, key).Err()
}

func (d *Dao) RedisDelGroupMember(ctx context.Context, groupid, userid string) error {
	key := "relation_group_{" + groupid + "}"
	return d.imredis.HDel(ctx, key, userid).Err()
}

// -------------------------------------------gate redis-------------------------------------
func (d *Dao) GetSession(ctx context.Context, userid string) (*model.IMSession, error) {
	key := "raw_user_{" + userid + "}"
	rs, e := d.gateredis.HGetAll(ctx, key).Result()
	if e == gredis.Nil {
		e = ecode.ErrSession
	}
	if e != nil {
		return nil, e
	}
	remoteaddr, ok := rs["remote_addr"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	realip, ok := rs["real_ip"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	gate, ok := rs["gate"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	str, ok := rs["netlag"]
	if !ok {
		return nil, ecode.ErrCacheDataBroken
	}
	netlag, e := strconv.ParseUint(str, 10, 64)
	if e != nil {
		return nil, ecode.ErrCacheDataBroken
	}
	return &model.IMSession{RemoteAddr: remoteaddr, RealIP: realip, Gate: gate, Netlag: netlag}, nil
}

// data can't be nil or empty
func (d *Dao) Unicast(ctx context.Context, rawname, userid string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return d.gateredis.PubUnicast(ctx, rawname, 32, userid, userid+"_"+common.BTS(data))
}
