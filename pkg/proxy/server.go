package proxy

import (
	"sync"

	"crypto/tls"

	"k8s.io/client-go/kubernetes"
)

type Server struct {
	clientset *kubernetes.Clientset
	host      string

	proxies     map[int32]*proxyInstance
	recordLock  *sync.RWMutex
	defaultCert *tls.Certificate

	nodes    map[string]string
	nodeLock *sync.RWMutex
}

func New(clientset *kubernetes.Clientset, defaultCert *tls.Certificate, host string) *Server {
	s := &Server{
		clientset:   clientset,
		host:        host,
		proxies:     make(map[int32]*proxyInstance),
		recordLock:  &sync.RWMutex{},
		defaultCert: defaultCert,
		nodes:       make(map[string]string),
		nodeLock:    &sync.RWMutex{},
	}
	return s
}

func (s *Server) Start() error {
	go s.startWatchNodes()

	return nil
}
