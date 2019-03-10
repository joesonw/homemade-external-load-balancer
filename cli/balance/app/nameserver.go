package app

import (
	"time"
	"log"
	"context"
)

func (c *Client) startSyncNameServer() error {
	c.doSyncNameServer()
	for range time.Tick(c.config.SyncPeriod) {
		c.doSyncNameServer()
	}
	return nil
}

func (c *Client) doSyncNameServer() {
	ip, err := c.ip.Get()
	if err != nil {
		log.Printf("unable to get ip: %s \n", err.Error())
	}
	ctx := context.TODO()
	err = c.nameserver.Set(ctx, int32(c.config.TTL.Seconds()), c.config.Domain, c.config.SubDomain, ip)
	if err != nil {
		log.Printf("unable to sync ip to nameserver: %s \n", err.Error())
	}
}
