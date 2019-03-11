package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns"
)

type dnsMapping struct {
	DNSName string
	Port    int32
}

func (c *Client) startDNS() error {
	dns.Handle(fmt.Sprintf("%s.%s.", c.config.SubDomain, c.config.Domain), c)
	dnsServer := &dns.Server{Addr: fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), Net: "udp"}
	return dnsServer.ListenAndServe()
}

func (c *Client) ServeDNS(res dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Compress = false

	if req.Opcode == dns.OpcodeQuery {
		c.handleDNSQuery(m)
	}
	if err := res.WriteMsg(m); err != nil {
		log.Printf("uanble to write dns response: %s", err.Error())
	}
}

func (c *Client) handleDNSQuery(m *dns.Msg) {
	suffix := fmt.Sprintf("%s.%s.", c.config.SubDomain, c.config.Domain)
	ip, err := c.ip.Get()
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
		serviceName := q.Name[0 : len(q.Name)-len(suffix)-1]
		dm := c.serviceMapping[serviceName]
		if dm == nil {
			continue
		}
		r, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
		if err == nil {
			m.Answer = append(m.Answer, r)
		}
	}
}
