# PinShare Test Coverage Improvement Session Context

## Project Overview
- **Repo**: PinShare (Go project with IPFS/P2P file sharing; packages: main, internal/{api,app,cmd,config,p2p,psfs,store}; dev_code ignored).
- **Current Coverage**: ~22.4% overall (`go test ./... -cover`), but effective ~10-15% (many 0% packages; build failures in p2p/psfs; flaky app tests).
- **Key Constraints/Preferences** (persist):
  - **Ignore**: All `internal/dev_code/*` (experimental, out of scope).
  - **Deprioritize**: Security/psfs/* (user unsatisfied, plans to tear down/rebuild; skip sec tests/mocks in app too).
  - Focus: Core (store, p2p, api), CLI (cmd), config/app (non-sec).
  - Style: Table-driven tests, mocks (testify/mock; add to go.mod if needed), no real externals (mock IPFS/libp2p/VT).
  - Targets: 40% overall first (Phase 1-2), then 70%+ per pkg. Use `go test -coverprofile=cov.out && go tool cover -html=cov.out`.
  - Tools: testify/suite, httptest (api), cobra testing (cmd), `-race`, CI in `.github/workflows/main.yml`.

## What Was Done So Far
- **Analysis** (initial): Ran `go test ./... -cover`, `go list ./...`, glob `*_test.go`. Detailed package/test file breakdown (8 pkgs, 6 test files partial). Identified blocks: p2p/uploads.go (fmt %w), psfs/sec_test.go (assignment mismatch), app/app_test.go (flaky ports/deps/sec).
- **Planning Iterations**:
  - Full analysis/suggestions (store/p2p/api/cmd top priority).
  - Refined: Excl. dev_code → phased plan.
  - Refined: Excl. psfs/sec → updated plan.
  - Final: Wrote **docs/planning/plan-test-coverage-40.md** (Phased: Phase1 stabilize ~20-25%, Phase2 core to 40%, Phase3 CI).
- **No Code Changes Yet**: All planning; no files modified. Builds still fail in p2p/psfs.

## Current Status / What We're Working On
- **Active Plan**: Execute `docs/planning/plan-test-coverage-40.md` (40% target, 3-5 days).
  - **Phase 1 (Stabilize, 1 day; ready to start)**: Fix p2p/uploads.go fmt error; mock app/app_test.go (ports/websites/non-sec).
- **Files/Packages in Focus Now**:
  | Priority | Package/File | Status | Next Action |
  |----------|--------------|--------|-------------|
  | 1 | p2p/uploads.go | Build fail (fmt %w) | Fix `fmt.Printf("%v", err)` |
  | 2 | internal/app/app_test.go | Flaky (~22%) | Add mocks (net.Dial/http.Get/exec); skip sec |
  | 3 | internal/store/store.go | 0% (core metadata) | New table-driven tests (AddFile/ApplyGossipUpdate/SaveLoad/concurrency) |
  | 4 | internal/p2p/{host.go,pubsub.go} | Minimal/0% | Mock libp2p (NewHost/Bootstrap/PublishMetadata) |
  | 5 | internal/api/main_api.go | 0% | httptest endpoints (ListAllFiles/AddOrUpdateFile; mock store/p2p) |
  | Later | cmd/*, config/config.go, main.go | 0% | CLI cobra/golden, parse edges |

## What Needs to Be Done Next
1. **Immediate (Phase 1)**: Fix p2p/uploads.go → verify `go test ./internal/p2p/... -cover`. Mock/fix app tests → baseline 20-25%.
2. **Phase 2**: Implement store tests (fixtures from test/test01/ → mock JSON).
3. **Verify Each Phase**: `go test ./... -cover` → update coverage.html.
4. **Phase 3**: CI workflow update (fail <40%).
5. **Ongoing**: User approval per phase; questions on testify/E2E prefs.

**Continue Here**: Respond to "Ready to execute Phase 1? Start with p2p/uploads.go fix?" or user directives. Reference plan MD for details. Maintain ignores; aim mocks/table-driven.

*Copy-paste this entire markdown as the first message in a new session to resume seamlessly.*