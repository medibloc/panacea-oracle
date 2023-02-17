package limit

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/medibloc/panacea-oracle/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type results struct {
	sync.Mutex
	sync.WaitGroup
	results []result
}

func (r *results) waitAndPrint() {
	r.Wait()

	sort.SliceStable(r.results, func(i, j int) bool {
		return r.results[i].idx < r.results[j].idx
	})

	for _, res := range r.results {
		log.Infof("idx(%d), req(%s), res(%s), err(%v)",
			res.idx,
			res.req,
			res.res,
			res.err,
		)
	}

}

type result struct {
	idx int
	req string
	res string
	err error
}

func TestRateLimitInterceptorSameRequestAndLimit(t *testing.T) {
	reqCnt := 10
	limitCnt := 10
	waitTimeout := time.Second * 1

	results := handling(reqCnt, limitCnt, waitTimeout)
	results.waitAndPrint()

	for _, res := range results.results {
		require.NoError(t, res.err)
	}
}

func TestRateLimitInterceptorMoreRequestsThanLimit(t *testing.T) {
	reqCnt := 30
	limitCnt := 10
	waitTimeout := time.Second * 1

	results := handling(reqCnt, limitCnt, waitTimeout)
	results.waitAndPrint()

	errCnt := 0
	for _, res := range results.results {
		if res.err != nil {
			errCnt++
			require.ErrorContains(t, res.err, "failed with timeout while waiting for rate limiting. please retry later.")
		}
	}

	// 30 request
	// 10 handling, 20 wait
	// after 1 sec, 10 handling, 10 wait and timeout
	// In the end, 20 successes and 10 failures
	require.Equal(t, 10, errCnt)
}

func TestRateLimitInterceptorRequestPerSecondSameTheLimit(t *testing.T) {
	reqCnt := 10
	limitCnt := 10
	waitTimeout := time.Second * 1

	results := handling(reqCnt, limitCnt, waitTimeout)

	time.Sleep(time.Second + time.Millisecond*100)

	results2 := handling(reqCnt, limitCnt, waitTimeout)

	results.waitAndPrint()
	results2.waitAndPrint()

	for _, res := range results.results {
		require.NoError(t, res.err)
	}

	for _, res := range results2.results {
		require.NoError(t, res.err)
	}
}

func handling(reqCnt, maxConnSize int, waitTimeout time.Duration) *results {
	cfg := config.GRPCConfig{
		RateLimits:           maxConnSize,
		RateLimitWaitTimeout: waitTimeout,
	}
	limitInterceptor := NewRateLimitInterceptor(cfg)

	results := &results{results: []result{}}
	for i := 0; i < reqCnt; i++ {
		results.Add(1)
		go func(i int) {
			defer results.Done()

			reqTime := time.Now()
			err := limitInterceptor.Interceptor()
			resTime := time.Now()

			res := result{
				idx: i,
				req: reqTime.Format(time.RFC3339),
				res: resTime.Format(time.RFC3339),
				err: err,
			}

			results.Lock()
			defer results.Unlock()
			results.results = append(results.results, res)
		}(i)
	}

	return results
}
