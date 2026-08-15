# Public Architecture

KINETOVELA has an enterprise control plane, a site and edge plane, robot adapter nodes, and an independent local safety boundary.

The control plane never forms a certified or hard-real-time safety loop. A robot-local safety supervisor, emergency stop integration, geofence enforcement, and safe-state behavior must remain independently operable when cloud connectivity or model inference is unavailable.
