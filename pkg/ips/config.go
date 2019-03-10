package ips

import "fmt"

type Config struct {
	Static *StaticConfig `yaml:"static"`
	Ipify  *IpifyConfig  `yaml:"ipify"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Static != nil {
		return NewStatic(cfg.Static)
	} else if cfg.Ipify != nil {
		return NewIpify(cfg.Ipify)
	}
	return nil, fmt.Errorf("unsupported ip provider")
}
