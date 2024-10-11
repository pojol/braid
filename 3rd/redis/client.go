package redis

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client              *redis.Client
	defaultClientConfig = &Parm{
		redisopt: redis.Options{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			MinIdleConns: 128,
			PoolSize:     1024,
		},
	}
)

func BuildClientWithOption(opts ...Option) *redis.Client {

	for _, opt := range opts {
		opt(defaultClientConfig)
	}
	return new(defaultClientConfig)
}

func new(p *Parm) *redis.Client {
	// 创建连接池
	client = redis.NewClient(&p.redisopt)
	ctx := context.Background()
	//判断是否能够链接到数据库

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	return client
}

func GetClient() *redis.Client {
	return client
}

func MockClient(cli *redis.Client) {
	client = cli
}

// Pipeline 管道
func Pipeline() redis.Pipeliner {
	return client.Pipeline()
}

type PipelineFunc func(pipe redis.Pipeliner) error

// TxPipelined 封装事务,将命令包装在MULTI、EXEC中,并直接执行事务
func TxPipelined(ctx context.Context, from string, fn PipelineFunc) ([]redis.Cmder, error) {
	span, err := doTracing(ctx, spanTag{"TxPipelined", from})
	if err == nil {
		defer span.End(ctx)
	}
	cmds, err := client.TxPipelined(ctx, fn)
	return cmds, err
}

func Pipelined(ctx context.Context, from string, fn PipelineFunc) ([]redis.Cmder, error) {
	span, err := doTracing(ctx, spanTag{"Pipelined", from})
	if err == nil {
		defer span.End(ctx)
	}
	return client.Pipelined(ctx, fn)
}

func ConnHGet(conn *redis.Conn, ctx context.Context, key, field string) (string, error) {
	if conn == nil {
		return "", fmt.Errorf("connection is nil")
	}
	span, err := doTracing(ctx, spanTag{"cmd", "HGet"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}

	cmd := conn.HGet(context.TODO(), key, field)
	return cmd.Val(), cmd.Err()
}

// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
func Get(ctx context.Context, key string) *redis.StringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "Get"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.Get(ctx, key)
}

func Set(ctx context.Context, key string, val string) *redis.StatusCmd {
	return client.Set(ctx, key, val, 0)
}

func Del(ctx context.Context, keys ...string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "Del"})
	if err == nil {
		defer span.End(ctx)
	}
	return client.Del(ctx, keys...)
}

// SetEx Redis `SETEx key expiration value` command.
func SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SetEx"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}

	return client.SetEx(ctx, key, value, expiration)
}

func SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SetNx"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SetNX(ctx, key, value, expiration)
}

// ------------ Set 集合 ------------

func SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SAdd"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SAdd(ctx, key, members...)
}

func SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SIsMember"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SIsMember(ctx, key, member)
}

func SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SMembers"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SMembers(ctx, key)
}

func SRandMember(ctx context.Context, key string) *redis.StringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SRandMember"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SRandMember(ctx, key)
}

func SRandMemberN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SRandMemberN"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SRandMemberN(ctx, key, count)
}

func SPop(ctx context.Context, key string) *redis.StringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "SPop"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.SPop(ctx, key)
}

// ------------ ZSet 有序集合 ----------

func ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZScore"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}

	return client.ZScore(ctx, key, member)
}

func ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZAdd"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}

	return client.ZAdd(ctx, key, members...)
}

func ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZRevRank"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.ZRevRank(ctx, key, member)
}

func ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZRevRange"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.ZRevRange(ctx, key, start, stop)
}

func ZCard(ctx context.Context, key string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZCard"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.ZCard(ctx, key)
}

func ZRangeByScore(ctx context.Context, key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZRangeByScore"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.ZRangeByScore(ctx, key, &opt)
}

func ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "ZRem"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.ZRem(ctx, key, members...)
}

// ----------- Hash 哈希 ---------------

