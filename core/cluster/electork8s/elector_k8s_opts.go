package electork8s

import "time"

// Parm 选举器配置项
type Parm struct {
	WatchTick time.Duration

	RefreshTick time.Duration

	Namespace string

	Name string
}

// Option consul discover config wrapper
type Option func(*Parm)

func WithNamespace(ns string) Option {
	return func(c *Parm) {
		c.Namespace = ns
	}
}

func WithName(name string) Option {
	return func(c *Parm) {
		c.Name = name
	}
}

func WithWatchTick(t time.Duration) Option {
	return func(c *Parm) {
		c.WatchTick = t
	}
}

func WithRefreshTick(t time.Duration) Option {
	return func(c *Parm) {
		c.RefreshTick = t
	}
}
