# Versioned Contracts

Contracts are published before dependent adapters or SDKs change. Canonical object models are vendor-neutral; ROS 2, VDA 5050, Open-RMF, OPC UA, and vendor APIs are adapter concerns.

The initial contract set is synchronized from the core repository after review and compatibility validation.

## Synced Canonical Schemas

These ten schemas are synchronized byte-for-byte from `KINETOVELA` and recorded
in `sync-manifest.json`:

| Contract | Path |
| --- | --- |
| Offline lease | `authority/v1/offline-lease.schema.json` |
| Embodied capability binding | `capability/v1/embodied-capability-binding.schema.json` |
| Compatibility report | `compatibility/v1/compatibility-report.schema.json` |
| Effect record | `evidence/v1/effect-record.schema.json` |
| Fleet specification | `fleet/v1/fleet-spec.schema.json` |
| Governed Mission | `mission/v1/mission.schema.json` |
| Mission assignment | `mission/v1/mission-assignment.schema.json` |
| Robot descriptor | `robot/v1/robot-descriptor.schema.json` |
| Safety envelope | `safety/v1/safety-envelope.schema.json` |
| Robot state reference | `state/v1/robot-state-reference.schema.json` |

## Mission and Effect Contracts

The canonical Mission, assignment, operation and effect contracts and the
Mission API are published here:

| Contract | Path |
| --- | --- |
| Governed Mission | `mission/v1/mission.schema.json` |
| Mission assignment | `mission/v1/mission-assignment.schema.json` |
| Mission operation | `mission/v1/mission-operation.schema.json` |
| Mission API | `mission/v1/mission-api.openapi.json` |
| Effect report | `effect/v1/effect-report.schema.json` |

Review records live beside the contracts in `mission/v1/review.md` and
`effect/v1/review.md`.

## Fixture Coverage Versus Server Gates

Fixtures under `fixtures/v1/` enforce only the cases the review records map to a
fixture. They exercise schema shape, required fields, enums, minimums and
provenance-conditioned status rejection.

Stateful and authority cases are **not** proven by fixtures and remain server
gates: stale or duplicate fencing, replays, boot-epoch and node identity,
cross-tenant scope, operation reference mismatches, and sourced
`stop_confirmed`. The Mission and EffectReport review records name each of these
and state its real gate.
