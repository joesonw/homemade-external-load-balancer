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

type cloudFlareList struct {
	Success bool                `json:"success,omitempty"`
	Errors  []string            `json:"errors,omitempty"`
	Result  []*cloudFlareRecord `json:"result,omitempty"`
}

type cloudFlareRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	TTL     int32  `json:"ttl,omitempty"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
}

type CloudFlareConfig struct {
	URL   string `yaml:"url"`
	Key   string `yaml:"key"`
	Zone  string `yaml:"zone"`
	Email string `yaml:"email"`
}

type CloudFlare struct {
	config *CloudFlareConfig
	client *http.Client
	url    *url.URL
}

func NewCloudFlare(cfg *CloudFlareConfig) (Interface, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	return &CloudFlare{
		config: cfg,
		client: &http.Client{},
		url:    u,
	}, nil
}

// https://api.cloudflare.com/client/v4/zones/fce7881b3a77aaa04d406459e121163c/dns_records

func (in *CloudFlare) Set(ctx context.Context, ttl int32, domain, name, ip string) error {
	xip := fmt.Sprintf("%s.xip.io", ip)
	var current *cloudFlareRecord
	p := fmt.Sprintf("/client/v4/zones/%s/dns_records", in.config.Zone)
	{
		u := url.URL{
			Scheme: in.url.Scheme,
			Host:   in.url.Host,
			Path:   p,
		}
		u.Query().Set("type", "NS")
		u.Query().Set("name", fmt.Sprintf("%s.%s", name, domain))
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Email", in.config.Email)
		req.Header.Set("X-Auth-Key", in.config.Key)
		res, err := in.client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		result := &cloudFlareList{}
		err = json.Unmarshal(resBytes, &result)
		if err != nil {
			return err
		}

		if !result.Success {
			return fmt.Errorf(strings.Join(result.Errors, ": "))
		}

		current = result.Result[0]
	}
	if strings.EqualFold(current.Content, xip) {
		return nil
	}

	{
		current.Content = xip
		current.TTL = ttl
		bodyBytes, err := json.Marshal(current)
		if err != nil {
			return err
		}

		u := url.URL{
			Scheme: in.url.Scheme,
			Host:   in.url.Host,
			Path:   p + "/" + current.ID,
		}
		req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Email", in.config.Email)
		req.Header.Set("X-Auth-Key", in.config.Key)
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
