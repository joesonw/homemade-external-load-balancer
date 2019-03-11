package ips

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestIpifyGet(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.ipify.org").
		Get("/").
		Reply(200).
		JSON(map[string]string{"ip": "1.2.3.4"})

	ipify, err := NewIpify(&IpifyConfig{URL: "https://api.ipify.org?format=json"})
	assert.Nil(t, err)

	ip, err := ipify.Get()
	assert.Nil(t, err)
	assert.Equal(t, ip, "1.2.3.4")
}
