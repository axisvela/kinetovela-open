# KINETOVELA Open

[中文](README.zh-CN.md)

KINETOVELA is a governed physical-autonomy control plane for heterogeneous robots and edge agents. It turns approved capability plans into safety-bounded, observable, recoverable Missions with physical-effect evidence rather than command acceptance alone.

KINETOVELA-open is its public distribution for versioned API contracts, SDKs, public examples, public docs, and binary release information.

## Scope

KINETOVELA governs Fleet and Robot lifecycle, Mission assignment, offline execution bounds, release operations, and effect reconciliation. It integrates with rather than replaces ROS 2, Open-RMF, VDA 5050, OPC UA, vendor fleets, navigation stacks, drivers, and certified safety systems.

Mission/assignment/fencing and effect authority stay with KINETOVELA; robot-local
safety stays independent. AKROVELA targets local operations, execution attempts,
and receipts, but the existing Site Gateway has not migrated. AKROVELA and
DIDACTOVELA are unreleased `0.1.0` bootstrap only, with no functional integrations
or changes to defaults and offline authority bounds. See [public architecture](docs/architecture.md),
the [family architecture](https://github.com/axisvela/axisvela/blob/main/EDGE-AND-LEARNING-ARCHITECTURE.md),
and [authority boundaries](https://github.com/axisvela/axisvela/blob/main/AUTHORITY-BOUNDARIES.md).

## Repositories

| Repository | Purpose | License |
| --- | --- | --- |
| [`KINETOVELA-open`](https://github.com/axisvela/kinetovela-open) | Public contracts, SDKs, public examples, public docs, and binary release information | Apache-2.0 |
| `KINETOVELA` | Control-plane core implementation (private) | AGPL-3.0-only |
| `KINETOVELA-ee` | Enterprise extensions and internal design documentation | Enterprise License |

## API Contracts

Public contracts are published before dependent adapters or SDKs change. Their vendor-neutral canonical models keep ROS 2, VDA 5050, Open-RMF, OPC UA, and vendor APIs as adapter concerns.

## Layout

```text
contracts/        Versioned API schemas and compatibility fixtures
sdk/go/           Go SDK module
examples/         Public example documentation and materials
docs/             Public architecture and API documentation
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
