package kinetovela

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveCore exercises a real KINETOVELA control plane. Set
// KINETOVELA_TEST_URL, e.g. http://127.0.0.1:1926, to run it.
func TestLiveCore(t *testing.T) {
	base := os.Getenv("KINETOVELA_TEST_URL")
	if base == "" {
		t.Skip("KINETOVELA_TEST_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	c, err := New(Config{BaseURL: base})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}
	if _, err := c.Version(ctx); err != nil {
		t.Fatalf("version: %v", err)
	}
	if _, err := c.FleetSLO(ctx); err != nil {
		t.Fatalf("fleet slo: %v", err)
	}
}
