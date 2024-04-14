package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Zanda256/rate-limiter-go/app/services/rate-limiter/handlers"
	"github.com/Zanda256/rate-limiter-go/business/data/cache"
	v1 "github.com/Zanda256/rate-limiter-go/business/web/v1"
	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

var build = "develop"

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT ******")
		},
		// Info: func(ctx context.Context, r logger.Record) {
		// 	log.Info(ctx, "**  Important  **")
		// },
	}

	//traceIDFunc := func(ctx context.Context) string {
	//	return web.GetTraceID(ctx)
	//}
	traceIDFunc := func(ctx context.Context) string {
		//return web.GetTraceID(ctx)
		return "not_set_up_yet"
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "Rate-limited-API", traceIDFunc, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "msg", err)
		return
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)
	type (
		Version struct {
			Build string
			Desc  string
		}
		Web struct {
			ReadTimeout     time.Duration // `conf:"default:5s"`
			WriteTimeout    time.Duration // `conf:"default:10s"`
			IdleTimeout     time.Duration // `conf:"default:120s"`
			ShutdownTimeout time.Duration // `conf:"default:20s,mask"`
			APIHost         string        // `conf:"default:0.0.0.0:3000"`

		}
		RedisConf struct {
			URL string
		}
		RateLimitConf map[string]ratelimiter.Tier
	)

	// map[string]ratelimiter.Tier{
	// 	"basic": {
	// 		Algo:     ratelimiter.TokenBucket,
	// 		Period:   60,
	// 		Capacity: 5,
	// 	},
	// },

	cfg := struct {
		Version
		Web
		RedisConf
		RateLimitConf
	}{
		Version: Version{
			Build: build,
			Desc:  "My first attempt at replicating the project",
		},
		Web: Web{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 20 * time.Second,
			APIHost:         "0.0.0.0:3000",
		},
		RateLimitConf: func() RateLimitConf {
			rlCfg := RateLimitConf{}
			jsonStr := os.Getenv("TIER_CONFIG")
			err := json.Unmarshal([]byte(jsonStr), &rlCfg)
			if err != nil {
				panic(err)
			}
			return rlCfg
		}(),
		RedisConf: func() RedisConf {
			return RedisConf{
				URL: os.Getenv("REDIS_URL"),
			}
		}(),
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	redis := cache.NewRedisCache(cfg.RedisConf.URL)

	cfgMux := v1.APIMuxConfig{
		Tiers:    cfg.RateLimitConf,
		RedisKv:  redis,
		Build:    build,
		Shutdown: shutdown,
		Log:      log,
	}

	apiMux := v1.APIMux(cfgMux, handlers.Routes{})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
