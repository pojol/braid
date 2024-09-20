package pubsub

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	thdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/lib/unbounded"
	"github.com/pojol/braid/router"
	"github.com/redis/go-redis/v9"
)

type Channel struct {
	topic    string
	channel  string // stream group
	consumer string // group consumer

	exitFlag int32
	msgCh    *unbounded.Unbounded
}

func newChannel(ctx context.Context, topic, channel string, p ChannelParm) (*Channel, error) {

	c := &Channel{
		topic:    topic,
		channel:  channel,
		consumer: uuid.New().String(),
		msgCh:    unbounded.NewUnbounded(),
	}

	// 从头部开始消费，还是从最新的消息开始 (默认从尾部开始进行消费，只处理新消息
	_, err := thdredis.XGroupCreate(ctx, topic, c.channel, p.ReadMode).Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return nil, err
	}
	c.loop()

	return c, nil
}

func (c *Channel) loop() {
	go func() {
		for {
			msgs := thdredis.XReadGroup(context.TODO(), &redis.XReadGroupArgs{
				Group:    c.channel,
				Consumer: c.consumer,
				Streams:  []string{c.topic, ">"},
				Block:    100 * time.Millisecond,
				Count:    10,
			}).Val()

			for _, v := range msgs {
				for _, msg := range v.Messages {

					val := msg.Values["msg"].(string)

					if atomic.LoadInt32(&c.exitFlag) == 1 {
						log.Warn("cannot write to the exiting channel %v", c.channel)
						return
					}

					msg := &router.Message{
						Header: &router.Header{
							ID:    msg.ID,
							Event: msg.Values["event"].(string),
						},
						Body: []byte(val),
					}
					c.msgCh.Put(msg)
				}
			}

		}
	}()
}

func (c *Channel) addHandlers(queue *mpsc.Queue) {
	go func() {
		for {
			m, ok := <-c.msgCh.Get()
			if !ok {
				goto EXT
			}
			c.msgCh.Load()

			pipe := thdredis.Pipeline()
			recvmsg, ok := m.(*router.Message)
			if !ok {
				log.Warn("topic %v channel %v msg is not of type *router.Message", c.topic, c.channel)
				continue
			}

			msg := router.NewMsgWrap().
				WithReqHeader(&router.Header{ID: recvmsg.Header.ID, Event: recvmsg.Header.Event}).
				WithReqBody(recvmsg.Body).Build()
			queue.Push(msg)

			pipe.XAck(context.TODO(), c.topic, c.channel, recvmsg.Header.Event)
			pipe.XDel(context.TODO(), c.topic, recvmsg.Header.ID)

			_, err := pipe.Exec(context.TODO())
			if err != nil {
				log.Warn("topic %v channel %v id %v pipeline failed: %v", c.topic, c.channel, recvmsg.Header.ID, err)
			}
		}
	EXT:
		log.Info("channel %v stopping handler", c.channel)
	}()
}

func (c *Channel) Arrived(queue *mpsc.Queue) {
	c.addHandlers(queue)
}

func (c *Channel) Close() error {

	_, err := thdredis.XGroupDelConsumer(context.TODO(), c.topic, c.channel, c.consumer).Result()
	if err != nil {
		log.Warn("braid.pubsub topic %v channel %v redis channel del consumer err %v", c.topic, c.channel, err.Error())
		return err
	}

	consumers, err := thdredis.XInfoConsumers(context.TODO(), c.topic, c.channel).Result()
	if err != nil {
		log.Warn("braid.pubsub topic %v channel %v redis channel info consumers err %v", c.topic, c.channel, err.Error())
		return err
	}

	if len(consumers) == 0 {
		_, err := thdredis.XGroupDestroy(context.TODO(), c.topic, c.channel).Result()
		if err != nil {
			log.Warn("braid.pubsub topic %v channel %v redis channel destory err %v", c.topic, c.channel, err.Error())
		}
	}

	return err
}
