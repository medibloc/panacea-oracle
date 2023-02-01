package limit_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/limit"
	"google.golang.org/grpc"
)

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	// wait 2 sec
	time.Sleep(time.Second * 2)

	return fmt.Sprintf("res.%v", req), nil
}

func TestRateLimitInterceptorSameRequestAndLimit(t *testing.T) {
	reqCnt := 10
	limitCnt := 10

	limitInterceptor := limit.NewRateLimitInterceptor(limitCnt)
	fn := limitInterceptor.UnaryServerInterceptor()

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{}

	wg := sync.WaitGroup{}
	results := make(map[int]error)
	for i := 0; i < reqCnt; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := fn(ctx, i, info, handler)
			fmt.Println(fmt.Printf("req: %d, res: %s, err: %v", i, res, err))
			results[i] = err
		}(i)
	}

	fmt.Println("wait when finish go routine")
	wg.Wait()
	fmt.Print(results)

}
