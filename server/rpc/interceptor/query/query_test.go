package query_test

import (
	"context"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/query"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func mockUnaryHandler(ctx context.Context, r interface{}) (interface{}, error) {
	return ctx.Value(panacea.ContextKeyQueryBlockHeight{}), nil
}

func mockStreamHandler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	lastBlockHeight := int64(1000)
	queryInterceptor := query.NewQueryInterceptor(mocks.MockQueryClient{
		LastBlockHeight: lastBlockHeight,
	})

	ctx := context.Background()
	req := "value"
	res, err := queryInterceptor.UnaryServerInterceptor()(
		ctx,
		req,
		&grpc.UnaryServerInfo{},
		mockUnaryHandler,
	)

	require.NoError(t, err)
	require.Equal(t, lastBlockHeight-1, res)
}

func TestStreamServerInterceptor(t *testing.T) {
	lastBlockHeight := int64(1000)
	queryInterceptor := query.NewQueryInterceptor(mocks.MockQueryClient{
		LastBlockHeight: lastBlockHeight,
	})

	ctx := context.Background()
	req := "value"

	serverStream := &grpc_middleware.WrappedServerStream{
		ServerStream:   nil,
		WrappedContext: ctx,
	}

	err := queryInterceptor.StreamServerInterceptor()(
		req,
		serverStream,
		nil,
		mockStreamHandler,
	)

	queryHeight := serverStream.Context().Value(panacea.ContextKeyQueryBlockHeight{})

	require.NoError(t, err)
	require.Equal(t, lastBlockHeight-1, queryHeight)
}
