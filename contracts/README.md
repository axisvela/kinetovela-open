# Versioned Contracts

Contracts are published before dependent adapters or SDKs change. Canonical object models are vendor-neutral; ROS 2, VDA 5050, Open-RMF, OPC UA, and vendor APIs are adapter concerns.

The initial contract set is synchronized from the core repository after review and compatibility validation.

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
`effect/v1/review.md`. Fixtures under `fixtures/v1/` exercise the valid and
invalid cases named by those reviews.
