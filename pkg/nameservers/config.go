package nameservers

import "fmt"

type Config struct {
	Dnspod  *DnspodConfig  `yaml:"dnspod"`
	Godaddy *GodaddyConfig `yaml:"godaddy"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Dnspod != nil {
		return NewDnspod(cfg.Dnspod)
	} else if cfg.Godaddy != nil {
		return NewGodday(cfg.Godaddy)
	}
	return nil, fmt.Errorf("usupported nameserver provider")
}
