package limit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rateLimitInterceptor struct {
	limiter *rate.Limiter
}

func NewRateLimitInterceptor(maxConnection int) *rateLimitInterceptor {
	return &rateLimitInterceptor{
		rate.NewLimiter(per(maxConnection, time.Second), maxConnection),
	}
}

func per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

func (ic *rateLimitInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !ic.limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "The maximum number of allowable connections has been exceeded. please retry later.")
		}
		return handler(ctx, req)
	}
}

func (ic *rateLimitInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !ic.limiter.Allow() {
			return status.Errorf(codes.ResourceExhausted, "The maximum number of allowable connections has been exceeded. please retry later.")
		}
		return handler(srv, stream)
	}
}
