package ips

type Interface interface {
	Get() (string, error)
}
