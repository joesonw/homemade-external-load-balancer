package nameservers

import "fmt"

type Config struct {
	Dnspod     *DnspodConfig     `yaml:"dnspod"`
	CloudFlare *CloudFlareConfig `yaml:"cloudflare"`
}

func New(cfg *Config) (Interface, error) {
	if cfg.Dnspod != nil {
		return NewDnspod(cfg.Dnspod)
	} else if cfg.CloudFlare != nil {
		return NewCloudFlare(cfg.CloudFlare)
	}
	return nil, fmt.Errorf("usupported nameserver provider")
}
