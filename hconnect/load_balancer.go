package hconnect

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog/v2"
)

const (
	hcloudLoadBalancerENVVar    = "HCLOUD_LOAD_BALANCER"
	hcloudPrivateNetworksENVVar = "HCLOUD_USE_PRIVATE_NETWORK"
)

type LoadBalancer struct {
	loadBalancerID int
	privateNetwork bool
}

func newLoadBalancer(c *hcloud.Client) (*LoadBalancer, error) {
	const op = "hcloud-connect/newLoadBalancer"

	var loadBalancerID int
	if v, ok := os.LookupEnv(hcloudLoadBalancerENVVar); ok {
		n, _, err := c.LoadBalancer.Get(context.Background(), v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if n == nil {
			return nil, fmt.Errorf("%s: Load Balancer %s not found", op, v)
		}
		loadBalancerID = n.ID
	}
	if loadBalancerID == 0 {
		klog.InfoS("%s: %s empty", op, hcloudLoadBalancerENVVar)
	}

	return &LoadBalancer{
		loadBalancerID: loadBalancerID,
		privateNetwork: os.Getenv(hcloudPrivateNetworksENVVar) != "",
	}, nil
}

func (l *LoadBalancer) Register(c *Cloud) error {
	const op = "hcloud-connect/registerLoadBalancer"

	server, _, err := c.client.Server.GetByName(context.Background(), c.nodeName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if server == nil {
		return fmt.Errorf("instance not found")
	}

	lb, _, err := c.client.LoadBalancer.GetByID(context.Background(), l.loadBalancerID)
	if err != nil {
		return err
	}

	opts := hcloud.LoadBalancerAddServerTargetOpts{
		Server:       server,
		UsePrivateIP: hcloud.Bool(l.privateNetwork),
	}

	_, _, err = c.client.LoadBalancer.AddServerTarget(context.Background(), lb, opts)
	if err != nil {
		return err
	}

	return nil
}

func (l *LoadBalancer) Deregister(c *Cloud) error {
	const op = "hcloud-connect/deregisterLoadBalancer"

	server, _, err := c.client.Server.GetByName(context.Background(), c.nodeName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if server == nil {
		return fmt.Errorf("instance not found")
	}

	lb, _, err := c.client.LoadBalancer.GetByID(context.Background(), l.loadBalancerID)
	if err != nil {
		return err
	}

	_, _, err = c.client.LoadBalancer.RemoveServerTarget(context.Background(), lb, server)
	if err != nil {
		return err
	}

	return nil
}
