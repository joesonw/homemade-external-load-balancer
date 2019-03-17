package nameservers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type godaddyRecord struct {
	Data string `json:"data,omitempty"`
	Name string `json:"name,omitempty"`
	TTL  int32  `json:"ttl,omitempty"`
	Type string `json:"type,omitempty"`
}

type GodaddyConfig struct {
	URL    string `yaml:"url"`
	Key    string `yaml:"key"`
	Secret string `yaml:"secret"`
	Prefix string `yaml:"prefix"`
}

type Godaddy struct {
	config *GodaddyConfig
	client *http.Client
	url    *url.URL
}

func NewGodday(cfg *GodaddyConfig) (Interface, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	return &Godaddy{
		config: cfg,
		client: &http.Client{},
		url:    u,
	}, nil
}

func (in *Godaddy) Set(ctx context.Context, ttl int32, domain, name, ip string) error {
	if err := in.set(ctx, ttl, domain, fmt.Sprintf("%s1", name), ip); err != nil {
		return err
	}
	return in.set(ctx, ttl, domain, fmt.Sprintf("%s2", name), ip)
}

func (in *Godaddy) set(ctx context.Context, ttl int32, domain, name, ip string) error {
	var current *godaddyRecord
	p := fmt.Sprintf("/v1/domains/%s/records/A/%s%s", domain, in.config.Prefix, name)
	{
		u := url.URL{
			Scheme: in.url.Scheme,
			Host:   in.url.Host,
			Path:   p,
		}
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", in.config.Key, in.config.Secret))
		res, err := in.client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("unable to get godaddy record: %s", string(resBytes))
		}

		var result []*godaddyRecord
		err = json.Unmarshal(resBytes, &result)
		if err != nil {
			return err
		}
		current = result[0]
	}
	if strings.EqualFold(current.Data, ip) {
		return nil
	}

	{
		current.Data = ip
		current.TTL = ttl
		bodyBytes, err := json.Marshal(&[]*godaddyRecord{current})
		if err != nil {
			return err
		}

		u := url.URL{
			Scheme: in.url.Scheme,
			Host:   in.url.Host,
			Path:   p,
		}
		req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", in.config.Key, in.config.Secret))
		req.Header.Set("Content-Type", "application/json")
		defer req.Body.Close()
		res, err := in.client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}

		defer res.Body.Close()
		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return errors.New(string(resBytes))
		}
	}

	return nil
}
