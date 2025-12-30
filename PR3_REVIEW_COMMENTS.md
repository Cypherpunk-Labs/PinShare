# PR #3 Review Comments

**Title:** [Feat] Windows Service / Tray / Installer
**URL:** https://github.com/Cypherpunk-Labs/PinShare/pull/3
**State:** OPEN

**Last Updated:** 2025-12-29

---

## Completed Items

Items that have been addressed in the codebase:

### Code Changes

| # | Original Item | Status | Notes |
|---|---------------|--------|-------|
| 1 | Stop service on tray exit, start on tray start | ✅ Done | Tray now manages service lifecycle |
| 2 | Parallelize file processing (uploads.go) | ✅ Done | Semaphore pattern with `maxConcurrentUploads=4` |
| 4 | Rename mutex to reflect protected variables | ✅ Done | `processMu` with comment at line 43-44 |
| 5 | Rename `GetIPFSRepoPath` to `GetIPFSDataPath` | ✅ Done | config.go:306-307 |
| 6-8 | Fix "repository" phrasing in log messages | ✅ Done | No "repository" occurrences in process.go |
| 12-13 | Clarify placeholders in SERVICE.md | ✅ Done | Updated with concrete backup/restore examples |
| 14 | Fix links in SERVICE.md | ✅ Done | Links at lines 519-520 are correct |
| 16 | Move ServiceState constants to winservice | ✅ Done | winservice/constants.go:60-69 |
| 17-22 | Extract magic numbers to constants | ✅ Done | service.go and service_control.go use winservice constants |

### Tray Constants Consolidation (2025-12-29)
| Item | Status | Notes |
|------|--------|-------|
| Remove duplicate `serviceRestartDelay` | ✅ Done | Now uses `winservice.ServiceRestartDelay` |
| Use `ServiceStartTimeout` for 60s timeout | ✅ Done | tray.go:512 |
| Use `ServiceStopTimeout` for 30s timeouts | ✅ Done | tray.go:566, 579 |
| Use `ServicePollInterval` for poll interval | ✅ Done | tray.go:611 |
| Use `HealthCheckPoll` for sleep | ✅ Done | tray.go:666 |

### SecurityCapability Type Refactoring (2025-12-29)
| Item | Status | Notes |
|------|--------|-------|
| Create `internal/types/security.go` | ✅ Done | New shared types package |
| Update config.go to use types.SecurityCapability | ✅ Done | config.go:74 |
| Update p2p/security.go to re-export from types | ✅ Done | Backwards compatible |
| Remove type casts in downloads.go | ✅ Done | Lines 12, 50 |
| Remove type casts in uploads.go | ✅ Done | Lines 116, 195 |

### P2P Code Quality (previously addressed)
| Item | Status | Notes |
|------|--------|-------|
| Use `filepath.Join()` for OS-agnostic paths | ✅ Done | downloads.go:31,52; uploads.go throughout |
| Use `fmt.Printf` instead of string concat | ✅ Done | All print statements updated |
| Switch statements for security capability | ✅ Done | downloads.go:62-82; uploads.go:130-150 |
| SecurityCapability enum constants | ✅ Done | security.go with helper methods |
| Extract `processFileWithLimit` with godoc | ✅ Done | uploads.go:45-54 |
| Rename variables to `filename` | ✅ Done | uploads.go:31 |
| Add concurrency comments | ✅ Done | uploads.go:25-27 |
| Reduce deep nesting | ✅ Done | Helper functions in uploads.go:100-236 |

### Files Consolidated/Removed
| File | Status | Notes |
|------|--------|-------|
| `installer/README-WIX6.md` | N/A | Merged into `installer/README.md` |
| `docs/windows-architecture.md` | N/A | Merged into `docs/windows/SERVICE.md` |

---

## Outstanding Items

Items that still need attention:

| # | File | Description | Thread ID |
|---|------|-------------|-----------|
| 9 | `go.mod` | Go version change - needs discussion | `PRRT_kwDOPFH43M5mVf3O` |
| 10 | `go.mod` | New dependencies - explanation needed | `PRRT_kwDOPFH43M5mVgWQ` |
| 11 | `docs/windows/QUICKSTART.md` | Review "More Info" section (116-131) | `PRRT_kwDOPFH43M5mazBp` |
| 23 | `docs/windows/SERVICE.md` | Add named anchor to section link | `PRRT_kwDOPFH43M5m4Fnb` |

---

## Other Comments (Review Requests, etc.)

These don't require code changes - they're acknowledgments or review requests for others.

| # | File | Description | Thread ID |
|---|------|-------------|-----------|
| 1 | `cmd/pinshare-tray/main.go` | Acknowledged unused function for future UI | `PRRT_kwDOPFH43M5kjoML` |
| 2 | `docs/windows/README.md` | @kempy007 review request (line 213) | `PRRT_kwDOPFH43M5kxtI6` |
| 3 | `docs/windows/README.md` | @kempy007 review request (line 273) | `PRRT_kwDOPFH43M5kxty5` |
| 4 | `docs/windows/TESTING.md` | Not thoroughly reviewed | `PRRT_kwDOPFH43M5kx0s-` |
| 5 | `cmd/pinsharesvc/ui_server.go` | Acknowledged file stays until UI integration | `PRRT_kwDOPFH43M5kx4iE` |
| 6 | `internal/config/config.go` | @kempy007 must review | `PRRT_kwDOPFH43M5k-C8l` |
