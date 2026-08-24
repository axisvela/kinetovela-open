# KINETOVELA-open Repository Positioning Design

## Purpose

Make the public repository explain KINETOVELA before explaining the repository. Visitors should understand the operational problem KINETOVELA solves, the system boundary, and the role of KINETOVELA-open without needing to visit another repository.

## Scope

This change updates public repository positioning only:

- Rewrite the English README as the primary entry point.
- Add a Chinese README with equivalent scope and structure.
- Add a language link from the English README to the Chinese README.
- Set the GitHub repository description and discovery topics.

It does not change contracts, SDKs, examples, binaries, licensing, release procedure, or control-plane behavior.

## Content Design

### English README

The README starts with a link to `README.zh-CN.md`.

Its opening uses this order:

1. Define KINETOVELA as a governed physical-autonomy control plane for heterogeneous robots and edge agents.
2. Explain the problem it solves: turning approved capability plans into safety-bounded, observable, recoverable Missions with physical-effect evidence. Command acceptance is distinguished from evidence of a completed effect.
3. Define KINETOVELA-open as the public distribution of the project's contracts, SDKs, examples, documentation, and core binary releases.

The remainder retains the existing repository's useful information while making product boundaries explicit:

- Scope and integration boundary, including that KINETOVELA complements rather than replaces ROS 2, Open-RMF, VDA 5050, OPC UA, vendor fleets, navigation stacks, drivers, and certified safety systems.
- Relationship of KINETOVELA-open, KINETOVELA, and KINETOVELA-ee.
- Public contract model and versioning.
- Repository layout and Apache-2.0 license.

### Chinese README

`README.zh-CN.md` is a Chinese translation of the English README. It preserves the same section order, repository boundaries, and safety/effect-evidence distinction. The English README remains the canonical primary entry point; the Chinese file must not introduce different product claims.

### GitHub Metadata

The About description uses the same order as the README: first KINETOVELA and its problem space, then KINETOVELA-open. It must fit GitHub's 350-character description limit.

Proposed description:

> KINETOVELA is a governed physical-autonomy control plane for safe, observable robot fleet execution. KINETOVELA-open provides its public contracts, SDKs, examples, docs, and binary releases.

Topics:

`robotics`, `fleet-management`, `robot-fleet-management`, `physical-ai`, `autonomous-systems`, `edge-computing`, `industrial-automation`, `golang`, `postgresql`, `open-source`.

## Verification

- Check that English README links to the Chinese README and that the target exists.
- Compare the two READMEs to confirm equivalent claims and section coverage.
- Read repository metadata with GitHub CLI after update and confirm the exact description and all ten topics.
- Review the working-tree diff to ensure only documentation and repository metadata are changed.

## Self-Review

The scope is limited to positioning and discoverability. The content order, exact metadata description, target files, topic set, and verification criteria are explicit. No product behavior or external API is affected.
