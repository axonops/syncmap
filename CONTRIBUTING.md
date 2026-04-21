# Contributing to syncmap

Thank you for your interest in contributing to `github.com/axonops/syncmap`. This document covers the expectations for code, tests, documentation, and release discipline.

## Contributor License Agreement

Every contributor must sign our [Contributor License Agreement](./CLA.md) before a pull request can be merged. This is a one-time step per GitHub account and covers every future contribution you make to any AxonOps open-source project.

The CLA Assistant bot will comment on your first pull request with the signing instructions — you reply with one sentence and you are done. The process takes under a minute. Your signature is recorded in `signatures/version1/cla.json` (the audit trail) and you appear in the auto-generated [`CONTRIBUTORS.md`](./CONTRIBUTORS.md) (the public thank-you list).

**Why we require it.** The CLA makes it explicit that (a) you have the right to contribute the code, (b) AxonOps has the licence to distribute your contributions under the project's Apache Licence 2.0, and (c) the project is legally protected if a dispute arises about contributed code. Signing the CLA does NOT change your rights to use your own contributions for any other purpose.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you agree to uphold its standards. Report unacceptable behaviour privately to `oss@axonops.com`.

## Attribution policy

Commit messages, PR descriptions, code comments, commit trailers, and any other artefact that lands on `main` must not reference AI-tooling product names or mark content as AI-produced. The specific token list and enforcement regex live in [`llms.txt`](./llms.txt) and the CI `attribution-guard` job. This applies whether the contribution was produced by a human, an AI assistant, or both — the tooling is irrelevant to the audit trail.

## Your first pull request

1. **Fork** the [repository](https://github.com/axonops/syncmap) and clone your fork.
2. Create a feature branch from `main`: `feature/<short-name>` or `fix/<short-name>`.
3. Make your changes. Every change that touches `syncmap.go` or the test suite must go through the agent-gate stack below.
4. Push the branch to your fork and open a PR against `axonops/syncmap:main`.
5. Sign the CLA when the bot prompts you (only needed on your first PR).
6. A maintainer reviews; the agent gates run in CI; once everything is green the PR is squash-merged.

## Branching and commits

- **Main branch:** `main` — always buildable, always passes CI.
- **Feature work:** `feature/<short-name>` branched from `main`.
- **Bug fixes:** `fix/<short-name>` branched from `main`.
- Never commit directly to `main`.
- **Conventional commits** — `feat:`, `feat!:`, `fix:`, `test:`, `docs:`, `chore:`, `refactor:`, `perf:`, `ci:`. One logical change per commit. Subject ≤ 72 characters including the `(#<issue>)` suffix.
- **Every commit references an issue.** `TODO` comments in source must carry a GitHub issue number.
- No merge commits — rebase workflow.

## Test requirements

- Every change runs the full quality gate: `make check`.
- Unit tests use external black-box package (`syncmap_test`) with `testify` and `goleak`. Every top-level `Test…` and every `t.Run` subtest must call `t.Parallel()`. No `time.Sleep` as synchronisation; use `sync.WaitGroup`. No `fmt.Println` / `fmt.Printf` in test code.
- Tests run under `-race` always.
- Coverage gate: **95 %** on the library package. CI fails the build below the threshold.
- Concurrency-touching changes require at least one test that exercises the concurrent path. `TestLoadOrStoreContention`, `TestSwapContention`, `TestCompareAndSwapContention`, `TestConcurrentWritersReaders`, `TestRangeDuringWrites`, `TestDeleteDuringRange` are the existing patterns — match their shape.
- **BDD is the contract.** Every new public symbol or behavioural change adds a scenario to `tests/bdd/features/syncmap.feature` and any new step definitions to `tests/bdd/steps/steps.go`. Scenarios run under godog strict mode; unimplemented or pending steps fail the build.
- **Benchmarks.** Every public method has a benchmark in `syncmap_bench_test.go`. Implementation changes that could affect performance require a regenerated `bench.txt` in the same PR; `benchstat-regression-guard` fails any time/op regression ≥ 10 % at p ≤ 0.05 or any positive allocs/op delta.

## Performance baseline

The committed `bench.txt` is the reference against which every PR's performance is measured. Regenerate locally with:

```bash
make bench > current.txt        # raw five-sample run
make bench-regression           # benchstat diff vs committed baseline
```

CI runs the same comparison on `ubuntu-latest`. Shared GitHub-hosted runners exhibit ±5–15 % variance for nanosecond-scale benchmarks, which the 10 % time/op threshold absorbs. If the regression guard becomes chronically flaky, options are (a) a dedicated runner, (b) a higher time/op threshold (allocs/op stays strict since allocation counts are deterministic), or (c) making the job advisory. Open an issue before changing the policy.

## Agent-gate stack

Every PR flows through a fixed sequence of review agents in addition to the human reviewer. The gates are enforced by `CLAUDE.md` in the repo root and are a condition of merge. They fall into three buckets by lifecycle:

**Before filing an issue:**

- `issue-writer` — verifies the issue has binary, testable acceptance criteria.

**During feature work (run as you code):**

- `test-writer` / `test-analyst` — before and after writing tests.
- `code-reviewer` / `security-reviewer` / `performance-reviewer` — after changing source.
- `docs-writer` — after changing documentation.
- `devops` — after changing CI/CD, Makefile, GoReleaser.

**Before every commit (non-negotiable):**

- `go-quality` — final Go quality sweep.
- `commit-message-reviewer` — enforces conventional-commit format, issue reference, subject length, no AI attribution.

**Before closing any issue:**

- `issue-closer` — walks each acceptance criterion and confirms it is met.

## Documentation

- Every exported symbol has a godoc comment that starts with the symbol name and is at least 20 characters of real prose (the `TestDocumentation_EveryExportedSymbolHasGodoc` test enforces this).
- The README Quick Start block is compile-tested by `TestReadmeQuickStart_Compiles` — if you change the snippet, you change behaviour and the test catches the drift.
- `llms.txt` is the concise AI-assistant summary (≤ 2250 words). `llms-full.txt` is the concatenated corpus, regenerated by `scripts/gen-llms-full.sh`. Any edit to the source documents (README, `doc.go`, SECURITY, CHANGELOG, CONTRIBUTING) requires running `make llms-full` and committing the regenerated file. The `llms-full-up-to-date` CI job will catch any drift.

## Releases

Releases happen exclusively through the [release workflow](./.github/workflows/release.yml) triggered via `workflow_dispatch`. Never create tags locally — the `tag` job in the workflow is the only permitted tag-creation path. The release workflow runs the full quality gate first, creates an annotated tag under the `github-actions[bot]` identity, runs GoReleaser, and warms the Go module proxy so `pkg.go.dev` indexes the new version promptly.

## Reporting security issues

Do **not** open a public issue for a suspected vulnerability. Use GitHub's private advisory flow via the [Security tab](https://github.com/axonops/syncmap/security/advisories/new). See [`SECURITY.md`](./SECURITY.md) for the full disclosure process and response timeline.

## Licence

By contributing to this project, you agree that your contributions will be licensed under the project's [Apache Licence 2.0](./LICENSE), as documented in the [CLA](./CLA.md).
