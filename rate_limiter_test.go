package grab

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// testRateLimiter is a naive rate limiter that limits throughput to r tokens
// per second. The total number of tokens issued is tracked as n.
type testRateLimiter struct {
	r, n int
}

func (c *testRateLimiter) WaitN(ctx context.Context, n int) (err error) {
	c.n += n
	time.Sleep(
		time.Duration(1.00 / float64(c.r) * float64(n) * float64(time.Second)))
	return
}

func TestRateLimiter(t *testing.T) {
	// download a 128 byte file, 8 bytes at a time, with a naive 512bps limiter
	// should take > 250ms
	filesize := 128
	filename := ".testRateLimiter"
	defer os.Remove(filename)

	req, err := NewRequest(filename, fmt.Sprintf("%s?size=%d", ts.URL, filesize))
	if err != nil {
		t.Fatal(err)
	}

	// ensure multiple trips to the rate limiter by downloading 8 bytes at a time
	req.BufferSize = 8

	// limit to 512bps
	lim := &testRateLimiter{r: 512}
	req.RateLimiter = lim

	resp := DefaultClient.Do(req)
	if err = resp.Err(); err != nil {
		t.Fatal(err)
	}
	testComplete(t, resp)
	if lim.n != filesize {
		t.Errorf("expected %d bytes to pass through limiter, got %d", filesize, lim.n)
	}
	if resp.Duration().Seconds() < 0.25 {
		// BUG: this test can pass if the transfer was slow for unrelated reasons
		t.Errorf("expected transfer to take >250ms, took %v", resp.Duration())
	}
}

func ExampleRateLimiter() {
	req, _ := NewRequest("", "http://www.golang-book.com/public/pdf/gobook.pdf")

	// Attach a rate limiter, using the token bucket implementation from
	// golang.org/x/time/rate. Limit to 1Mbps with burst up to 2Mbps.
	req.RateLimiter = rate.NewLimiter(1048576, 2097152)

	resp := DefaultClient.Do(req)
	if err := resp.Err(); err != nil {
		log.Fatal(err)
	}
}
