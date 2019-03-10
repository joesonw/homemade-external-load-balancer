package nameservers

import (
	"context"
)

type Interface interface {
	Set(ctx context.Context, ttl int32, domain, name, ip string) error
}
