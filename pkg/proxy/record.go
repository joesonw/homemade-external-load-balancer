package proxy

import (
	"pkg/annotations"
	"strings"

	"strconv"

	"log"

	"crypto/tls"

	corev1 "k8s.io/api/core/v1"
)

type serviceRecord struct {
	hosts    []string
	port     int32
	nodePort int32
	secure   bool
}

func (s *Server) AddService(svc *corev1.Service, hosts []string) (err error) {
	var securePorts []*corev1.ServicePort
	if str := strings.TrimSpace(svc.Annotations[annotations.SecurePorts]); str != "" {
		portSlice := strings.Split(str, ",")
		for _, s := range portSlice {
			var i int
			i, err = strconv.Atoi(strings.TrimSpace(s))
			if err == nil {
				for _, sp := range svc.Spec.Ports {
					if sp.Port == int32(i) {
						securePorts = append(securePorts, sp.DeepCopy())
						break
					}
				}
			} else {
				for _, sp := range svc.Spec.Ports {
					if sp.Name == s {
						securePorts = append(securePorts, sp.DeepCopy())
						break
					}
				}
			}
		}
	}

	var cert tls.Certificate
	if len(securePorts) > 0 {
		certParts := strings.Split(strings.TrimSpace(svc.Annotations[annotations.Cert]), "/")
		var sslCert *sslCert
		sslCert, err = s.fetchSSLSecret(strings.TrimSpace(certParts[1]), strings.TrimSpace(certParts[0]))
		if err != nil {
			return
		}
		cert, err = tls.X509KeyPair(sslCert.Cert, sslCert.Key)
		if err != nil {
			return
		}
	}

	var ports []*corev1.ServicePort
	for _, sp := range svc.Spec.Ports {
		found := false
		for _, ssp := range securePorts {
			if ssp.Name == sp.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}
		ports = append(ports, sp.DeepCopy())
	}

	var records []*serviceRecord

	for _, p := range ports {
		records = append(records, &serviceRecord{
			port:     p.Port,
			nodePort: p.NodePort,
			hosts:    hosts,
			secure:   false,
		})
	}
	for _, p := range securePorts {
		records = append(records, &serviceRecord{
			port:     p.Port,
			nodePort: p.NodePort,
			hosts:    hosts,
			secure:   true,
		})
	}

	for _, r := range records {
		s.recordLock.Lock()
		if s.proxies[r.port] == nil {
			instance := newProxyInstance(r.port, []tls.Certificate{*s.defaultCert}, s.getRandomNodeIP)
			log.Printf("starting proxy for %s on port: %d", strings.Join(hosts, ", "), r.port)
			go func(secure bool, host string) {
				if err := instance.start(secure, host); err != nil {
					log.Printf("unable to start proxy instace: %s", err.Error())
				}
			}(r.secure, s.host)
			s.proxies[r.port] = instance
		}
		for _, h := range hosts {
			s.proxies[r.port].addTarget(h, r.nodePort)
		}
		if r.secure {
			s.proxies[r.port].addCert(&cert)
		}
		s.recordLock.Unlock()
	}

	return nil
}

func (s *Server) RemoveService(host string) {
	s.recordLock.Lock()
	defer s.recordLock.Unlock()
	for port, p := range s.proxies {
		p.removeHost(host)
		if p.len() == 0 {
			if err := p.stop(); err != nil {
				log.Printf("unable to stop proxy: %s", err.Error())
			}
			delete(s.proxies, port)
		}
	}
}
