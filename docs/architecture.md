# FlakeHawk Architecture

## Phase 2: CI ingestion

CI reports enter through the JUnit parser and are enriched with execution metadata before being persisted to an append-only JSONL ingestion store.

Flow: GitHub Actions -> JUnit XML -> flakehawk ingest -> metadata adapter -> run identity -> duplicate check -> JSONL store -> detector.

## GitHub Actions metadata

When flakehawk ingest runs inside GitHub Actions it reads:
- GITHUB_SHA -> commit SHA
- GITHUB_REF_NAME or GITHUB_HEAD_REF -> branch/ref
- RUNNER_OS + RUNNER_ARCH -> runner identity
- GITHUB_RUN_ATTEMPT -> retry attempt

Every parsed execution receives those values.

CLI flags override environment-derived values, which makes local replay and other CI providers possible.

## Duplicate-run protection

A run ID is a SHA-256 digest of the JUnit report bytes plus commit, branch, runner, and attempt metadata.

The JSONL store is append-only. Before writing a record, FlakeHawk scans existing run IDs and ignores an exact duplicate.

This protects against a workflow accidentally uploading the same report twice. A future PostgreSQL backend should enforce the same invariant with a unique constraint.

## Design note

The JSONL store is intentionally dependency-free for this phase. PostgreSQL remains the persistent multi-run backend planned for the next storage phase.
