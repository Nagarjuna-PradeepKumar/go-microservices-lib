package cache

import (
	"github.com/beezlabs-org/go-microservices-lib/internal/cache"
)

type RedisCache cache.RedisCache
type RedisEndpointContext cache.RedisEndpointContext

var New = cache.New
