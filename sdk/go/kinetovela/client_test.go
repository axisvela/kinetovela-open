package kinetovela

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	bodies := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath()+"?"+r.URL.RawQuery)
		raw, _ := io.ReadAll(r.Body)
		bodies[r.Method+" "+r.URL.Path] = string(raw)
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

	wantSubmitBody := `{"schema_version":"1.0.0","mission_id":"mission:wh3/patrol-17","plan_ref":"plan:wh3/patrol","capability_refs":["cap:warehouse.navigate@1.0.0"],"authority_ref":"grant:wh3-ops","safety_envelope_ref":"safe:wh3","expected_effects":["robot:tb4-01.pose=target-17"]}`
	if got := bodies["POST /missions"]; got != wantSubmitBody {
		t.Fatalf("submit body = %q, want %q", got, wantSubmitBody)
	}

	wantAssignBody := `{"edge_id":"edge-1","robot_ref":"robot:tb4-01","lease_ref":"lease:wh3/tb4-01/4","fencing_token":42,"assignment_reason":"nearest"}`
	if got := bodies["POST /missions/mission:wh3/patrol-17/assignments"]; got != wantAssignBody {
		t.Fatalf("assign body = %q, want %q", got, wantAssignBody)
	}
}

func TestMissionOperationCalls(t *testing.T) {
	var requests []string
	bodies := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath()+"?"+r.URL.RawQuery)
		raw, _ := io.ReadAll(r.Body)
		bodies[r.Method+" "+r.URL.Path] = string(raw)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/missions/mission:wh3/patrol-17/operations":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(MissionView{Status: "executing"})
		case r.Method == http.MethodGet && r.URL.Path == "/missions/mission:wh3/patrol-17/operations":
			_ = json.NewEncoder(w).Encode(MissionOperationList{Operations: []MissionOperation{{OperationID: "op-1"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/missions/mission:wh3/patrol-17/cancel":
			_ = json.NewEncoder(w).Encode(MissionView{Status: "cancelled", CancelRequested: true})
		case r.Method == http.MethodPost && r.URL.Path == "/missions/mission:wh3/patrol-17/effect-reports":
			_ = json.NewEncoder(w).Encode(EffectReportResult{Accepted: true, EffectStatus: "observed", MissionStatus: "completed"})
		case r.Method == http.MethodGet && r.URL.Path == "/edges/edge-1":
			_ = json.NewEncoder(w).Encode(Edge{ID: "edge-1", TenantID: "tenant-1", Version: "1.0.0", Endpoint: "wss://edge-1", RegisteredAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)})
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
	missionID := "mission:wh3/patrol-17"

	operation := MissionOperation{
		SchemaVersion:      "1.0.0",
		OperationID:        "operation:wh3/patrol-17/1",
		MissionID:          missionID,
		AssignmentID:       "assignment:wh3/patrol-17/1",
		EdgeID:             "akrovela-node-wh3-01",
		FencingToken:       42,
		CapabilityRef:      "cap:warehouse.navigate@1.0.0",
		DelegatedOperation: json.RawMessage(`{"schema_version":"akrovela.delegated-operation/v1","contract_mode":"live"}`),
	}
	view, err := c.SubmitMissionOperation(ctx, missionID, operation)
	if err != nil || view.Status != "executing" {
		t.Fatalf("submit operation = %+v, err = %v", view, err)
	}
	ops, err := c.GetMissionOperations(ctx, missionID)
	if err != nil || len(ops) != 1 || ops[0].OperationID != "op-1" {
		t.Fatalf("operations = %+v, err = %v", ops, err)
	}
	cancelled, err := c.CancelMission(ctx, missionID, "reason")
	if err != nil || !cancelled.CancelRequested {
		t.Fatalf("cancel = %+v, err = %v", cancelled, err)
	}
	result, err := c.ReportEffect(ctx, missionID, EffectReport{SchemaVersion: "1.0.0", ReportID: "report-1", MissionID: missionID, Status: "reported", Provenance: "akrovela-receipt"})
	if err != nil || !result.Accepted || result.EffectStatus != "observed" || result.MissionStatus != "completed" {
		t.Fatalf("effect report = %+v, err = %v", result, err)
	}
	edge, err := c.GetEdge(ctx, "edge-1")
	if err != nil || edge.ID != "edge-1" || edge.TenantID != "tenant-1" || !edge.RegisteredAt.Equal(time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("edge = %+v, err = %v", edge, err)
	}

	want := []string{
		"POST /missions/mission:wh3%2Fpatrol-17/operations?",
		"GET /missions/mission:wh3%2Fpatrol-17/operations?",
		"POST /missions/mission:wh3%2Fpatrol-17/cancel?",
		"POST /missions/mission:wh3%2Fpatrol-17/effect-reports?",
		"GET /edges/edge-1?",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v", requests)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
	for i := 0; i < 4; i++ {
		if !strings.Contains(requests[i], "%2F") {
			t.Fatalf("request[%d] = %q, want encoded slash %%2F", i, requests[i])
		}
	}

	if got := bodies["POST /missions/mission:wh3/patrol-17/cancel"]; !strings.Contains(got, `"reason":"reason"`) {
		t.Fatalf("cancel body = %q, want reason", got)
	}
	wantOperationBody := `{"schema_version":"1.0.0","operation_id":"operation:wh3/patrol-17/1","mission_id":"mission:wh3/patrol-17","assignment_id":"assignment:wh3/patrol-17/1","edge_id":"akrovela-node-wh3-01","fencing_token":42,"capability_ref":"cap:warehouse.navigate@1.0.0","delegated_operation":{"schema_version":"akrovela.delegated-operation/v1","contract_mode":"live"}}`
	if got := bodies["POST /missions/mission:wh3/patrol-17/operations"]; got != wantOperationBody {
		t.Fatalf("submit operation body = %q, want %q", got, wantOperationBody)
	}
	effectBody := bodies["POST /missions/mission:wh3/patrol-17/effect-reports"]
	if !strings.Contains(effectBody, `"report_id":"report-1"`) {
		t.Fatalf("effect report body = %q", effectBody)
	}
	if !strings.Contains(effectBody, `"status":"reported"`) || !strings.Contains(effectBody, `"provenance":"akrovela-receipt"`) {
		t.Fatalf("effect report body = %q, want reported akrovela-receipt", effectBody)
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

func TestMissionErrorIncompleteFallsBackToHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "x"})
	}))
	defer srv.Close()
	c, _ := New(Config{BaseURL: srv.URL})
	_, err := c.GetMission(context.Background(), "mission:wh3/patrol-17")
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want HTTPError fallback", err)
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %v, want HTTPError", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("httpErr = %+v", httpErr)
	}
}

