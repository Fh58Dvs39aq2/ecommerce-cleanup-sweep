package infrai

import "testing"

func TestRetryDelay(t *testing.T) {
	if got := retryDelay("3", 0); got.Seconds() != 3 {
		t.Fatalf("Retry-After delay = %s, want 3s", got)
	}
	if got := retryDelay("", 2); got.Seconds() != 4 {
		t.Fatalf("exponential delay = %s, want 4s", got)
	}
}
