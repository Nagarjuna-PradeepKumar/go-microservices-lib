package common

import (
	"context"
	"flag"
	"fmt"
	"github.com/beezlabs-org/go-microservices-lib/internal/cache"
	redisCache "github.com/beezlabs-org/go-microservices-lib/internal/cache"
	postgres "github.com/beezlabs-org/go-microservices-lib/internal/db"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-redis/redis/v8"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var CloseDB = postgres.CloseDB
var CloseRedisClient = redisCache.CloseClient

func OnServiceClose(logger log.Logger) {
	_ = level.Info(logger).Log("msg", "service terminating")
}

func startHttpServer(logger log.Logger, h http.Handler, httpAddr *string, errChannel chan error) {
	err := level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
	if err != nil {
		return
	}
	server := &http.Server{
		Addr:    *httpAddr,
		Handler: h,
	}
	errChannel <- server.ListenAndServe()
}

func waitForInterrupt(errChannel chan error) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errChannel <- fmt.Errorf("%s", <-c)
}

func initDurationMetrics() metrics.Histogram {
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "beezlabs",
			Subsystem: "service",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"endpoint", "success"})
	}
	return duration
}

func InitLogger(serviceName string) log.Logger {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger,
			"service", serviceName,
			"time:", log.DefaultTimestampUTC,
			"caller:", log.DefaultCaller,
		)
	}
	//Add the first log for the service
	err := level.Info(logger).Log("msg", "service started")
	if err != nil {
		os.Exit(-1)
	}
	return logger
}

func initRedisClient(ctx context.Context, logger log.Logger, enabled *bool) *redis.Client {
	//Create a redis client for cache
	if *enabled {
		redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			_ = logger.Log("error in redis db conversion", err)
			os.Exit(-1)
		}
		redisClient := cache.NewClient(ctx,
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT"),
			os.Getenv("REDIS_PASSWORD"),
			redisDb,
			logger)
		return redisClient
	} else {
		return nil
	}
}

func setDBConnectionString() string {
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("SSL_MODE"))
	return dbSource
}

func getCommandLineArguments() (*string, *bool) {
	var httpAddr = flag.String("http_port", ":8080", "HTTP listen port")
	var isRedisEnabled = flag.Bool("redis_cache", true, "is Redis Cache needed")
	flag.Parse()
	return httpAddr, isRedisEnabled
}
