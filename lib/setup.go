package lib

import (
	redisCache "github.com/beezlabs-org/go-microservices-lib/internal/cache"
	postgres "github.com/beezlabs-org/go-microservices-lib/internal/db"
)

var CloseDB = postgres.CloseDB
var CloseRedisClient = redisCache.CloseClient
var InitDB = postgres.InitDB
