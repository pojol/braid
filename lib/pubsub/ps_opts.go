package pubsub

const (
	BraidPubsubTopic = "braid.pubsub.streams"
)

/*
	redis server v6.2 以上 (支持 xack.opts
	go-redis v9.0.6 以上
*/

type Parm struct {
}

// Option config wraps
type Option func(*Parm)

const (
	ReadModeBeginning = "0-0"
	ReadModeLatest    = "$"
)

type ChannelParm struct {
	ReadMode string
}

type ChannelOption func(*ChannelParm)

func WithReadMode(mode string) ChannelOption {
	return func(p *ChannelParm) {
		p.ReadMode = mode
	}
}