func TestEffectExecutionStopConfirmedPointer(t *testing.T) {
	stop := false
	report := EffectReport{
		SchemaVersion: "1.0.0",
		ReportID:      "report-1",
		MissionID:     "mission:wh3/patrol-17",
		AssignmentID:  "assignment-1",
		OperationKey:  "operation-1",
		AttemptID:     "attempt-1",
		NodeID:        "node-1",
		BootEpoch:     1,
		FencingToken:  1,
		Status:        "reported",
		Provenance:    "akrovela-receipt",
		Execution:     &EffectExecution{StopConfirmed: &stop},
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"stop_confirmed":false`) {
		t.Fatalf("body = %s, want explicit stop_confirmed:false", raw)
	}

	report.Execution = &EffectExecution{}
	raw, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "stop_confirmed") {
		t.Fatalf("body = %s, want stop_confirmed omitted", raw)
	}
}

func TestEdgeReadErrorMapsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "edge_not_found", "message": "edge does not exist"})
	}))
	defer srv.Close()
	c, _ := New(Config{BaseURL: srv.URL})
	_, err := c.GetEdge(context.Background(), "edge-1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "edge_not_found" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestGetEdgeEscapesSlash(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.Method + " " + r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Edge{ID: "edge:wh3/01", TenantID: "tenant-1"})
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetEdge(context.Background(), "edge:wh3/01"); err != nil {
		t.Fatal(err)
	}
	if path != "GET /edges/edge:wh3%2F01" {
		t.Fatalf("path = %q, want encoded slash %%2F", path)
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
