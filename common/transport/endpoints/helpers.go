package endpoints

import (
	"context"
	redis "github.com/beezlabs-org/go-microservices-lib/common/cache"
	endpoints "github.com/beezlabs-org/go-microservices-lib/internal/transport/endpoints"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

func SetEndpoint(ctx context.Context, endpoint endpoint.Endpoint, name string, circuitBreakerSettings gobreaker.Settings,
	rateLimit int, rateLimitDuration time.Duration, isCacheable bool, ttl time.Duration, logger log.Logger,
	duration metrics.Histogram, tracer stdopentracing.Tracer, cache redis.RedisCache) endpoint.Endpoint {
	endpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(
		rate.Every(rateLimitDuration), rateLimit))(endpoint)
	endpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(circuitBreakerSettings))(endpoint)
	endpoint = opentracing.TraceEndpoint(tracer, name)(endpoint)
	endpoint = endpoints.InstrumentationMiddleware(duration.With("endpoint", name))(endpoint)
	if isCacheable && cache != nil {
		cache.SetCacheForEndpoint(ctx, name, ttl)
		endpoint = endpoints.RedisCacheMiddleware(cache, name, logger)(endpoint)
	}
	endpoint = AuthorizationMiddleware(name, logger)(endpoint)
	return endpoint
}
