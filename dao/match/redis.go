package match

import (
	"context"
	"strconv"
	"time"

	"github.com/chenjie199234/im/api"
	"github.com/chenjie199234/im/ecode"
	gredis "github.com/redis/go-redis/v9"
)

const defaultExpire = 21600 //6 hours

var startmatch *gredis.Script
var cancelmatch *gredis.Script
var matchsuccess *gredis.Script

func init() {
	startmatch = gredis.NewScript(`local tmp=redis.call("HGET","match_"..KEYS[1],"_status_")
if(tmp) then
	if(tmp == "0") then
		return -1
	else
		return -2
	end
end
local lan=""
local city=""
local regions={}
local activities={}
for i=1,#ARGV,2 do
	if(ARGV[i+1]=="0") then
		lan=ARGV[i]
	elseif(ARGV[i+1]=="1") then
		city=ARGV[i]
	elseif(ARGV[i+1]=="2") then
		table.insert(regions,ARGV[i])
	elseif(ARGV[i+1]=="3") then
		table.insert(activities,ARGV[i])
	end
end
table.insert(ARGV,"_status_")
table.insert(ARGV,0)
table.insert(ARGV,"_matched_region_")
table.insert(ARGV,"")
table.insert(ARGV,"_matched_activity_")
table.insert(ARGV,"")
redis.call("HSET","match_"..KEYS[1],unpack(ARGV))
for i=1,#regions,1 do
	for j=1,#activities,1 do
		local k="match_"..lan.."_"..city.."_"..regions[i].."_"..activities[j]
		redis.call("SADD",k,KEYS[1])
	end
end
return 1`)
	cancelmatch = gredis.NewScript(`if(redis.call("EXISTS","match_"..KEYS[1])==0) then
	return 1
end
local tmp=redis.call("HGETALL","match_"..KEYS[1])
local lan=""
local city=""
local regions={}
local activities={}
for i=1,#tmp,2 do
	if(tmp[i]=="_status_") then
		if(tmp[i+1]~="0") then
			return 0
		end
	elseif(tmp[i]=="_matched_region_" or tmp[i]=="_matched_activity_") then
	elseif(tmp[i+1]=="0") then
		lan=tmp[i]
	elseif(tmp[i+1]=="1") then
		city=tmp[i]
	elseif(tmp[i+1]=="2") then
		table.insert(regions,tmp[i])
	elseif(tmp[i+1]=="3") then
		table.insert(activities,tmp[i])
	end
end
redis.call("DEL","match_"..KEYS[1])
for i=1,#regions,1 do
	for j=1,#activities,1 do
		local k="match_"..lan.."_"..city.."_"..regions[i].."_"..activities[j]
		redis.call("SREM",k,KEYS[1])
	end
end
return 1`)
	matchsuccess = gredis.NewScript(`for i=1,#KEYS,1 do
	local tmp=redis.call("HGET","match_"..KEYS[i],"_status_")
	if(not tmp or tmp~="0")then
		return 0
	end
end
for i=1,#KEYS,1 do
	local tmp=redis.call("HGETALL","match_"..KEYS[i])
	local lan=""
	local city=""
	local regions={}
	local activities={}
	for x=1,#tmp,2 do
		if(tmp[x]~="_status_" and tmp[x]~="_matched_region_" and tmp[x]~="_matched_activity_") then
			if(tmp[x+1]=="0") then
				lan=tmp[x]
			elseif(tmp[x+1]=="1") then
				city=tmp[x]
			elseif(tmp[x+1]=="2") then
				table.insert(regions,tmp[x])
			elseif(tmp[x+1]=="3") then
				table.insert(activities,tmp[x])
			end
		end
	end
	for x=1,#regions,1 do
		for y=1,#activities,1 do
			local k="match_"..lan.."_"..city.."_"..regions[x].."_"..activities[y]
			redis.call("SREM",k,KEYS[i])
		end
	end
	redis.call("HSET","match_"..KEYS[i],"_status_",ARGV[2],"_matched_region_",ARGV[3],"_matched_activity_",ARGV[4])
	redis.call("EXPIRE","match_"..KEYS[i],ARGV[1])
end
return #KEYS`)
}
func (d *Dao) RedisStartMatch(ctx context.Context, userid, lan, city string, regions, activities []string) error {
	args := make([]interface{}, 0, (len(regions)+len(activities)+2)*2)
	args = append(args, lan, 0, city, 1)
	for _, region := range regions {
		args = append(args, region, 2)
	}
	for _, activity := range activities {
		args = append(args, activity, 3)
	}
	r, e := startmatch.Run(ctx, d.redis, []string{userid}, args...).Int()
	if e != nil {
		return e
	}
	if r == -1 {
		return ecode.ErrAlreadyInMatching
	}
	if r == -2 {
		return ecode.ErrAlreadyMatched
	}
	return nil
}
func (d *Dao) RedisGetUserMatchStatus(ctx context.Context, userid string) (*api.StatusResp, error) {
	kvs, e := d.redis.HGetAll(ctx, "match_"+userid).Result()
	if e != nil && e != gredis.Nil {
		return nil, e
	}
	if e == gredis.Nil || len(kvs) == 0 {
		return nil, nil
	}
	r := &api.StatusResp{}
	for k, v := range kvs {
		if k == "_status_" {
			status, e := strconv.ParseUint(v, 10, 32)
			if e != nil {
				return nil, ecode.ErrCacheDataBroken
			}
			r.Status = uint32(status)
			continue
		}
		if k == "_matched_region_" {
			r.MatchedRegion = v
		}
		if k == "_matched_activity_" {
			r.MatchedActivity = v
		}
		switch v {
		case "0":
			r.Lan = k
		case "1":
			r.City = k
		case "2":
			r.Regions = append(r.Regions, k)
		case "3":
			r.Activities = append(r.Activities, k)
		}
	}
	if r.Lan == "" || r.City == "" || len(r.Regions) == 0 || len(r.Activities) == 0 {
		return nil, ecode.ErrCacheDataBroken
	}
	return r, nil
}
func (d *Dao) RedisCancelMatch(ctx context.Context, userid string) error {
	r, e := cancelmatch.Run(ctx, d.redis, []string{userid}).Int()
	if e == nil && r == 0 {
		e = ecode.ErrAlreadyMatched
	}
	return e
}
func (d *Dao) RedisGetWaitList(ctx context.Context, lan, city, region, activity string) ([]string, error) {
	key := "match_" + lan + "_" + city + "_" + region + "_" + activity
	return d.redis.SMembers(ctx, key).Result()
}
func (d *Dao) RedisMatchSuccess(ctx context.Context, userids []string, region, activity string) error {
	r, e := matchsuccess.Run(ctx, d.redis, userids, defaultExpire, time.Now().Unix(), region, activity).Int()
	if e == nil && r == 0 {
		e = ecode.ErrUserConflictInMatching
	}
	return e
}
