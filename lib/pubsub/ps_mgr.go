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
	nps.RLock()
	t, ok := nps.topicMap[name]
	nps.RUnlock()
	if ok {
		return t
	}
	return nil
}

func (nps *Pubsub) CreateTopic(name string, opts ...TopicOption) *Topic {
	nps.Lock()
	defer nps.Unlock()

	// Check again in case another goroutine created the topic
	if t, ok := nps.topicMap[name]; ok {
		return t
	}

	t := newTopic(name, nps, opts...)
	nps.topicMap[name] = t
	return t
}

func (nps *Pubsub) GetOrCreateTopic(name string, opts ...TopicOption) *Topic {
	if t := nps.GetTopic(name); t != nil {
		return t
	}
	return nps.CreateTopic(name, opts...)
}
