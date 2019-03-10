package nameservers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type recordListResponse struct {
	Status  *responseStatus             `json:"status,omitempty"`
	Records []*recordListResponseRecord `json:"records,omitempty"`
}

type recordModifyResponse struct {
	Status *responseStatus `json:"status,omitempty"`
}

type responseStatus struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type recordListResponseRecord struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
	Line  string `json:"line,omitempty"`
}

type DnspodConfig struct {
	Token string `yaml:"token"`
}

type Dnspod struct {
	config *DnspodConfig
	client *http.Client
}

func NewDnspod(cfg *DnspodConfig) (Interface, error) {
	return &Dnspod{
		config: cfg,
		client: &http.Client{},
	}, nil
}

func (in *Dnspod) Set(ctx context.Context, ttl int32, domain, name, ip string) error {
	var current *recordListResponseRecord
	{
		form := url.Values{}
		form.Set("login_token", in.config.Token)
		form.Set("domain", domain)
		form.Set("sub_domain", name)
		form.Set("format", "json")
		form.Set("record_type", "NS")
		req, err := http.NewRequest(http.MethodPost, "https://dnsapi.cn/Record.List", strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("UserAgent", "home-made-external-load-balancer/0.1.0 (https://github.com/joesonw/homemade-external-load-balancer)")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		defer req.Body.Close()
		res, err := in.client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer res.Body.Close()
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		result := recordListResponse{}
		err = json.Unmarshal(bytes, &result)
		if result.Status == nil {
			return fmt.Errorf("unkown http error")
		}
		if result.Status.Code != "1" {
			return fmt.Errorf(result.Status.Message)
		}
		if len(result.Records) == 0 {
			return fmt.Errorf("no current records were set")
		}
		current = result.Records[0]
	}
	if strings.EqualFold(current.Value, ip+".") {
		return nil
	}
	{
		form := url.Values{}
		form.Set("login_token", in.config.Token)
		form.Set("domain", domain)
		form.Set("sub_domain", name)
		form.Set("format", "json")
		form.Set("record_type", "NS")
		form.Set("record_id", current.ID)
		form.Set("record_line", current.Line)
		form.Set("value", ip)
		form.Set("ttl", fmt.Sprintf("%d", ttl))
		req, err := http.NewRequest(http.MethodPost, "https://dnsapi.cn/Record.Modify", strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("UserAgent", "home-made-external-load-balancer/0.1.0 (https://github.com/joesonw/homemade-external-load-balancer)")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		defer req.Body.Close()
		res, err := in.client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer res.Body.Close()
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		result := recordModifyResponse{}
		err = json.Unmarshal(bytes, &result)
		if result.Status == nil {
			return fmt.Errorf("unkown http error")
		}
		if result.Status.Code != "1" {
			return fmt.Errorf(result.Status.Message)
		}
	}

	return nil
}
