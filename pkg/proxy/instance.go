package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"net"

	"github.com/libp2p/go-reuseport"
)

type loadBalaceFunc func() string

type proxyTarget struct {
	host string
	port int32
}

type proxyInstance struct {
	port            int32
	targets         []*proxyTarget
	listener        net.Listener
	server          *http.Server
	loadBalanceFunc loadBalaceFunc
	tlsConfig       *tls.Config
}

func newProxyInstance(port int32, certs []tls.Certificate, loadBalanceFunc loadBalaceFunc) *proxyInstance {
	tlsConfig := &tls.Config{
		Certificates:       certs,
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate()
	return &proxyInstance{
		port:            port,
		tlsConfig:       tlsConfig,
		loadBalanceFunc: loadBalanceFunc,
	}
}

func (i *proxyInstance) addCert(cert *tls.Certificate) {
	i.tlsConfig.Certificates = append(i.tlsConfig.Certificates, *cert)
	i.tlsConfig.BuildNameToCertificate()
}

func (i *proxyInstance) addTarget(host string, port int32) {
	i.targets = append(i.targets, &proxyTarget{host, port})
}

func (i *proxyInstance) removeHost(host string) {
	var targets []*proxyTarget
	for _, target := range i.targets {
		if target.host != host {
			targets = append(targets, target)
		}
	}
	i.targets = targets
}

func (i *proxyInstance) len() int {
	return len(i.targets)
}

func (i *proxyInstance) start(secure bool, host string) error {
	listener, err := reuseport.Listen("tcp", fmt.Sprintf("%s:%d", host, i.port))
	if err != nil {
		return err
	}
	server := &http.Server{}
	if secure {
		listener = tls.NewListener(listener, i.tlsConfig)
		server.Handler = http.HandlerFunc(i.handleSecure)
	} else {
		server.Handler = http.HandlerFunc(i.handlePlain)
	}
	i.listener = listener
	i.server = server
	return server.Serve(listener)
}

func (i *proxyInstance) stop() error {
	if err := i.server.Close(); err != nil {
		return err
	}
	return i.listener.Close()
}

func (i *proxyInstance) handlePlain(res http.ResponseWriter, req *http.Request) {
	i.handle(false, res, req)
}

func (i *proxyInstance) handleSecure(res http.ResponseWriter, req *http.Request) {
	i.handle(true, res, req)
}

func (i *proxyInstance) handle(secure bool, res http.ResponseWriter, req *http.Request) {
	var target *proxyTarget
	for _, t := range i.targets {
		if t.host == req.Host {
			target = t
			break
		}
	}

	if target == nil {
		http.Error(res, "service not found", http.StatusBadGateway)
		return
	}

	scheme := "http"
	if secure {
		scheme = "https"
	}

	host := i.loadBalanceFunc()
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", scheme, host, target.port))

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	req.URL.Host = u.Host
	req.URL.Scheme = u.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = u.Host
	proxy.ServeHTTP(res, req)

}
