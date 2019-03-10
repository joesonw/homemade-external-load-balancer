package ips

type StaticConfig struct {
	IP string `yaml:"ip"`
}

type Static struct {
	ip string
}

func NewStatic(cfg *StaticConfig) (*Static, error) {
	return &Static{ip: cfg.IP}, nil
}

func (s *Static) Get() (string, error) {
	return s.ip, nil
}
