package limit

import (
	"context"
	"time"

	"github.com/medibloc/panacea-oracle/config"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rateLimitInterceptor struct {
	waitTimeout int64
	limiter     *rate.Limiter
}

func NewRateLimitInterceptor(cfg config.GRPCConfig) *rateLimitInterceptor {
	maxConnectionSize := cfg.RateLimitPerSecond
	return &rateLimitInterceptor{
		waitTimeout: cfg.WaitTimeout,
		limiter:     rate.NewLimiter(per(maxConnectionSize, time.Second), maxConnectionSize),
	}
}

func per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

func (ic *rateLimitInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := ic.Interceptor(); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (ic *rateLimitInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := ic.Interceptor(); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func (ic *rateLimitInterceptor) Interceptor() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ic.waitTimeout)*time.Second)
	defer cancel()

	if err := ic.limiter.Wait(ctx); err != nil {
		return status.Errorf(codes.ResourceExhausted, "failed with timeout while waiting for rate limiting. please retry later. %v", err)
	}

	return nil
}
