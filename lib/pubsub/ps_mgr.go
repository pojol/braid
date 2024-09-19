package pubsub

import (
	"sync"
)

type Pubsub struct {
	parm Parm

	sync.RWMutex

	topicMap map[string]*Topic
}

func BuildWithOption(opts ...Option) *Pubsub {

	p := Parm{}

	for _, opt := range opts {
		opt(&p)
	}

	ps := &Pubsub{
		parm:     p,
		topicMap: make(map[string]*Topic),
	}

	return ps
}

func (nps *Pubsub) GetTopic(name string) *Topic {
	var t *Topic

	nps.RLock()
	t, ok := nps.topicMap[name]
	nps.RUnlock()
	if ok {
		return t
	}

	nps.Lock()
	t = newTopic(name, nps)
	nps.Unlock()

	return t
}
