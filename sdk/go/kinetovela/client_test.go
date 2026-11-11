package kinetovela

import (
	"context"
	"encoding/json"
	"errors"
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

func TestMissionSubmitGetListAssign(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath()+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/missions":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(MissionView{Status: "validated"})
		case r.Method == http.MethodGet && r.URL.Path == "/missions/mission:wh3/patrol-17":
			_ = json.NewEncoder(w).Encode(MissionView{Status: "assigned"})
		case r.Method == http.MethodGet && r.URL.Path == "/missions":
			_ = json.NewEncoder(w).Encode(MissionList{Missions: []MissionView{{Status: "validated"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/missions/mission:wh3/patrol-17/assignments":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(MissionView{Status: "assigned"})
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
	mission := Mission{SchemaVersion: "1.0.0", MissionID: "mission:wh3/patrol-17", PlanRef: "plan:wh3/patrol", CapabilityRefs: []string{"cap:warehouse.navigate@1.0.0"}, AuthorityRef: "grant:wh3-ops", SafetyEnvelopeRef: "safe:wh3", ExpectedEffects: []string{"robot:tb4-01.pose=target-17"}}
	if _, err := c.SubmitMission(ctx, mission); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetMission(ctx, "mission:wh3/patrol-17"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListMissions(ctx, MissionFilter{EdgeID: "edge-1", Status: "validated", Limit: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AssignMission(ctx, "mission:wh3/patrol-17", AssignmentRequest{EdgeID: "edge-1", RobotRef: "robot:tb4-01", LeaseRef: "lease:wh3/tb4-01/4", FencingToken: 42, AssignmentReason: "nearest"}); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"POST /missions?",
		"GET /missions/mission:wh3%2Fpatrol-17?",
		"GET /missions?edge_id=edge-1&limit=10&status=validated",
		"POST /missions/mission:wh3%2Fpatrol-17/assignments?",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v", requests)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestMissionErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "stale_fence", "message": "fence is stale"})
	}))
	defer srv.Close()
	c, _ := New(Config{BaseURL: srv.URL})
	_, err := c.GetMission(context.Background(), "mission:wh3/patrol-17")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusConflict || apiErr.Code != "stale_fence" || apiErr.Message != "fence is stale" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestEdgeErrorStillHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer srv.Close()
	c, _ := New(Config{BaseURL: srv.URL})
	err := c.Heartbeat(context.Background(), "e1", 1)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %v, want HTTPError", err)
	}
}
