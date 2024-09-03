package k8s

import (
	"context"
	"fmt"

	"github.com/pojol/braid/def"
	v1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	clientset *kubernetes.Clientset
	parm      Parm
}

func BuildWithOption(opts ...Option) *Client {

	var clientset *kubernetes.Clientset
	var err error

	var parm Parm
	for _, v := range opts {
		v(&parm)
	}

	clientset, err = kubernetes.NewForConfig(parm.config)
	if err != nil {
		panic(fmt.Errorf("[braid-go] k8s NewForConfig err %v", err))
	}

	return &Client{
		clientset: clientset,
		parm:      parm,
	}
}

// tmp
func (c *Client) ListServices(ctx context.Context, namespace string) ([]def.Service, error) {

	var service []def.Service

	endpoints, err := c.clientset.CoreV1().Endpoints(namespace).List(ctx, c.parm.ListOpts)
	if err != nil {
		return service, err
	}

	// 遍历每个Endpoints
	for _, endpoint := range endpoints.Items {
		nods := []def.Node{}

		svc, err := c.clientset.CoreV1().Services(namespace).Get(context.TODO(), endpoint.GetName(), c.parm.GetOpts)
		if err != nil {
			return service, err
		}

		// 遍历每个子网段
		for _, subset := range endpoint.Subsets {
			for _, address := range subset.Addresses {
				nods = append(nods, def.Node{
					Name:    address.Hostname,
					ID:      address.IP,
					Address: address.IP,
				})
			}
		}

		tags := []string{}

		for _, lab := range svc.Labels {
			tags = append(tags, lab)
		}
		service = append(service, def.Service{
			Info: def.ServiceInfo{
				Name: endpoint.GetName(),
			},
			Nodes: nods,
			Tags:  tags,
		})
	}

	return service, nil

}

func (c *Client) CreateLeases(ctx context.Context, namespace, name, identity string) (string, error) {

	duration := int32(60)

	// 创建一个Lease对象
	lease, err := c.clientset.CoordinationV1().Leases(namespace).Create(ctx, &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.LeaseSpec{
			HolderIdentity:       &identity,
			LeaseDurationSeconds: &duration,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return "", err
	}

	return *lease.Spec.HolderIdentity, nil
}

func (c *Client) GetLeases(ctx context.Context, namespace, name string) (string, error) {

	lease, err := c.clientset.CoordinationV1().Leases(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return *lease.Spec.HolderIdentity, nil
}

func (c *Client) RenewLeases(ctx context.Context, namespace, name string) error {

	lease, err := c.clientset.CoordinationV1().Leases(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	lease.Reset()
	return nil
}

func (c *Client) RmvLeases(ctx context.Context, namespace, name string) error {
	return c.clientset.CoordinationV1().Leases(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
