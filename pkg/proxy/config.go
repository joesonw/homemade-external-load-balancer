package proxy

import "fmt"

type Config struct {
	Traefik *TraefikConfig `yaml:"traefik"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Traefik != nil {
		return NewTraefik(cfg.Traefik)
	}
	return nil, fmt.Errorf("unsupported proxy provider")
}

