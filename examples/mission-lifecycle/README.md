# Mission Lifecycle Example

This example follows the governed flow:

```text
Goal or Plan -> Governed Mission -> Assignment -> Execution -> Observed Effect -> Recover or Replan
```

See the published v1 contracts for the flow: `contracts/mission/v1/mission.schema.json`,
`contracts/mission/v1/mission-assignment.schema.json`,
`contracts/mission/v1/mission-operation.schema.json`,
`contracts/mission/v1/mission-api.openapi.json` and
`contracts/effect/v1/effect-report.schema.json`. The matching review records are
`contracts/mission/v1/review.md` and `contracts/effect/v1/review.md`.

Mission dispatch is still not implemented; the contracts define the interface
only.
