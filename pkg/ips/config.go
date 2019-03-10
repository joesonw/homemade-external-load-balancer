package ips

import "fmt"

type Config struct {
	Static *StaticConfig `yaml:"static"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Static != nil {
		return NewStatic(cfg.Static)
	}
	return nil, fmt.Errorf("unsupported ip provider")
}
