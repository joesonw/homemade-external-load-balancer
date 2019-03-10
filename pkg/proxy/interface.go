package proxy

type Record struct {
	URL    string
	Host   string
	Port   int32
	Secure bool
}

type Interface interface {
	Refresh(records []*Record) error
}
