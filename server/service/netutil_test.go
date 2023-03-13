package service_test

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/netutil"
)

func TestNetUtil(t *testing.T) {

	lis := &fakeListener{timeWait: 1}

	limitLis := netutil.LimitListener(lis, 2)

	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			conn, err := limitLis.Accept()
			require.NoError(t, err)
			defer conn.Close()
		}(i)
	}

	wg.Wait()
	end := time.Now()

	// Send 10 requests, process 2 at a time, and take 1 second per request.
	// This request test should take 5 to 6 seconds.
	require.True(t, start.Add(time.Second*5).Before(end))
	require.True(t, start.Add(time.Second*6).After(end))

}

type fakeListener struct {
	timeWait int64
}

// Accept waits for and returns the next connection to the listener.
func (l *fakeListener) Accept() (net.Conn, error) {
	time.Sleep(time.Second * time.Duration(l.timeWait))

	return fakeNetConn{}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *fakeListener) Close() error {
	return nil
}

// Addr returns the listener's network address.
func (l *fakeListener) Addr() net.Addr {
	return fakeAddr(1)
}

type fakeNetConn struct {
	io.Reader
	io.Writer
}

func (c fakeNetConn) Close() error                       { return nil }
func (c fakeNetConn) LocalAddr() net.Addr                { return localAddr }
func (c fakeNetConn) RemoteAddr() net.Addr               { return remoteAddr }
func (c fakeNetConn) SetDeadline(t time.Time) error      { return nil }
func (c fakeNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (c fakeNetConn) SetWriteDeadline(t time.Time) error { return nil }

type fakeAddr int

var (
	localAddr  = fakeAddr(1)
	remoteAddr = fakeAddr(2)
)

func (a fakeAddr) Network() string {
	return "net"
}

func (a fakeAddr) String() string {
	return "str"
}
