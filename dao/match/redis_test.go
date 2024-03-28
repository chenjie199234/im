package match

import (
	"context"
	"testing"
	"time"

	"github.com/chenjie199234/im/ecode"

	"github.com/chenjie199234/Corelib/redis"
	"github.com/chenjie199234/Corelib/util/ctime"
)

func Test_Redis(t *testing.T) {
	c := &redis.Config{
		RedisName:       "test",
		RedisMode:       "direct",
		Addrs:           []string{"127.0.0.1:6379"},
		MaxOpen:         256,
		MaxConnIdletime: ctime.Duration(time.Minute * 5),
		DialTimeout:     ctime.Duration(time.Second),
		IOTimeout:       ctime.Duration(time.Second),
	}
	client, e := redis.NewRedis(c, nil)
	if e != nil {
		t.Fatal(e)
		return
	}
	dao := NewDao(nil, client, nil)
	status, e := dao.RedisGetUserMatchStatus(context.Background(), "testuser")
	if e != nil {
		t.Fatal(e)
		return
	}
	if status != nil {
		t.Fatal("status should be nil")
		return
	}
	if e := dao.RedisCancelMatch(context.Background(), "testuser"); e != nil {
		t.Fatal(e)
		return
	}
	if e := dao.RedisStartMatch(context.Background(), "testuser", "zh_CN", "上海", []string{"杨浦", "徐汇", "浦东"}, []string{"摄影", "聚餐", "摩旅"}); e != nil {
		t.Fatal(e)
		return
	}
	status, e = dao.RedisGetUserMatchStatus(context.Background(), "testuser")
	if e != nil {
		t.Fatal(e)
		return
	}
	if status == nil {
		t.Fatal("should not be nil")
		return
	}
	t.Log(status)
	if e := dao.RedisStartMatch(context.Background(), "testuser", "zh_CN", "上海", []string{"杨浦", "浦东"}, []string{"摄影", "摩旅"}); e != ecode.ErrAlreadyInMatching {
		t.Fatal(e)
		return
	}
	if _, e := client.HSet(context.Background(), "match_testuser", "_status_", 1209377).Result(); e != nil {
		t.Fatal(e)
		return
	}
	status, e = dao.RedisGetUserMatchStatus(context.Background(), "testuser")
	if e != nil {
		t.Fatal(e)
		return
	}
	if status == nil {
		t.Fatal("should not be nil")
		return
	}
	t.Log(status)
	if e := dao.RedisStartMatch(context.Background(), "testuser", "en_US", "上海", []string{"杨浦", "青浦"}, []string{"自驾", "电影"}); e != ecode.ErrAlreadyMatched {
		t.Fatal(e)
		return
	}
	if e := dao.RedisCancelMatch(context.Background(), "testuser"); e != ecode.ErrAlreadyMatched {
		t.Fatal(e)
		return
	}
	if _, e := client.HSet(context.Background(), "match_testuser", "_status_", 0).Result(); e != nil {
		t.Fatal(e)
		return
	}
	if e := dao.RedisStartMatch(context.Background(), "anotheruser", "zh_CN", "上海", []string{"徐汇", "宝山", "金山"}, []string{"摄影", "自驾", "电影"}); e != nil {
		t.Fatal(e)
		return
	}
	if e := dao.RedisMatchSuccess(context.Background(), []string{"testuser", "anotheruser"}, "徐汇", "摄影"); e != nil {
		t.Fatal(e)
		return
	}
	status, e = dao.RedisGetUserMatchStatus(context.Background(), "anotheruser")
	if e != nil {
		t.Fatal(e)
		return
	}
	t.Log(status)
}
