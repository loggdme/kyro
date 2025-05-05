package limiter_test

import (
	"testing"
	"time"

	"github.com/loggdme/kyro/pkg/limiter"
)

func TestRateLimiter_Wait(t *testing.T) {
	rl := limiter.NewRateLimiter(2, 2)

	// The first call should not block (due to burst)
	start := time.Now()
	if err := rl.Wait(); err != nil {
		t.Fatalf("Wait failed on first call: %v", err)
	}
	duration := time.Since(start)
	if duration > 10*time.Millisecond {
		t.Errorf("First Wait took too long: %v", duration)
	}

	// The second call should also not block immediately (due to burst)
	start = time.Now()
	if err := rl.Wait(); err != nil {
		t.Fatalf("Wait failed on second call: %v", err)
	}
	duration = time.Since(start)
	if duration > 10*time.Millisecond {
		t.Errorf("Second Wait took too long: %v", duration)
	}

	// The third call should block because the burst is used up
	// and we exceed the rate of 2 events per second
	start = time.Now()
	if err := rl.Wait(); err != nil {
		t.Fatalf("Wait failed on third call: %v", err)
	}
	duration = time.Since(start)
	expectedMinDelay := 450 * time.Millisecond
	if duration < expectedMinDelay {
		t.Errorf("Third Wait did not block long enough. Expected at least %v, got %v", expectedMinDelay, duration)
	}
}