func HGet(ctx context.Context, key, field string) *redis.StringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HGet"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HGet(ctx, key, field)
}

func HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HGetAll"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HGetAll(ctx, key)
}

func HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HSet"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HSet(ctx, key, values...)
}

func HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HDel"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HDel(ctx, key, fields...)
}

func HLen(ctx context.Context, key string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HLen"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HLen(ctx, key)
}

func HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HKeys"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HKeys(ctx, key)
}

func Exists(ctx context.Context, key string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "Exists"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.Exists(ctx, key)
}

func HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "HExists"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.HExists(ctx, key, field)
}

func CAD(ctx context.Context, key, field string) bool {
	ret, err := client.Do(ctx, "CAD", key, field).Bool()
	if err != nil {
		fmt.Println("cad err", err.Error())
	}
	return ret
}

func XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XInfoGroups"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XInfoGroups(ctx, key)
}

func XGroupCreate(ctx context.Context, key, group, start string) *redis.StatusCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XGroupCreate"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XGroupCreate(ctx, key, group, start)
}

func XGroupDelConsumer(ctx context.Context, key, group, consumer string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XGroupDelConsumer"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XGroupDelConsumer(ctx, key, group, consumer)
}

func XInfoConsumers(ctx context.Context, key, group string) *redis.XInfoConsumersCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XInfoConsumers"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XInfoConsumers(ctx, key, group)
}

func XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XReadGroup"}, spanTag{"key", a.Streams[0]})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XReadGroup(ctx, a)
}

func XGroupDestroy(ctx context.Context, key, group string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XGroupDestroy"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XGroupDestroy(ctx, key, group)
}

func XLen(ctx context.Context, key string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XLen"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XLen(ctx, key)
}

func XDel(ctx context.Context, key, id string) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XDel"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XDel(ctx, key, id)
}

func ScriptRun(ctx context.Context, script *redis.Script, keys []string, args ...any) (interface{}, error) {
	span, err := doTracing(ctx, spanTag{"cmd", "ScriptRun"}, spanTag{"key", strings.Join(keys, ",")})
	if err == nil {
		defer span.End(ctx)
	}
	val, err := script.Run(ctx, client, keys, args...).Result()
	return val, err
}

func ScriptRunInt64s(ctx context.Context, script *redis.Script, keys []string, args ...any) ([]int64, error) {
	span, err := doTracing(ctx, spanTag{"cmd", "ScriptRun"}, spanTag{"key", strings.Join(keys, ",")})
	if err == nil {
		defer span.End(ctx)
	}
	val, err := script.Run(ctx, client, keys, args...).Int64Slice()
	return val, err
}

func LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "LRange"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.LRange(ctx, key, start, stop)
}

func Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "Expire"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.Expire(ctx, key, expiration)
}

func RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "RPush"}, spanTag{"key", key})
	if err == nil {
		defer span.End(ctx)
	}
	return client.RPush(ctx, key, values...)
}

func FlushDB(ctx context.Context) *redis.StatusCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "FlushDB"}, spanTag{"key", ""})
	if err == nil {
		defer span.End(ctx)
	}
	return client.FlushDB(ctx)
}

func FlushAll(ctx context.Context) *redis.StatusCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "FlushAll"}, spanTag{"key", ""})
	if err == nil {
		defer span.End(ctx)
	}
	return client.FlushAll(ctx)
}

func PoolStats(ctx context.Context) *redis.PoolStats {
	span, err := doTracing(ctx, spanTag{"cmd", "PoolStats"}, spanTag{"key", ""})
	if err == nil {
		defer span.End(ctx)
	}
	return client.PoolStats()
}

func XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	span, err := doTracing(ctx, spanTag{"cmd", "XAdd"}, spanTag{"key", ""})
	if err == nil {
		defer span.End(ctx)
	}
	return client.XAdd(ctx, a)
}

func Incr(ctx context.Context, key string) *redis.IntCmd {
	return client.Incr(ctx, key)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
