package endpoints

import (
	"context"
	"fmt"
	redis "github.com/beezlabs-org/go-microservices-lib/internal/cache"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"time"
)

func InstrumentationMiddleware(duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func RedisCacheMiddleware(redis redis.RedisCache, endpointName string, logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			isCacheable, ttl, err := redis.CheckIfEndPointIsCacheableAndGetTTL(ctx, endpointName)
			if err != nil || !isCacheable {
				_ = level.Info(logger).Log("function", "RedisCacheMiddleware", "err", err)
			} else {
				if isCacheable {
					endpointContext := redis.GetEndpointContext(ctx)
					if endpointContext.Cacheable {
						key := endpointName + ":" + endpointContext.Key
						//Check if the cache has the required data
						if !endpointContext.NotFromCache {
							response, err := getResponseFromRedisCache(ctx, redis, key, response, logger)
							if err == nil {
								return response, nil
							}
						}
						//If the cache does not have the required data, get the response from domain services and
						//store the response in the cache
						return getResponseFromDomainAndCacheIt(ctx, request, next, redis, key, ttl, logger)
					}
				}
			}
			_ = level.Info(logger).Log("function", "RedisCacheMiddleware", "data-from-cache", "false")
			return next(ctx, request)
		}
	}
}

func getResponseFromRedisCache(ctx context.Context, redis redis.RedisCache, key string, response interface{},
	logger log.Logger) (interface{}, error) {
	err := redis.Get(ctx, key, &response)
	if err == nil {
		_ = level.Info(logger).Log("function", "RedisCacheMiddleware", "key", key, "data-from-cache", "true")
		return response, nil
	}
	return nil, err
}

func getResponseFromDomainAndCacheIt(ctx context.Context, request interface{}, next endpoint.Endpoint,
	redis redis.RedisCache, key string, ttl time.Duration, logger log.Logger) (interface{}, error) {
	//Get the response from the Domain/Business Layer
	returnedResponse, err := next(ctx, request)
	//Store the response in the cache
	if err == nil {
		_ = redis.CacheEndpointResponse(ctx, key, ttl, returnedResponse)
	}
	_ = level.Info(logger).Log("function", "RedisCacheMiddleware", "key", key, "data-from-cache", "false")
	return returnedResponse, err
}
