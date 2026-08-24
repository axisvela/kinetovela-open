# KINETOVELA Open

[中文](README.zh-CN.md)

KINETOVELA is a governed physical-autonomy control plane for heterogeneous robots and edge agents. It turns approved capability plans into safety-bounded, observable, recoverable Missions with physical-effect evidence rather than command acceptance alone.

KINETOVELA-open is its public distribution for versioned API contracts, SDKs, public examples, public docs, and binary release information.

## Scope

KINETOVELA governs Fleet and Robot lifecycle, Mission assignment, offline execution bounds, release operations, and effect reconciliation. It integrates with rather than replaces ROS 2, Open-RMF, VDA 5050, OPC UA, vendor fleets, navigation stacks, drivers, and certified safety systems.

## Repositories

| Repository | Purpose | License |
| --- | --- | --- |
| [`KINETOVELA-open`](https://github.com/axisrobo/kinetovela-open) | Public contracts, SDKs, public examples, public docs, and binary release information | Apache-2.0 |
| `KINETOVELA` | Control-plane core implementation (private) | Apache-2.0 |
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
