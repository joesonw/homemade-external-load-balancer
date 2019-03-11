package proxy

type Record struct {
	URL  string
	Host string
	Port int32
}

type Interface interface {
	Refresh(records []*Record) error
}
