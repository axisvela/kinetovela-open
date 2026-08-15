# KINETOVELA Open

KINETOVELA is the governed physical-autonomy control plane for heterogeneous robots and edge agents. It turns approved capability plans into safety-bounded, observable, recoverable Missions with physical effect evidence.

This public repository publishes the API contracts, SDKs, examples, documentation, and binary release information for KINETOVELA.

## Scope

KINETOVELA governs Fleet and Robot lifecycle, Mission assignment, offline execution bounds, release operations, and effect reconciliation. It integrates with, rather than replaces, ROS 2, Open-RMF, VDA 5050, OPC UA, vendor fleets, navigation stacks, drivers, and certified safety systems.

## Repositories

| Repository | Purpose | License |
| --- | --- | --- |
| [`KINETOVELA-open`](https://github.com/axisrobo/KINETOVELA-open) | Public contracts, SDKs, examples, docs, and core binary releases | Apache-2.0 |
| [`KINETOVELA`](https://github.com/axisrobo/KINETOVELA) | Control-plane core implementation | AGPL-3.0-only |
| `KINETOVELA-ee` | Enterprise extensions and internal design documentation | Enterprise License |

## API Contracts

Canonical public schemas use semantic versioning (`major.minor.patch`). A Mission is an execution commitment, not a planning tree: it binds plan, capability, authority, safety envelope, robot assignment, checkpoints, and expected effects.

The first contract freeze covers `RobotDescriptor`, `FleetSpec`, `Mission`, `MissionAssignment`, `SafetyEnvelope`, `OfflineLease`, and `EffectRecord`.

## Layout

```text
contracts/        Versioned API schemas and compatibility fixtures
sdk/go/           Go SDK module
examples/         Runnable integration examples
docs/             Public architecture and API documentation
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
