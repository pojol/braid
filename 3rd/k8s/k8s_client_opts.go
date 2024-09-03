package k8s

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Parm struct {
	config *rest.Config

	ListOpts v1.ListOptions
	GetOpts  v1.GetOptions
}

type Option func(*Parm)

func WithConfigPath(path string) Option {
	return func(c *Parm) {
		var err error
		c.config, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			panic(fmt.Errorf("[braid-go] k8s WithConfigPath err %v", err))
		}
	}
}

func WithListOpts(opts v1.ListOptions) Option {
	return func(c *Parm) {
		c.ListOpts = opts
	}
}
