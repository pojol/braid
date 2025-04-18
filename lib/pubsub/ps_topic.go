package pubsub

import (
	"context"
	"fmt"
	"strings"
	"sync"

	thdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/lib/log"
	"github.com/redis/go-redis/v9"
)

type Topic struct {
	sync.RWMutex

	topic string

	ps *Pubsub

	channelMap map[string]*Channel
}

func newTopic(name string, mgr *Pubsub, opts ...TopicOption) *Topic {

	rt := &Topic{
		ps:         mgr,
		topic:      name,
		channelMap: make(map[string]*Channel),
	}

	ctx := context.TODO()

	options := &topicOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cnt, _ := thdredis.Exists(ctx, rt.topic).Result()
	if cnt == 0 {
		id, err := thdredis.XAdd(ctx, &redis.XAddArgs{
			Stream: rt.topic,
			Values: []string{"msg", "init", "event", ""},
		}).Result()

		if err != nil {
			log.WarnF("[braid.pubsub ]Topic %v init failed %v", rt.topic, err)
		} else {

			thdredis.XDel(ctx, rt.topic, id)
			if options.ttl > 0 {
				err = thdredis.Expire(ctx, rt.topic, options.ttl).Err()
				if err != nil {
					log.WarnF("[braid.pubsub ]Failed to set TTL for topic %v: %v", rt.topic, err)
				}
			}

			err = thdredis.SAdd(ctx, BraidPubsubTopic, rt.topic).Err()
			if err != nil {
				log.WarnF("[braid.pubsub] Failed to add topic %v to BraidPubsubTopic set: %v", rt.topic, err)
			}

		}

	}

	return rt
}

func (rt *Topic) Pub(ctx context.Context, event string, body []byte) error {

	if event == "" {
		return fmt.Errorf("cannot send a message without an event")
	}

	_, err := thdredis.XAdd(ctx, &redis.XAddArgs{
		Stream: rt.topic,
		ID:     "*",
		Values: []string{"msg", string(body), "event", event},
	}).Result()

	return err
}

func (rt *Topic) Sub(ctx context.Context, channel string, opts ...interface{}) (*Channel, error) {
	p := ChannelParm{
		ReadMode: ReadModeLatest,
	}

	for _, opt := range opts {
		copt, ok := opt.(ChannelOption)
		if ok {
			copt(&p)
		}
	}

	rt.Lock()
	c, err := rt.getOrCreateChannel(ctx, channel, p)
	rt.Unlock()

	return c, err
}

func (rt *Topic) Close() error {

	ctx := context.Background()
	groups, err := thdredis.XInfoGroups(ctx, rt.topic).Result()
	if err != nil && err != redis.Nil {
		// 忽略 "no such key" 错误
		if !strings.Contains(err.Error(), "no such key") {
			return fmt.Errorf("failed to get XInfoGroups: %w", err)
		}
		groups = []redis.XInfoGroup{} // 设置为空切片
		err = nil
	}

	if len(groups) == 0 {
		cnt, err := thdredis.XLen(ctx, rt.topic).Result()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get XLen: %w", err)
		}

		if cnt == 0 {
			cleanpipe := thdredis.Pipeline()
			cleanpipe.Del(ctx, rt.topic)
			cleanpipe.SRem(ctx, BraidPubsubTopic, rt.topic)

			_, err = cleanpipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to clean topic %s: %w", rt.topic, err)
			}
			log.InfoF("[braid.pubsub] Topic %v cleaned successfully", rt.topic)
		} else {
			log.InfoF("[braid.pubsub] Topic %v not cleaned: non-empty stream", rt.topic)
		}
	}

	return err
}

func (rt *Topic) getOrCreateChannel(ctx context.Context, name string, p ChannelParm) (*Channel, error) {

	//channel, ok := rt.channelMap[name]
	//var err error
	//if !ok {
	channel, err := newChannel(ctx, rt.topic, name, p)
	if err != nil {
		return nil, err
	}
	rt.channelMap[name] = channel

	log.InfoF("[braid.pubsub ]Topic %v new channel %v", rt.topic, name)
	return channel, nil
	//}

	//return channel, nil
}
