package query

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"server/rpc/interceptor/query/query.go
	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type queryInterceptor struct {
	panaceaQueryClient panacea.QueryClient
}

func NewQueryInterceptor(queryClient panacea.QueryClient) *queryInterceptor {
	return &queryInterceptor{queryClient}
}

func (ic *queryInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := setHeightInContext(ctx, ic.panaceaQueryClient)
		return handler(newCtx, req)
	}
}

func (ic *queryInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		newCtx := setHeightInContext(ctx, ic.panaceaQueryClient)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

func (ic *queryInterceptor) Interceptor(ctx context.Context) context.Context {
	log.Debug("Call query interceptor")
	return setHeightInContext(ctx, ic.panaceaQueryClient)
}

func setHeightInContext(ctx context.Context, queryClient panacea.QueryClient) context.Context {
	height, err := queryClient.GetLastBlockHeight(ctx)
	if err == nil {
		log.Debugf("Set the previous height of the last block height. LastHeight: %v, SetHeight: %v", height, height-1)
		return panacea.SetQueryBlockHeightToContext(ctx, height-1)
	}

	log.Warnf("failed to get last block height. %v", err)
	return ctx
}
