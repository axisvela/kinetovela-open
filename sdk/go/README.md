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
