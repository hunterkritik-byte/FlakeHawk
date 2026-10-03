# 🦅 FlakeHawk

> **Explainable flaky-test detection and CI reliability analytics for engineering teams.**

FlakeHawk turns historical CI results into evidence about flaky tests. It focuses on transparent signals, reproducible analysis, and actionable CI data instead of an opaque ML prediction.

**Made by Kritik Bhattarai** · **Security / inquiries:** hunterkritik@gmail.com

[![CI](https://github.com/hunterkritik-byte/FlakeHawk/actions/workflows/ci.yml/badge.svg)](https://github.com/hunterkritik-byte/FlakeHawk/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

## 🚀 Quick Start for Users

### 1. Clone FlakeHawk

```bash
git clone https://github.com/hunterkritik-byte/FlakeHawk.git
cd FlakeHawk
```

### 2. Build

Make sure Go 1.23+ is installed.

```bash
go build -o flakehawk ./cmd/flakehawk
```

Check the build:

```bash
./flakehawk version
```

### 3. Run tests

```bash
go test ./...
go vet ./...
```

### 4. Analyze JUnit reports

Analyze a single report:

```bash
./flakehawk report ./tests/testdata/results.xml
```

Analyze a directory:

```bash
./flakehawk report ./tests/testdata
```

JSON output:

```bash
./flakehawk report --format json ./tests/testdata
```

### 5. Ingest CI results

Store historical CI executions locally:

```bash
./flakehawk ingest ./test-results
```

By default, records are written to:

```text
.flakehawk/runs.jsonl
```

You can also provide metadata manually:

```bash
./flakehawk ingest \
  --commit "$(git rev-parse HEAD)" \
  --branch "$(git branch --show-current)" \
  --runner "Linux/X64" \
  --attempt 1 \
  ./test-results
```

### 6. GitHub Actions

FlakeHawk can ingest JUnit reports directly from GitHub Actions:

```yaml
- name: Generate test reports
  run: |
    # Run your test suite here.
    # Your test runner should produce JUnit XML.

- name: FlakeHawk ingestion
  uses: hunterkritik-byte/FlakeHawk/action@main
  with:
    reports: test-results
```

Inside GitHub Actions, FlakeHawk automatically reads:

- `GITHUB_SHA` — commit SHA
- `GITHUB_REF_NAME` / `GITHUB_HEAD_REF` — branch/ref
- `RUNNER_OS` + `RUNNER_ARCH` — runner
- `GITHUB_RUN_ATTEMPT` — CI attempt

### 7. Available commands

```text
flakehawk report <reports>
flakehawk ingest <reports>
flakehawk version
```

Run:

```bash
./flakehawk help
```

### 📁 Example project layout

```text
your-project/
├── .github/
│   └── workflows/
│       └── tests.yml
├── test-results/
│   ├── unit.xml
│   └── integration.xml
└── ...
```

### 🔄 Update from source

```bash
cd FlakeHawk
git pull
go build -o flakehawk ./cmd/flakehawk
```

### Requirements

For the current ingestion MVP:

- Git
- Go 1.23+
- JUnit XML test results

PostgreSQL is not required for local ingestion. The persistent PostgreSQL backend is planned for the next phase.

## Why FlakeHawk?

A CI failure does not always mean a product regression. Tests can fail because of nondeterministic timing, shared resources, retries, runner differences, or ordering.

FlakeHawk records evidence around failures and answers:
- Did the same test both pass and fail for the same commit?
- Did it fail and then pass on retry?
- How often does it fail?
- How much CI runtime is associated with observed failures?
- How strong is the available evidence?

The goal is not to hide failures automatically. The goal is to make reliability problems visible and explainable.

## Current MVP

- JUnit XML file and directory ingestion
- Same-commit disagreement signal
- Retry-pass signal
- Sample-size-aware confidence calculation
- Human-readable terminal reports
- JSON reports
- PostgreSQL starter schema
- GitHub Action
- Deterministic tests and fixtures

A flake assessment is evidence, not proof of root cause. Infrastructure failures and real regressions can also produce retry-pass patterns.

## Quick start

Requirements: Go 1.23+.

```bash
go build -o flakehawk ./cmd/flakehawk
./flakehawk report ./tests/testdata
./flakehawk report --format json ./tests/testdata
go test ./...
go vet ./...
```

For experiments with small fixtures:

```bash
./flakehawk report --min-runs 2 ./tests/testdata
```

## Architecture

```text
CI → JUnit XML → Parser → TestExecution → Detector → FlakeAssessment
                                      │
                                      ├── Text / JSON
                                      └── PostgreSQL → API → Dashboard
```

See [docs/architecture.md](docs/architecture.md).

## Roadmap

### Phase 1 — Local detector
- [x] JUnit parser
- [x] Same-commit disagreement
- [x] Retry-pass detection
- [x] Explainable confidence
- [x] Text/JSON reporting
- [x] Tests and fixtures

### Phase 2 — CI ingestion
- [ ] `flakehawk ingest`
- [x] Commit/branch/runner/attempt metadata
- [x] GitHub Actions metadata adapter
- [x] Duplicate-run protection

### Phase 3 — Persistent analytics
- [x] PostgreSQL schema
- [ ] PostgreSQL implementation
- [ ] Historical aggregation
- [ ] Retention policies
- [ ] ClickHouse adapter

### Phase 4 — Developer experience
- [ ] `flakehawk serve`
- [ ] REST API
- [ ] Per-test history
- [ ] GitHub App PR comments
- [ ] CODEOWNERS routing

### Phase 5 — Dashboard
- [ ] React/Vite dashboard
- [ ] Flaky-test history
- [ ] CI-hours wasted
- [ ] Runner/time correlation
- [ ] Commit correlation

### Phase 6 — Safe automation
- [ ] Quarantine configuration
- [ ] Owner/reason/expiry
- [ ] Audit trail
- [ ] No silent suppression

### Phase 7 — Advanced analysis
- [ ] Test-order dependency detection
- [ ] Root-cause hints
- [ ] Regression/bisect assistance
- [ ] pytest/Jest/Go/Bazel/Gradle adapters
- [ ] Public OSS CI-history research

## Repository layout

```text
cmd/flakehawk/          CLI
internal/model/         domain models
internal/parser/junit/  JUnit parser
internal/detector/      detection logic
internal/report/        text + JSON output
internal/storage/       storage abstraction
migrations/             PostgreSQL schema
action/                 GitHub Action
tests/testdata/         fixtures
docs/                   design documentation
```

## Design principles

1. Evidence first.
2. No silent suppression.
3. Reproducible analysis.
4. Small-history aware.
5. CI-provider neutral core.
6. Secure by default.
7. Reversible automation.

## Contributing

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./cmd/flakehawk
```

See [SECURITY.md](SECURITY.md) for security reporting.

## License

Apache License 2.0. See [LICENSE](LICENSE).
