package nameservers

import "fmt"

type Config struct {
	Dnspod *DnspodConfig `yaml:"dnspod"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Dnspod != nil {
		return NewDnspod(cfg.Dnspod)
	}
	return nil, fmt.Errorf("usupported nameserver provider")
}
