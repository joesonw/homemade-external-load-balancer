package app

import (
	"pkg/ips"
	"pkg/nameservers"
	"pkg/proxy"
	"time"
)

type Config struct {
	Host       string           `yaml:"host"`
	Port       int              `yaml:"port"`
	Domain     string           `yaml:"domain"`
	SubDomain  string           `yaml:"subDomain"`
	TTL        time.Duration    `yaml:"ttl"`
	SyncPeriod time.Duration    `yaml:"syncPeriod"`
	Providers  *ProvidersConfig `yaml:"providers"`
}

type ProvidersConfig struct {
	Nameserver *nameservers.Config `yaml:"nameserver"`
	IP         *ips.Config         `yaml:"ip"`
	Proxy      *proxy.Config       `yaml:"proxy"`
}
