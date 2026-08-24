# KINETOVELA Open

[English](README.md)

KINETOVELA 是面向异构机器人和边缘代理的受治理物理自主控制平面。它将已批准的能力计划转化为受安全边界约束、可观测、可恢复的 Mission，并以物理效果证据而非仅凭命令接受来闭合关键结果。

KINETOVELA-open 是其公开发行版，提供版本化 API 合约、SDK、公开示例、公开文档和二进制发布信息。

## 范围

KINETOVELA 负责 Fleet 和 Robot 生命周期、Mission 分配、离线执行边界、发布操作和效果核对。它与 ROS 2、Open-RMF、VDA 5050、OPC UA、供应商车队、导航栈、驱动程序及认证安全系统集成，而非取代它们。

## 仓库

| 仓库 | 用途 | 许可证 |
| --- | --- | --- |
| [`KINETOVELA-open`](https://github.com/axisrobo/kinetovela-open) | 公开合约、SDK、公开示例、公开文档和二进制发布信息 | Apache-2.0 |
| `KINETOVELA` | 控制平面核心实现（私有） | AGPL-3.0-only |
| `KINETOVELA-ee` | 企业扩展和内部设计文档 | Enterprise License |

## API 合约

公开合约在依赖它们的适配器或 SDK 变更之前发布。其与供应商无关的规范模型将 ROS 2、VDA 5050、Open-RMF、OPC UA 和供应商 API 作为适配器关注点。

## 目录结构

```text
contracts/        版本化 API 模式和兼容性夹具
sdk/go/           Go SDK 模块
examples/         公开示例文档和材料
docs/             公开架构和 API 文档
```

## 许可证

Apache License 2.0。参见 [LICENSE](LICENSE)。
