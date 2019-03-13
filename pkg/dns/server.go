package dns

import (
	"fmt"
	"log"
	"strings"

	"pkg/ips"

	"sync"

	"github.com/miekg/dns"
)

type mapping struct {
	DNSName string
	Port    int32
}

type Server struct {
	domain    string
	subDomain string
	host      string
	port      int32
	mappings  map[string]*mapping
	lock      *sync.RWMutex
	ip        ips.Interface
}

func New(domain, subDomain, host string, port int32, ip ips.Interface) *Server {
	mappings := make(map[string]*mapping)
	lock := &sync.RWMutex{}
	s := &Server{
		domain,
		subDomain,
		host,
		port,
		mappings,
		lock,
		ip,
	}
	dns.Handle(fmt.Sprintf("%s.%s.", subDomain, domain), s)
	return s
}

func (s *Server) StartUDP() error {
	dns.Handle(fmt.Sprintf("%s.%s.", s.subDomain, s.domain), s)
	dnsServer := &dns.Server{Addr: fmt.Sprintf("%s:%d", s.host, s.port), Net: "udp"}
	return dnsServer.ListenAndServe()
}

func (s *Server) StartTCP() error {
	dnsServer := &dns.Server{Addr: fmt.Sprintf("%s:%d", s.host, s.port), Net: "tcp"}
	return dnsServer.ListenAndServe()
}

func (s *Server) ServeDNS(res dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Compress = false

	if req.Opcode == dns.OpcodeQuery {
		s.query(m)
	}
	if err := res.WriteMsg(m); err != nil {
		log.Printf("uanble to write dns response: %s", err.Error())
	}
}

func (s *Server) Add(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.mappings[name] = &mapping{}
}

func (s *Server) Remove(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.mappings, name)
}

func (s *Server) query(m *dns.Msg) {
	suffix := fmt.Sprintf("%s.%s.", s.subDomain, s.domain)
	ip, err := s.ip.Get()
	if err != nil {
		log.Printf("unable to get ip: %s", err.Error())
		return
	}
	for _, q := range m.Question {
		if q.Qtype != dns.TypeA {
			continue
		}
		if !strings.HasSuffix(q.Name, suffix) {
			continue
		}
		s.lock.RLock()
		dm := s.mappings[q.Name[0:len(q.Name)-1]]
		s.lock.RUnlock()
		if dm == nil {
			continue
		}
		r, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
		if err == nil {
			m.Answer = append(m.Answer, r)
		}
	}
}
