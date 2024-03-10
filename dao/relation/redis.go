package relation

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/chenjie199234/im/ecode"
	"github.com/chenjie199234/im/model"

	gredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const defaultExpire = 86400
const maxExpire = 604800

// in redis's SortedSet,member num <=128,use listpack to store
const defaultRequest = 32
const maxRequest = 128

var addRequestScript *gredis.Script
var getRequestScript *gredis.Script
var refreshRequestScript *gredis.Script

var setRelationScript *gredis.Script
var addRelationScript *gredis.Script
var refreshRelationScript *gredis.Script

func init() {
	addRequestScript = gredis.NewScript(`local time=redis.call("TIME")
redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]-ARGV[2])
redis.call("ZREM",KEYS[1],ARGV[1])
if(redis.call("ZCARD",KEYS[1])>=ARGV[3])
then
	return -1
end
local num=redis.call("ZADD",KEYS[1],time[1],ARGV[1])
redis.call("EXPIRE",KEYS[1],ARGV[2])
return num`)
	getRequestScript = gredis.NewScript(`local time=redis.call("TIME")
redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]-ARGV[1])
redis.call("EXPIRE",KEYS[1],ARGV[1])
return redis.call("ZRANGE",0,-1)`)
	refreshRequestScript = gredis.NewScript(`local time=redis.call("TIME")
redis.call("ZREMRANGEBYSCORE",KEYS[1],0,time[1]-ARGV[2])
if(redis.call("ZSCORE",KEYS[1],ARGV[1])==nil)
then
	return nil
end
local num=redis.call("ZADD",KEYS[1],time[1],ARGV[1])
redis.call("EXPIRE",KEYS[1],ARGV[2])
return num`)

	setRelationScript = gredis.NewScript(`redis.call("DEL",KEYS[1])
local num=redis.call("HSET",KEYS[1],unpack(ARGV,2)
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
	addRelationScript = gredis.NewScript(`if(redis.call("EXISTS",KEYS[1])==0)
then
	return nil
end
local tmp=redis.call("HGET",KEYS[1],ARGV[2])
if(tmp==nil or tmp=="")
then
	return -1
end
local num=redis.call("HSET",KEYS[1],ARGV[3],ARGV[4])
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
	refreshRelationScript = gredis.NewScript(`if(redis.call("EXISTS",KEYS[1])==0)
then
	return nil
end
local tmp=redis.call("HGET",KEYS[1],ARGV[2])
if(tmp==nil or tmp=="")
then
	return -1
end
if(redis.call("HEXISTS",KEYS[1],ARGV[3])==0)
then
	return 0
end
local num=redis.call("HSET",KEYS[1],ARGV[3],ARGV[4])
redis.call("EXPIRE",KEYS[1],ARGV[1])
return num`)
}

//-----------------------request--------------------------------------

func (d *Dao) RedisAddMakeFriendRequest(ctx context.Context, requester, accepter primitive.ObjectID) error {
	if requester.IsZero() || accepter.IsZero() {
		return ecode.ErrReq
	}
	key := "request_user_{" + accepter.Hex() + "}"
	value := "user_" + requester.Hex()
	num, e := addRequestScript.Run(ctx, d.redis, []string{key}, value, defaultExpire, defaultRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}

func (d *Dao) RedisAddGroupInviteRequest(ctx context.Context, accepter, groupid primitive.ObjectID) error {
	if accepter.IsZero() || groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "request_user_{" + accepter.Hex() + "}"
	value := "group_" + groupid.Hex()
	num, e := addRequestScript.Run(ctx, d.redis, []string{key}, value, defaultExpire, defaultRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}

func (d *Dao) RedisAddGroupApplyRequest(ctx context.Context, requester, groupid primitive.ObjectID) error {
	if requester.IsZero() || groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "request_group_{" + groupid.Hex() + "}"
	num, e := addRequestScript.Run(ctx, d.redis, []string{key}, requester.Hex(), defaultExpire, maxRequest).Int()
	if e != nil {
		return e
	}
	if num == -1 {
		return ecode.ErrTooPopular
	}
	return nil
}

// return key:userid or groupid,value:user or group
func (d *Dao) RedisGetUserRequests(ctx context.Context, userid primitive.ObjectID) (map[string]string, error) {
	if userid.IsZero() {
		return nil, ecode.ErrReq
	}
	key := "request_user_{" + userid.Hex() + "}"
	strs, e := getRequestScript.Run(ctx, d.redis, []string{key}, defaultExpire).StringSlice()
	if e != nil {
		return nil, e
	}
	r := make(map[string]string, len(strs))
	for _, str := range strs {
		pieces := strings.Split(str, "_")
		if len(pieces) != 2 || (pieces[0] != "user" && pieces[0] != "group") {
			return nil, ecode.ErrCacheDataBroken
		}
		r[pieces[0]] = pieces[1]
	}
	return r, nil
}
func (d *Dao) RedisRefreshUserRequest(ctx context.Context, userid, target primitive.ObjectID, targetType string) error {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") {
		return ecode.ErrReq
	}
	key := "request_user_{" + userid.Hex() + "}"
	value := targetType + "_" + target.Hex()
	e := refreshRequestScript.Run(ctx, d.redis, []string{key}, value, defaultExpire).Err()
	if e != nil && e == gredis.Nil {
		e = ecode.ErrRequestNotExist
	}
	return e
}
func (d *Dao) RedisDelUserRequest(ctx context.Context, userid, target primitive.ObjectID, targetType string) error {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") {
		return ecode.ErrReq
	}
	key := "request_user_{" + userid.Hex() + "}"
	value := targetType + "_" + target.Hex()
	return d.redis.ZRem(ctx, key, value).Err()
}

// return userids
func (d *Dao) RedisGetGroupRequests(ctx context.Context, groupid primitive.ObjectID) ([]string, error) {
	if groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	key := "request_group_{" + groupid.Hex() + "}"
	return getRequestScript.Run(ctx, d.redis, []string{key}, defaultExpire).StringSlice()
}
func (d *Dao) RedisRefreshGroupRequest(ctx context.Context, groupid, userid primitive.ObjectID) error {
	if groupid.IsZero() || userid.IsZero() {
		return ecode.ErrReq
	}
	key := "request_group_{" + groupid.Hex() + "}"
	e := refreshRequestScript.Run(ctx, d.redis, []string{key}, userid.Hex(), defaultExpire).Err()
	if e != nil && e == gredis.Nil {
		e = ecode.ErrRequestNotExist
	}
	return e
}
func (d *Dao) RedisDelGroupRequest(ctx context.Context, groupid, userid primitive.ObjectID) error {
	if groupid.IsZero() || userid.IsZero() {
		return ecode.ErrReq
	}
	key := "request_group_{" + groupid.Hex() + "}"
	return d.redis.ZRem(ctx, key, userid.Hex()).Err()
}

//-----------------------user--------------------------------------

func (d *Dao) RedisSetUserName(ctx context.Context, userid primitive.ObjectID, name string) error {
	if userid.IsZero() {
		return ecode.ErrReq
	}
	key := "name_user_{" + userid.Hex() + "}"
	return d.redis.SetEx(ctx, key, name, maxExpire*time.Second).Err()
}

func (d *Dao) RedisGetUserName(ctx context.Context, userid primitive.ObjectID) (string, error) {
	if userid.IsZero() {
		return "", ecode.ErrReq
	}
	key := "name_user_{" + userid.Hex() + "}"
	username, e := d.redis.GetEx(ctx, key, maxExpire*time.Second).Result()
	if e != nil {
		return "", e
	}
	if username == "" {
		return "", ecode.ErrUserNotExist
	}
	return username, nil
}

func (d *Dao) RedisDelUserName(ctx context.Context, userid primitive.ObjectID) error {
	if userid.IsZero() {
		return ecode.ErrReq
	}
	key := "name_user_{" + userid.Hex() + "}"
	return d.redis.Del(ctx, key).Err()
}

// targets must be empty or contain user self(target == primitive.NilObjectID)
func (d *Dao) RedisSetUserRelations(ctx context.Context, userid primitive.ObjectID, targets []*model.RelationTarget) error {
	if userid.IsZero() {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	args := make([]interface{}, 0, len(targets)*2+3)
	args = append(args, maxExpire, "user_"+primitive.NilObjectID.Hex(), "")

	var selfname string
	for _, v := range targets {
		if (v.TargetType != "user" && v.TargetType != "group") || v.Name == "" {
			return ecode.ErrReq
		}
		if v.Target.IsZero() {
			selfname = v.Name
		}
		args = append(args, v.TargetType+"_"+v.Target.Hex(), v.Name)
	}
	if len(targets) > 0 && selfname == "" {
		return ecode.ErrReq
	}
	return setRelationScript.Run(ctx, d.redis, []string{key}, args...).Err()
}

func (d *Dao) RedisCountUserRelations(ctx context.Context, userid, exceptTarget primitive.ObjectID, exceptTargetType string) (uint64, error) {
	if userid.IsZero() || (!exceptTarget.IsZero() && exceptTargetType != "user" && exceptTargetType != "group") {
		return 0, ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	count, e := d.redis.HLen(ctx, key).Uint64()
	if e != nil {
		return 0, e
	}
	if count == 0 {
		return 0, gredis.Nil
	} else if count == 1 {
		username, e := d.redis.HGet(ctx, key, "user_"+primitive.NilObjectID.Hex()).Result()
		if e != nil {
			return 0, e
		}
		if username == "" {
			return 0, ecode.ErrUserNotExist
		}
	} else if !exceptTarget.IsZero() {
		exist, e := d.redis.HExists(ctx, key, exceptTargetType+"_"+exceptTarget.Hex()).Result()
		if e != nil {
			return 0, e
		}
		if exist {
			return count - 2, nil
		}
	}
	return count - 1, nil
}

func (d *Dao) RedisAddUserRelation(ctx context.Context, userid primitive.ObjectID, target primitive.ObjectID, targetType, targetname string) error {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") || targetname == "" {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	args := []interface{}{maxExpire, "user_" + primitive.NilObjectID.Hex(), targetType + "_" + target.Hex(), targetname}
	r, e := addRelationScript.Run(ctx, d.redis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrUserNotExist
	}
	return nil
}

// won't add new relation,if the relation exist:update,if the relation not exist:do nothing
func (d *Dao) RedisRefreshUserRelation(ctx context.Context, userid primitive.ObjectID, target primitive.ObjectID, targetType, targetname string) error {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") || targetname == "" {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	args := []interface{}{maxExpire, "user_" + primitive.NilObjectID.Hex(), targetType + "_" + target.Hex(), targetname}
	r, e := refreshRelationScript.Run(ctx, d.redis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrUserNotExist
	}
	return nil
}

func (d *Dao) RedisGetUserRelations(ctx context.Context, userid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if userid.IsZero() {
		return nil, ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	all, e := d.redis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, gredis.Nil
	}
	r := make([]*model.RelationTarget, 0, len(all))
	for k, name := range all {
		if strings.HasPrefix(k, "user_") {
			targetType := "user"
			target, e := primitive.ObjectIDFromHex(k[5:])
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			if target.IsZero() && name == "" {
				return nil, ecode.ErrUserNotExist
			}
			tmp := &model.RelationTarget{
				Target:     target,
				TargetType: targetType,
				Name:       name,
			}
			r = append(r, tmp)
		} else if strings.HasPrefix(k, "group_") {
			targetType := "group"
			target, e := primitive.ObjectIDFromHex(k[6:])
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			r = append(r, &model.RelationTarget{
				Target:     target,
				TargetType: targetType,
				Name:       name,
			})
		} else {
			return nil, ecode.ErrCacheDataBroken
		}
	}
	return r, nil
}

func (d *Dao) RedisGetUserRelation(ctx context.Context, userid, target primitive.ObjectID, targetType string) (*model.RelationTarget, error) {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") {
		return nil, ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	rs, e := d.redis.HMGet(ctx, key, "user_"+primitive.NilObjectID.Hex(), targetType+"_"+target.Hex()).Result()
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
		}
		return nil, ecode.ErrNotInGroup
	}
	return &model.RelationTarget{
		Target:     target,
		TargetType: targetType,
		Name:       rs[1].(string),
	}, nil
}

func (d *Dao) RedisDelUserRelations(ctx context.Context, userid primitive.ObjectID) error {
	if userid.IsZero() {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	return d.redis.Del(ctx, key).Err()
}

func (d *Dao) RedisDelUserRelation(ctx context.Context, userid, target primitive.ObjectID, targetType string) error {
	if userid.IsZero() || target.IsZero() || (targetType != "user" && targetType != "group") {
		return ecode.ErrReq
	}
	key := "relation_user_{" + userid.Hex() + "}"
	field := targetType + "_" + target.Hex()
	return d.redis.HDel(ctx, key, field).Err()
}

//-----------------------group--------------------------------------

func (d *Dao) RedisSetGroupName(ctx context.Context, groupid primitive.ObjectID, name string) error {
	if groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "name_group_{" + groupid.Hex() + "}"
	return d.redis.SetEx(ctx, key, name, maxExpire*time.Second).Err()
}

func (d *Dao) RedisGetGroupName(ctx context.Context, groupid primitive.ObjectID) (string, error) {
	if groupid.IsZero() {
		return "", ecode.ErrReq
	}
	key := "name_group_{" + groupid.Hex() + "}"
	groupname, e := d.redis.GetEx(ctx, key, maxExpire*time.Second).Result()
	if e != nil {
		return "", e
	}
	if groupname == "" {
		return "", ecode.ErrGroupNotExist
	}
	return groupname, nil
}

func (d *Dao) RedisDelGroupName(ctx context.Context, groupid primitive.ObjectID) error {
	if groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "name_group_{" + groupid.Hex() + "}"
	return d.redis.Del(ctx, key).Err()
}

// members must be empty or contain group self(target == primitive.NilObjectID)
func (d *Dao) RedisSetGroupMembers(ctx context.Context, groupid primitive.ObjectID, users []*model.RelationTarget) error {
	if groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	args := make([]interface{}, 0, len(users)*2+3)
	args = append(args, maxExpire, primitive.NilObjectID.Hex(), "")

	var groupname string
	for _, v := range users {
		if v.TargetType != "user" || v.Name == "" {
			return ecode.ErrReq
		}
		if v.Target.IsZero() {
			groupname = v.Name
		}
		args = append(args, v.Target.Hex(), strconv.Itoa(int(v.Duty))+"_"+v.Name)
	}
	if len(users) > 0 && groupname == "" {
		return ecode.ErrReq
	}
	return setRelationScript.Run(ctx, d.redis, []string{key}, args...).Err()
}

func (d *Dao) RedisCountGroupMembers(ctx context.Context, groupid primitive.ObjectID, exceptMember primitive.ObjectID) (uint64, error) {
	if groupid.IsZero() {
		return 0, ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	count, e := d.redis.HLen(ctx, key).Uint64()
	if e != nil {
		return 0, e
	}
	if count == 0 {
		return 0, gredis.Nil
	} else if count == 1 {
		groupname, e := d.redis.HGet(ctx, key, primitive.NilObjectID.Hex()).Result()
		if e != nil {
			return 0, e
		}
		if groupname == "" {
			return 0, ecode.ErrGroupNotExist
		}
	} else if !exceptMember.IsZero() {
		exist, e := d.redis.HExists(ctx, key, exceptMember.Hex()).Result()
		if e != nil {
			return 0, e
		}
		if exist {
			return count - 2, nil
		}
	}
	return count - 1, nil
}

func (d *Dao) RedisAddGroupMember(ctx context.Context, groupid, userid primitive.ObjectID, username string, userduty uint8) error {
	if groupid.IsZero() || userid.IsZero() || username == "" {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	args := []interface{}{maxExpire, primitive.NilObjectID.Hex(), userid.Hex(), strconv.Itoa(int(userduty)) + "_" + username}
	r, e := addRelationScript.Run(ctx, d.redis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrGroupNotExist
	}
	return nil
}

// won't add new relation,if the relation exist:update,if the relation not exist:do nothing
func (d *Dao) RedisRefreshGroupMember(ctx context.Context, groupid, userid primitive.ObjectID, username string, userduty uint8) error {
	if groupid.IsZero() || userid.IsZero() || username == "" {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	args := []interface{}{maxExpire, primitive.NilObjectID.Hex(), userid.Hex(), strconv.Itoa(int(userduty)) + "_" + username}
	r, e := refreshRelationScript.Run(ctx, d.redis, []string{key}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrGroupNotExist
	}
	return nil
}

func (d *Dao) RedisGetGroupMembers(ctx context.Context, groupid primitive.ObjectID) ([]*model.RelationTarget, error) {
	if groupid.IsZero() {
		return nil, ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	all, e := d.redis.HGetAll(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, gredis.Nil
	}
	r := make([]*model.RelationTarget, 0, len(all))
	for k, v := range all {
		target, e := primitive.ObjectIDFromHex(k)
		if e != nil {
			return nil, ecode.ErrCacheDataBroken
		}
		if target.IsZero() && v == "" {
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
		name := v[index+1:]
		tmp := &model.RelationTarget{
			Target:     target,
			TargetType: "user",
			Name:       name,
			Duty:       uint8(duty),
		}
		r = append(r, tmp)
	}
	return r, nil
}

func (d *Dao) RedisGetGroupMember(ctx context.Context, groupid, userid primitive.ObjectID) (*model.RelationTarget, error) {
	if groupid.IsZero() || userid.IsZero() {
		return nil, ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	rs, e := d.redis.HMGet(ctx, key, primitive.NilObjectID.Hex(), userid.Hex()).Result()
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
	if name == "" {
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

func (d *Dao) RedisDelGroupMembers(ctx context.Context, groupid primitive.ObjectID) error {
	if groupid.IsZero() {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	return d.redis.Del(ctx, key).Err()
}

func (d *Dao) RedisDelGroupMember(ctx context.Context, groupid, userid primitive.ObjectID) error {
	if groupid.IsZero() || userid.IsZero() {
		return ecode.ErrReq
	}
	key := "relation_group_{" + groupid.Hex() + "}"
	return d.redis.HDel(ctx, key, userid.Hex()).Err()
}
