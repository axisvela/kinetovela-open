# KINETOVELA Go SDK

Module: `github.com/axisvela/KINETOVELA-open/sdk/go/v3`

```go
import sdk "github.com/axisvela/KINETOVELA-open/sdk/go/v3/kinetovela"

c, err := sdk.New(sdk.Config{BaseURL: "http://127.0.0.1:1926", Token: "..."})
if err != nil { /* ... */ }

err = c.Health(ctx)
version, err := c.Version(ctx)
slo, err := c.FleetSLO(ctx)
err = c.RegisterEdge(ctx, sdk.EdgeRegistration{ID: "edge-1", TenantID: "acme"})
err = c.Heartbeat(ctx, "edge-1", 5)
err = c.BindMission(ctx, "edge-1", "mission-1", "instance-1")
err = c.ReportSync(ctx, "edge-1", 3, 42)
```

The edge-fleet methods target the control-plane edge API. The current public
KINETOVELA server exposes `healthz`, `version`, and `fleet/slo`; edge routes are
available in deployments that enable the fleet control plane.

## Mission lifecycle

```go
mission, err := c.SubmitMission(ctx, sdk.Mission{
	SchemaVersion:     "1.0.0",
	MissionID:         "mission:wh3/patrol-17",
	PlanRef:           "plan:wh3/patrol",
	CapabilityRefs:    []string{"cap:warehouse.navigate@1.0.0"},
	AuthorityRef:      "grant:wh3-ops",
	SafetyEnvelopeRef: "safe:wh3",
	ExpectedEffects:   []string{"robot:tb4-01.pose=target-17"},
})

view, err := c.GetMission(ctx, mission.Mission.MissionID)
missions, err := c.ListMissions(ctx, sdk.MissionFilter{EdgeID: "edge-1", Status: "validated", Limit: 10})

view, err = c.AssignMission(ctx, mission.Mission.MissionID, sdk.AssignmentRequest{
	EdgeID:           "edge-1",
	RobotRef:         "robot:tb4-01",
	LeaseRef:         "lease:wh3/tb4-01/4",
	FencingToken:     42,
	AssignmentReason: "nearest",
})

view, err = c.SubmitMissionOperation(ctx, mission.Mission.MissionID, sdk.MissionOperation{
	SchemaVersion: "1.0.0",
	OperationID:   "op-1",
	MissionID:     mission.Mission.MissionID,
	EdgeID:        "edge-1",
	FencingToken:  42,
})
operations, err := c.GetMissionOperations(ctx, mission.Mission.MissionID)

view, err = c.CancelMission(ctx, mission.Mission.MissionID, "operator request")

result, err := c.ReportEffect(ctx, mission.Mission.MissionID, sdk.EffectReport{
	SchemaVersion: "1.0.0",
	ReportID:      "report-1",
	MissionID:     mission.Mission.MissionID,
	AssignmentID:  "assignment-1",
	OperationKey:  "op-1",
	AttemptID:     "attempt-1",
	NodeID:        "node-1",
	BootEpoch:     1,
	FencingToken:  42,
	Status:        "observed",
	ObservedAt:    time.Now().UTC(),
	Provenance:    "akrovela-receipt",
})

edge, err := c.GetEdge(ctx, "edge-1")
```

These methods target the Mission API implemented in KINETOVELA core Phase 2a
and operate on process-local state on the core server. They do not perform live
execution, ROS 2 telemetry or command transport, push dispatch, or any
physical-effect claim. `ReportEffect` records the report the core accepted; it
does not confirm that an effect occurred on hardware.

Mission IDs may contain reserved characters such as `/`. The SDK URL-encodes
the mission ID on every mission path (for example `mission:wh3/patrol-17`
becomes `mission:wh3%2Fpatrol-17`); callers pass the raw ID.

Mission and edge read methods map structured error responses to `*sdk.APIError`.
The type carries `StatusCode`, `Code` and `Message` and is returned when the
error body includes both a code and a message; otherwise the SDK falls back to
`*sdk.HTTPError`. The older edge-fleet methods continue to return `*sdk.HTTPError`.

```go
var apiErr *sdk.APIError
if errors.As(err, &apiErr) {
	log.Printf("status=%d code=%s message=%s", apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```
