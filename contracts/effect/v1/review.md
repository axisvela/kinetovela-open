# Contract Review — EffectReport v1

**Status:** reviewed for integration, 2026-11-11. This is an in-repository
negative-case review, not external customer sign-off, runtime conformance or
security certification.

## Identity

| Item | Value |
| --- | --- |
| Contract | `contracts/effect/v1/effect-report.schema.json` |
| Contract line | `v1` |
| Product version | `1.0.0` |
| Consumer | KINETOVELA core and AKROVELA live delegation |
| Maintainer | KINETOVELA maintainers |
| Review method | Fresh-context review of schema, fixtures and authority boundaries |
| Core reviewed | `KINETOVELA` commit `545023997bff5f6287a80fe5817c1782e52ae570` |
| Contracts reviewed | `KINETOVELA-open-contracts` commits `5c67cbcc195c3a481b0180e94f548e3c4955b0f1` (schema) and `423e1760021853d797e61245648ea3526d170976` (OpenAPI) |

## Ownership

KINETOVELA owns Mission state, assignment, fencing and physical effect authority.
AKROVELA owns local operations, attempts, receipts and the Execution facts carried
in a report. The EffectReport transports those facts; it is not a completion
signal, a safety claim or a substitute for delegated operation authority. This
in-repository review is not external sign-off.

## Negative and Fault Cases

| Case | Required outcome | Coverage |
| --- | --- | --- |
| Local `verified` relabelled as `observed` | Reject; preserve the local status and provenance rather than promoting or relabelling it | `effect-report-observed-claim.json`; status/provenance constraint |
| `stop_confirmed: true` without a source | Reject; stop confirmation requires a sourced fact | Phase 2 server gate; the schema cannot see a source |
| `mission_completed` field | Reject as an unknown property; a report never completes a Mission | `effect-report-completion-claim.json`; `additionalProperties: false` |
| Bad status | Reject values outside `reported`, `observed`, `failed`, `unknown_effect` | `effect-report-bad-status.json` |
| Missing fence | Reject as invalid; a report requires a positive fencing token | `effect-report-missing-fence.json` |
| Wrong boot epoch | Reject a stale node boot epoch; no effect accepted | Schema minimum and server gate; stateful test required |
| Wrong node | Reject a report from a node other than the assigned edge/node | Identity binding and review |
| Replayed report | Accept at most once; no second effect or transition | `report_id` and review; stateful test required |
| Wrong tenant or edge | Reject; no cross-tenant read or write | Scope binding and review |

## Findings Resolved Before Review

- Reports require `mission_id`, `assignment_id`, `operation_key`, `attempt_id`,
  `node_id`, `boot_epoch`, `fencing_token`, `status`, `observed_at` and
  `provenance`.
- Provenance is fixed to `akrovela-receipt`; the report cannot claim an
  independent verifier. With that provenance, `status` is constrained to
  `reported`, `failed` or `unknown_effect`; an `observed` claim is invalid.
- `execution` separates effect status, cancellation status and completion from
  Mission completion; `stop_confirmed` is a boolean fact, not a cancel result.
- Top-level completion fields such as `mission_completed` are rejected by
  `additionalProperties: false`; `details` is opaque and not interpreted as
  completion.
- The report result conveys acceptance and Mission status without asserting
  completion from the report itself.

## Residual Risks

| Risk | Effect | Next gate |
| --- | --- | --- |
| No external consumer sign-off | Contract may not match consumer needs | Consumer review before live integration |
| No runtime ingestion implementation | Contract is not runtime evidence | Live delegation gate |
| Stateful replay, boot-epoch and cross-tenant tests are not executed | Conflict behavior is specified but not proven | Deterministic integration gate |
| No behavioral security tests | Review is not security evidence | Independent security review |

## Decision

Publish `v1` as a reviewed integration contract. Keep Mission completion, stop
claims and hardware effect claims blocked until their owning gates produce
evidence.
