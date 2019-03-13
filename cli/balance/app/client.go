package app

import (
	"log"
	"pkg/ips"
	"pkg/nameservers"
	"pkg/proxy"

	"pkg/dns"

	"sync"

	"k8s.io/client-go/kubernetes"
)

type Client struct {
	client *kubernetes.Clientset
	config *Config

	serviceCache     map[string]string
	serviceCacheLock *sync.RWMutex

	nameserver nameservers.Interface
	ip         ips.Interface
	proxy      *proxy.Server
	dns        *dns.Server
}

func New(cfg *Config, client *kubernetes.Clientset, nameserver nameservers.Interface, dnsServer *dns.Server, ip ips.Interface, proxyServer *proxy.Server) *Client {
	return &Client{
		config:           cfg,
		client:           client,
		nameserver:       nameserver,
		ip:               ip,
		proxy:            proxyServer,
		dns:              dnsServer,
		serviceCache:     make(map[string]string),
		serviceCacheLock: &sync.RWMutex{},
	}
}

func (c *Client) Start() error {
	log.Printf("starting client\n")

	go c.startSyncNameServer()
	go c.startWatchCluster()

	return nil
}
