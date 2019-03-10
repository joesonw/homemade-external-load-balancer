package ips

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type IpifyConfig struct {
	URL string `yaml:"url"`
}

type Ipify struct {
	client *http.Client
	url    string
}

func NewIpify(cfg *IpifyConfig) (Interface, error) {
	return &Ipify{client: &http.Client{}, url: cfg.URL}, nil
}

func (s *Ipify) Get() (string, error) {
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return "", err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	result := make(map[string]string)
	err = json.Unmarshal(bytes, &result)
	return result["ip"], err
}
