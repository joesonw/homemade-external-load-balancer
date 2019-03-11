package ips

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticGet(t *testing.T) {
	static, err := NewStatic(&StaticConfig{IP: "1.2.3.4"})
	assert.Nil(t, err)
	ip, err := static.Get()
	assert.Nil(t, err)
	assert.Equal(t, ip, "1.2.3.4")
}
