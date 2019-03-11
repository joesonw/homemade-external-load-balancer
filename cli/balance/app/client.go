package app

import (
	"log"
	"pkg/ips"
	"pkg/nameservers"
	"pkg/proxy"

	"k8s.io/client-go/kubernetes"
)

type Client struct {
	client *kubernetes.Clientset
	config *Config

	serviceMapping map[string]*dnsMapping
	nameserver     nameservers.Interface
	ip             ips.Interface
	proxy          proxy.Interface
}

func New(cfg *Config, client *kubernetes.Clientset, nameserver nameservers.Interface, ip ips.Interface, p proxy.Interface) *Client {
	return &Client{
		config:         cfg,
		client:         client,
		nameserver:     nameserver,
		ip:             ip,
		proxy:          p,
		serviceMapping: make(map[string]*dnsMapping),
	}
}

func (c *Client) Start() error {
	go c.startSyncNameServer()
	go c.startWatchCluster()

	log.Printf("starting client\n")
	return c.startDNS()
}
