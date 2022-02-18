package cache

import (
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-redis/redis/v8"
	"os"
	"strconv"
	"time"
)

type RedisCache interface {
	Get(ctx context.Context, key string, destination *interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetCacheForEndpoint(ctx context.Context, name string, ttl time.Duration)
	CheckIfEndPointIsCacheableAndGetTTL(ctx context.Context, endpointName string) (bool, time.Duration, error)
	GetEndpointContext(ctx context.Context) RedisEndpointContext
	CacheEndpointResponse(ctx context.Context, key string, ttl time.Duration,
		response interface{}) error
}

type RedisEndpointContext struct {
	Cacheable    bool
	Key          string
	NotFromCache bool
}

type cacheStruct struct {
	client *redis.Client
	logger log.Logger
}

func NewClient(ctx context.Context,
	options *redis.Options,
	logger log.Logger) *redis.Client {
	client := redis.NewClient(options)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	_ = level.Info(logger).Log("msg", "redis connection established")
	return client
}

func CloseClient(client *redis.Client, logger log.Logger) {
	if client != nil {
		err := client.Close()
		if err != nil {
			_ = level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
		_ = level.Info(logger).Log("msg", "redis connection closed")
	}
}

func New(client *redis.Client, logger log.Logger) RedisCache {
	if client == nil {
		return nil
	}
	return cacheStruct{
		client: client,
		logger: log.With(logger, "prefix", "redis-cache"),
	}
}

func (r cacheStruct) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	_ = level.Info(r.logger).Log("function", "set", "key", key, "ttl", ttl)
	stringValue, err := json.Marshal(value)
	if err != nil {
		_ = level.Error(r.logger).Log("json marshal error", err)
		return err
	}
	return r.setStringValue(ctx, key, string(stringValue), ttl)
}
func (r cacheStruct) setStringValue(ctx context.Context, key string, value string, ttl time.Duration) error {
	_ = level.Info(r.logger).Log("function", "SetStringValue", "key", key, "ttl", ttl)
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		_ = level.Error(r.logger).Log("redis set error", err)
		return err
	}
	return nil
}

func (r cacheStruct) Get(ctx context.Context, key string, destination *interface{}) error {
	_ = level.Info(r.logger).Log("function", "get", "key", key)
	result, err := r.readFromRedis(ctx, key)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(result), destination)
	if err != nil {
		_ = level.Error(r.logger).Log("json unmarshal error", err)
		return err
	}
	return nil
}

func (r cacheStruct) readFromRedis(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			_ = level.Error(r.logger).Log("function", "readFromRedis", "key", key, "redis data not found", err)
		} else {
			_ = level.Error(r.logger).Log("function", "readFromRedis", "redis get error", err)
		}
	}
	return result, err
}

func (r cacheStruct) SetCacheForEndpoint(ctx context.Context, endpointName string, ttl time.Duration) {
	_ = level.Info(r.logger).Log("function", "initCacheForEndpoint", "endpointName",
		endpointName, "ttl", ttl)
	err := r.Set(ctx, "cacheable-endpoint-ttl:"+endpointName, ttl, 0)
	if err != nil {
		_ = r.logger.Log("msg", "SetCacheForEndpoint", "err", err)
		os.Exit(-1)
	}
}

func (r cacheStruct) CheckIfEndPointIsCacheableAndGetTTL(ctx context.Context, endpointName string) (bool, time.Duration, error) {
	_ = level.Info(r.logger).Log("function", "CheckIfEndPointIsCacheableAndGetTTL", "endpointName",
		endpointName)
	data, err := r.checkIfEndpointIsCacheable(ctx, endpointName)
	if err != nil {
		return false, 0, err
	}
	ttl, err := convertStringToTimeDuration(data)
	if err != nil {
		return false, 0, err
	}
	return true, ttl, nil
}

func (r cacheStruct) checkIfEndpointIsCacheable(ctx context.Context, endpointName string) (string, error) {
	return r.readFromRedis(ctx, "cacheable-endpoint-ttl:"+endpointName)
}

func convertStringToTimeDuration(data string) (time.Duration, error) {
	var ttl time.Duration
	i, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, err
	}
	ttl = time.Duration(i)
	return ttl, nil
}

func (r cacheStruct) GetEndpointContext(ctx context.Context) RedisEndpointContext {
	_ = level.Info(r.logger).Log("function", "GetEndpointContext")
	endpointContext := ctx.Value("cacheable-endpoint-context").(RedisEndpointContext)
	return endpointContext
}

func (r cacheStruct) CacheEndpointResponse(ctx context.Context, key string, ttl time.Duration, response interface{}) error {
	_ = level.Info(r.logger).Log("function", "CacheEndpointResponse")
	err := r.Set(ctx, key, response, ttl)
	if err != nil {
		_ = r.logger.Log("msg", "CacheEndpointResponse", "err", err)
		return err
	}
	return nil
}
