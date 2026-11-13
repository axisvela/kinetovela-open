# Contract Review — Governed Mission v1

**Status:** reviewed for integration, 2026-11-11. This is an in-repository
negative-case review, not external customer sign-off, runtime conformance or
security certification.

## Identity

| Item | Value |
| --- | --- |
| Contract | `contracts/mission/v1/mission.schema.json`, `contracts/mission/v1/mission-assignment.schema.json`, `contracts/mission/v1/mission-operation.schema.json`, `contracts/mission/v1/mission-api.openapi.json` |
| Contract line | `v1` |
| Product version | `1.0.0` |
| Consumer | KINETOVELA core and AKROVELA live delegation |
| Maintainer | KINETOVELA maintainers |
| Review method | Fresh-context review of contracts, schemas, fixtures and authority boundaries |
| Core reviewed | `KINETOVELA` commit `545023997bff5f6287a80fe5817c1782e52ae570` |
| Contracts reviewed | `KINETOVELA-open-contracts` commit `423e1760021853d797e61245648ea3526d170976` |

## Ownership

KINETOVELA owns Mission intent, assignment, lease fencing and physical effect
authority. AKROVELA owns local operation admission, attempts, execution facts,
cancellation facts and receipts. The Mission API records delegated operations and
effect reports; it does not execute operations or certify physical effects. No
component may substitute its fact for another's authority, and acceptance in this
repository is not external sign-off.

## Negative and Fault Cases

| Case | Required outcome | Coverage |
| --- | --- | --- |
| Missing authority reference | Reject as invalid; no Mission accepted | `mission-missing-authority.json` |
| Zero fencing token | Reject as invalid at assignment and operation submission | `mission-assignment-zero-fence.json`, `mission-operation-zero-fence.json`, `assignment-request-zero-fence.json` |
| Stale fencing token | Reject with a conflict; no later operation or effect accepted | Schema minimum and server gate; stateful test required |
| Duplicate assignment | Reject the second assignment; do not silently replace robot or lease | Mission state gate; stateful test required |
| Cross-tenant edge assignment | Reject; no cross-tenant read or write | Authority scope binding and review |
| Unknown Mission | Reject with `404`; do not create an implicit Mission | Mission API response contract |
| Operation reference mismatch | Reject when path `mission_id`, body `mission_id`, `assignment_id` or `edge_id` disagree | Phase 2 server gate; `mission-operation-missing-assignment.json` only tests an absent field, not a mismatch |
| Receipt or effect report presented as Mission completion | Reject; an accepted report never transitions the Mission to `completed` | `effect-report-completion-claim.json`; Mission API response contract |
| Cancel request presented as remote stop | Reject; cancellation is a request and `stop_confirmed` remains a separate sourced fact | Cancel response contract and review |
| Remote `$ref` in a schema or the API document | Reject; all references stay in-document | `test_no_remote_schema_refs`, `test_mission_api_refs_resolve` |

## Findings Resolved Before Review

- Mission intent, assignment, delegated operation and effect report remain
  separate objects; none implies completion of another.
- Assignment carries a positive fencing token and a lease reference; zero fences
  are invalid.
- Mission operations bind `mission_id`, `assignment_id`, `edge_id`, a capability
  reference and a positive fencing token.
- The Mission API exposes submit, query, assignment, operation, cancel and
  effect-report paths without claiming execution or remote stop.
- The OpenAPI components match the canonical schemas after excluding only schema
  metadata, and every `$ref` resolves locally.
- Effect reports never carry a `mission_completed` field and cannot complete a
  Mission.

## Additive Pagination Update (2026-11-12)

An additive, backward-compatible update was reviewed in-repository:

- `GET /missions` accepts optional `limit` (`integer`, minimum 1) and `cursor`
  (`string`, `minLength: 1`) query parameters. Omitting `limit` uses the default
  page size of 100; values above 500 are capped at 500. Listing is ordered by
  `updated_at` descending then `mission_id` descending, so the first page is not
  the previous ascending unpaginated listing.
- `MissionList` may include `next_cursor` (`string`). It is optional, so existing
  consumers that read only `missions` remain valid; an empty or absent value
  means no further page.
- The Go SDK gained `MissionFilter.Cursor`, `MissionPage` and
  `ListMissionsPage`; `ListMissions` remains a compatibility wrapper returning
  `page.Missions`.

This update was verified by the repository contract tests and SDK tests only. It
is not a new external review, customer sign-off, runtime conformance or security
certification.

## Residual Risks

| Risk | Effect | Next gate |
| --- | --- | --- |
| No external consumer sign-off | Contract may not match consumer needs | Consumer review before live integration |
| No runtime Mission service implementation | Contract is not runtime evidence | Live delegation gate |
| Stateful assignment, fencing and replay tests are not executed | Conflict behavior is specified but not proven | Deterministic integration gate |
| No security certification | Review is not security evidence | Independent security review |

## Decision

Publish `v1` as a reviewed integration contract. Keep live execution, remote
stop claims and hardware effect claims blocked until their owning gates produce
evidence.
