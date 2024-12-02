package pubsub

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	thdredis "github.com/pojol/braid/3rd/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// 模拟 Redis 客户端
var mockRedis *redis.Client

func setupTest() (*miniredis.Miniredis, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	mockRedis = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	thdredis.MockClient(mockRedis)

	return mr, nil
}

func TestNewTopic(t *testing.T) {
	mr, err := setupTest()
	assert.NoError(t, err)
	defer mr.Close()

	ps := &Pubsub{}
	topic := newTopic("test_topic", ps)

	assert.NotNil(t, topic)
	assert.Equal(t, "test_topic", topic.topic)
	assert.Equal(t, ps, topic.ps)
}

func TestTopicPub(t *testing.T) {
	mr, err := setupTest()
	assert.NoError(t, err)
	defer mr.Close()

	topic := &Topic{topic: "test_topic"}

	err = topic.Pub(context.Background(), "test_event", []byte("test message"))
	assert.NoError(t, err)

	// 验证消息是否被正确添加到 Redis
	result, err := mockRedis.XRead(context.Background(), &redis.XReadArgs{
		Streams: []string{"test_topic", "0-0"}, // 使用 "0-0" 作为起始 ID
		Count:   1,
	}).Result()

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Messages, 1)
	assert.Equal(t, "test message", result[0].Messages[0].Values["msg"])
	assert.Equal(t, "test_event", result[0].Messages[0].Values["event"])
}

func TestTopicSub(t *testing.T) {
	mr, err := setupTest()
	assert.NoError(t, err)
	defer mr.Close()

	topic := &Topic{
		topic:      "test_topic",
		channelMap: make(map[string]*Channel),
	}

	// 首先，创建 Stream
	err = mockRedis.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "test_topic",
		ID:     "*",
		Values: map[string]interface{}{"key": "value"},
	}).Err()
	assert.NoError(t, err)

	// 然后尝试订阅
	channel, err := topic.Sub(context.Background(), "test_channel")
	assert.NoError(t, err)
	assert.NotNil(t, channel)

	// 验证 channel 是否被正确创建
	assert.Len(t, topic.channelMap, 1)
	assert.Contains(t, topic.channelMap, "test_channel")
}

func TestTopicClose(t *testing.T) {
	mr, err := setupTest()
	assert.NoError(t, err)
	defer mr.Close()

	topic := &Topic{topic: "test_topic"}

	// 测试场景1：空的 stream，没有消费者组
	err = topic.Close()
	assert.NoError(t, err)
	exists, _ := mockRedis.Exists(context.Background(), "test_topic").Result()
	assert.Equal(t, int64(0), exists)

	// 测试场景2：非空的 stream，没有消费者组
	_, err = mockRedis.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "test_topic",
		Values: map[string]interface{}{"key": "value"},
	}).Result()
	assert.NoError(t, err)

	err = topic.Close()
	assert.NoError(t, err)
	exists, _ = mockRedis.Exists(context.Background(), "test_topic").Result()
	assert.Equal(t, int64(1), exists, "Non-empty stream without consumer groups should not be deleted")

	// 清理 stream 以准备下一个测试
	mockRedis.Del(context.Background(), "test_topic")

	// 测试场景3：空的 stream，有消费者组
	err = mockRedis.XGroupCreateMkStream(context.Background(), "test_topic", "test_group", "$").Err()
	assert.NoError(t, err)

	err = topic.Close()
	assert.NoError(t, err)
	exists, _ = mockRedis.Exists(context.Background(), "test_topic").Result()
	assert.Equal(t, int64(1), exists, "Stream with consumer groups should not be deleted")

	// 测试场景4：非空的 stream，有消费者组
	_, err = mockRedis.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "test_topic",
		Values: map[string]interface{}{"key": "value"},
	}).Result()
	assert.NoError(t, err)

	err = topic.Close()
	assert.NoError(t, err)
	exists, _ = mockRedis.Exists(context.Background(), "test_topic").Result()
	assert.Equal(t, int64(1), exists, "Non-empty stream with consumer groups should not be deleted")

	// 额外检查：确保 BraidPubsubTopic 集合中的条目被正确处理
	err = mockRedis.SAdd(context.Background(), BraidPubsubTopic, "test_topic").Err()
	assert.NoError(t, err)

	isMember, err := mockRedis.SIsMember(context.Background(), BraidPubsubTopic, "test_topic").Result()
	assert.NoError(t, err)
	assert.True(t, isMember, "Topic should still be a member of BraidPubsubTopic set")

	// 清空 stream 并再次调用 Close
	mockRedis.Del(context.Background(), "test_topic")
	err = topic.Close()
	assert.NoError(t, err)

	isMember, err = mockRedis.SIsMember(context.Background(), BraidPubsubTopic, "test_topic").Result()
	assert.NoError(t, err)
	assert.False(t, isMember, "Topic should be removed from BraidPubsubTopic set when deleted")
}
