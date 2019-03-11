package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type traefikRequest struct {
	Frontends map[string]*traefikRequestFrontend `json:"frontends,omitempty"`
	Backends  map[string]*traefikRequestBackend  `json:"backends,omitempty"`
}

type traefikRequestFrontend struct {
	PassTLSCert bool                                    `json:"passTLSCert,omitempty"`
	Routes      map[string]*traefikRequestFrontendRoute `json:"routes,omitempty"`
	Backend     string                                  `json:"backend,omitempty"`
}

type traefikRequestFrontendRoute struct {
	Rule string `json:"rule,omitempty"`
}

type traefikRequestBackend struct {
	Servers *traefikRequestBackendServer `json:"servers,omitempty"`
}

type traefikRequestBackendServer struct {
	Service *traefikRequestBackendServerService `json:"service,omitempty"`
}

type traefikRequestBackendServerService struct {
	URL string `json:"url,omitempty"`
}

type TraefikConfig struct {
	URL string `yaml:"url"`
}

type Traefik struct {
	url    string
	client *http.Client
}

func NewTraefik(cfg *TraefikConfig) (Interface, error) {
	return &Traefik{
		url:    cfg.URL,
		client: &http.Client{},
	}, nil
}

func (in *Traefik) Refresh(records []*Record) error {
	config := traefikRequest{
		Frontends: make(map[string]*traefikRequestFrontend),
		Backends:  make(map[string]*traefikRequestBackend),
	}
	for _, record := range records {

		config.Frontends[record.URL] = &traefikRequestFrontend{
			Backend: record.URL,
			Routes: map[string]*traefikRequestFrontendRoute{
				record.URL: {
					Rule: fmt.Sprintf("Host:%s", record.URL),
				},
			},
		}
		p := "http"
		if record.Secure {
			p = "https"
			config.Frontends[record.URL].PassTLSCert = true
		}
		config.Backends[record.URL] = &traefikRequestBackend{
			Servers: &traefikRequestBackendServer{
				Service: &traefikRequestBackendServerService{
					URL: fmt.Sprintf("%s://%s:%d", p, record.Host, record.Port),
				},
			},
		}
	}

	body, err := json.Marshal(&config)
	if err != nil {
		return err
	}
	println(string(body))

	req, err := http.NewRequest(http.MethodPut, in.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer req.Body.Close()

	res, err := in.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
