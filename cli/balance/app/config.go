package app

import (
	"pkg/ips"
	"pkg/nameservers"
	"time"
)

type Config struct {
	DNS        *DNSConfig        `yaml:"dns"`
	Proxy      *ProxyConfig      `yaml:"proxy"`
	Domain     string            `yaml:"domain"`
	SubDomain  string            `yaml:"subDomain"`
	TLS        *TLSConfig        `yaml:"tls"`
	TTL        time.Duration     `yaml:"ttl"`
	SyncPeriod time.Duration     `yaml:"syncPeriod"`
	Providers  *ProvidersConfig  `yaml:"providers"`
	Kubernetes *KubernetesConfig `yaml:"kubernetes"`
}

type DNSConfig struct {
	Host string `yaml:"host"`
	Port int32  `yaml:"port"`
}

type ProxyConfig struct {
	Host string `yaml:"host"`
}

type TLSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type ProvidersConfig struct {
	Nameserver *nameservers.Config `yaml:"nameserver"`
	IP         *ips.Config         `yaml:"ip"`
}

type KubernetesConfig struct {
	Host  string `yaml:"host"`
	CA    string `yaml:"ca"`
	Token string `yaml:"token"`
}
