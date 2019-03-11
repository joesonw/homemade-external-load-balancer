package app

import (
	"fmt"
	"log"
	"pkg/proxy"
)

func (c *Client) refreshProxy() {
	var records []*proxy.Record
	for name, svc := range c.serviceMapping {
		records = append(records, &proxy.Record{
			URL:  fmt.Sprintf("%s.%s.%s", name, c.config.SubDomain, c.config.Domain),
			Host: svc.DNSName,
			Port: svc.Port,
		})
	}
	err := c.proxy.Refresh(records)
	if err != nil {
		log.Printf("uable to refresh proxy: %s", err.Error())
	}
}
