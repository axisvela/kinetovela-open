# Public Architecture

KINETOVELA has an enterprise control plane, a site and edge plane, robot adapter nodes, and an independent local safety boundary.

The control plane never forms a certified or hard-real-time safety loop. A robot-local safety supervisor, emergency stop integration, geofence enforcement, and safe-state behavior must remain independently operable when cloud connectivity or model inference is unavailable.

KINETOVELA retains Mission, assignment, fencing, and physical-effect lifecycle
authority. In the target family boundary, AKROVELA owns local operations,
execution attempts, and receipts; its receipts do not independently complete a
Mission. The existing KINETOVELA Site Gateway implementation is retained, not
already migrated to AKROVELA. AKROVELA and DIDACTOVELA are unreleased `0.1.0`
bootstrap only, with no functional integrations. Current defaults and offline
authority bounds are unchanged.

See the [family architecture](https://github.com/axisvela/axisvela/blob/main/EDGE-AND-LEARNING-ARCHITECTURE.md)
and [authority boundaries](https://github.com/axisvela/axisvela/blob/main/AUTHORITY-BOUNDARIES.md).
