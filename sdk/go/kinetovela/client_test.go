package kinetovela

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/healthz":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/version":
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.2.3"})
		case "/fleet/slo":
			_ = json.NewEncoder(w).Encode(FleetMetrics{FleetSize: 3, AvailableRobots: 2})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := c.Health(ctx); err != nil {
		t.Fatal(err)
	}
	v, err := c.Version(ctx)
	if err != nil || v != "1.2.3" {
		t.Fatalf("version = %q, err = %v", v, err)
	}
	slo, err := c.FleetSLO(ctx)
	if err != nil || slo.FleetSize != 3 {
		t.Fatalf("slo = %+v, err = %v", slo, err)
	}
}

func TestEdgeFleetCalls(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost && r.URL.Path == "/edges" {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := c.RegisterEdge(ctx, EdgeRegistration{ID: "e1", TenantID: "t1"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Heartbeat(ctx, "e1", 5); err != nil {
		t.Fatal(err)
	}
	if err := c.BindMission(ctx, "e1", "m1", "i1"); err != nil {
		t.Fatal(err)
	}
	if err := c.UnbindMission(ctx, "e1", "m1"); err != nil {
		t.Fatal(err)
	}
	if err := c.ReportSync(ctx, "e1", 2, 9); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"POST /edges", "POST /edges/e1/heartbeat", "POST /edges/e1/missions",
		"DELETE /edges/e1/missions/m1", "POST /edges/e1/sync",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("paths = %v", paths)
		}
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error for empty BaseURL")
	}
	if _, err := New(Config{BaseURL: "ftp://x"}); err == nil {
		t.Fatal("expected error for non-http URL")
	}
}
