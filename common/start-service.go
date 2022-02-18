package common

import (
	"context"
	"database/sql"
	postgres "github.com/beezlabs-org/go-microservices-lib/internal/db"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	stdopentracing "github.com/opentracing/opentracing-go"
	"net/http"
)

func InitServiceComponents(serviceName string) (*string, log.Logger, stdopentracing.Tracer, metrics.Histogram,
	context.Context, *sql.DB, *redis.Client) {
	//Load command line flags
	httpAddr, isRedisEnabled := getCommandLineArguments()
	//Create a logger
	logger := InitLogger(serviceName)
	//Add OpenTracing tacker
	tracer := stdopentracing.GlobalTracer()
	//Create sparse metrics
	duration := initDurationMetrics()
	//Create a context
	ctx := context.Background()
	//Load environment variables for the Database connection string
	dbSource := setDBConnectionString()
	//Create a Postgres DB connection
	db := postgres.InitDB(dbSource, logger, serviceName)
	//Init Redis client
	redisClient := initRedisClient(ctx, logger, isRedisEnabled)
	return httpAddr, logger, tracer, duration, ctx, db, redisClient
}

func StartHTTPServer(h http.Handler, logger log.Logger, httpAddr *string) {
	//Channel to listen for service exit
	errChannel := make(chan error)
	go waitForInterrupt(errChannel)
	//Start the HTTP server
	go startHttpServer(logger, h, httpAddr, errChannel)
	//Print the error on service exit
	_ = level.Error(logger).Log("exit", <-errChannel)
}
