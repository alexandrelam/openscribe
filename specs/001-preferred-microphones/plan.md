# Implementation Plan: Preferred Microphones Fallback

**Branch**: `001-preferred-microphones` | **Date**: 2025-10-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-preferred-microphones/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature adds automatic microphone fallback capability to OpenScribe. Users can configure an ordered list of preferred microphone names in the application configuration. The system will automatically select the first available device from this list when the application starts or when devices are connected/disconnected. If the preferred list is empty or no preferred microphones are available, the application falls back to the system default microphone. This eliminates manual reconfiguration when users frequently switch between multiple audio input devices (e.g., built-in mic, USB headset, studio microphone).

## Technical Context

**Language/Version**: Go 1.25.2 (macOS ARM64 only - Apple Silicon M1/M2/M3/M4)
**Primary Dependencies**: malgo v0.11.24 (audio), cobra v1.10.1 (CLI), gopkg.in/yaml.v3 (config)
**Storage**: YAML configuration files at `~/Library/Application Support/openscribe/config.yaml`
**Testing**: Go standard `testing` package - unit tests (CI) + integration tests with build tags (local only)
**Target Platform**: macOS ARM64 (Apple Silicon) - CLI application with daemon capabilities
**Project Type**: Single project (monolithic CLI with internal packages)
**Performance Goals**: Device selection within 2 seconds of device connection/disconnection
**Constraints**: Must not interrupt active recording sessions; case-insensitive string matching only (no fuzzy matching)
**Scale/Scope**: Single-user local application; ~5-10 microphone devices maximum in preferences list; minimal performance impact on audio pipeline

**Current Architecture**:
- Audio device enumeration: `internal/audio/devices.go` using malgo library
- Configuration: `internal/config/config.go` with YAML persistence
- Current selection: Single `Microphone` string field (device name or empty for default)
- CLI framework: Cobra commands in `cmd/openscribe/` and `internal/cli/`

**Extension Points**:
- Config struct can be extended with `PreferredMicrophones []string` field
- `FindMicrophoneByName()` exists for exact name matching
- `GetDefaultMicrophone()` provides fallback logic
- Device enumeration via `ListMicrophones()` returns all available devices

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: ✅ PASS (No constitution file exists yet - using OpenScribe codebase patterns)

### Code Quality Gates

| Gate | Status | Notes |
|------|--------|-------|
| **Testing Required** | ✅ PASS | Unit tests for config parsing + device selection logic; Integration tests for real device enumeration (local only) |
| **Backwards Compatibility** | ✅ PASS | New `PreferredMicrophones` field is optional; existing `Microphone` field continues to work; graceful fallback if preferences empty |
| **Error Handling** | ✅ PASS | Will use existing error patterns (detailed messages with context); non-matching device names handled gracefully |
| **Performance** | ✅ PASS | Device enumeration already exists; adding ordered list iteration adds negligible overhead (<1ms) |
| **Platform Support** | ✅ PASS | macOS-only feature (aligns with project constraint - Apple Silicon only) |
| **Simplicity** | ✅ PASS | Extends existing config pattern; reuses existing device enumeration functions; no new dependencies |

### Design Principles Compliance

- **Single Responsibility**: Config module handles preferences storage; Audio module handles device selection
- **Dependency Injection**: Existing `DeviceEnumerator` interface pattern supports testability
- **YAML Configuration**: Follows existing convention with clear field naming
- **CLI Integration**: New commands follow Cobra pattern established in codebase
- **No Breaking Changes**: Optional feature with graceful degradation

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/openscribe/
└── main.go                          # Application entry point (no changes)

internal/
├── audio/
│   ├── devices.go                   # [MODIFY] Add SelectPreferredMicrophone() function
│   ├── devices_test.go              # [MODIFY] Add integration tests for preference fallback
│   ├── devices_unit_test.go         # [MODIFY] Add unit tests for preference logic
│   └── recorder.go                  # [NO CHANGE] Recording logic
│
├── config/
│   ├── config.go                    # [MODIFY] Add PreferredMicrophones []string field
│   └── config_test.go               # [MODIFY] Add tests for preferences validation
│
└── cli/
    ├── config.go                    # [MODIFY] Add commands: list-preferences, add-preference, remove-preference, clear-preferences
    └── config_test.go               # [NEW] Add CLI command tests

tests/
├── integration/                     # Use build tags: //go:build integration
│   └── device_selection_test.go     # [NEW] End-to-end tests with real devices
└── fixtures/
    └── test_config.yaml             # [NEW] Sample configs for testing
```

**Structure Decision**: Single project structure (monolithic CLI application). This feature modifies existing `internal/audio` and `internal/config` packages and extends `internal/cli` with new commands for preference management. No new top-level directories required. All tests follow existing patterns: unit tests in `*_test.go` files, integration tests with build tags.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

**No violations** - This feature has no complexity concerns. All gates passed.

---

## Post-Design Constitution Check

*Re-evaluation after Phase 1 design completion*

**Date**: 2025-10-19
**Status**: ✅ PASS - All design artifacts conform to quality standards

### Design Quality Review

| Aspect | Status | Notes |
|--------|--------|-------|
| **Data Model** | ✅ PASS | Simple additive change to Config struct; clear validation rules; backward compatible migration |
| **API Contracts** | ✅ PASS | CLI commands follow Cobra patterns; intuitive command naming; comprehensive error handling |
| **Testing Strategy** | ✅ PASS | Unit tests for logic; integration tests with build tags; clear test scenarios documented |
| **User Experience** | ✅ PASS | Quickstart guide is clear and practical; covers common use cases; helpful troubleshooting section |
| **No New Dependencies** | ✅ PASS | Uses only existing libraries (malgo, cobra, yaml.v3); no external packages added |
| **Documentation** | ✅ PASS | Comprehensive research, data model, contracts, and quickstart; ready for implementation |

### Design Concerns Addressed

1. **Backward Compatibility**: ✅ Confirmed via migration strategy in data-model.md
2. **Performance Impact**: ✅ Minimal (O(P*D) selection, ~30ms polling) - acceptable
3. **Error Handling**: ✅ Comprehensive error scenarios documented in contracts
4. **Testing Approach**: ✅ Clear separation of unit/integration tests with build tags

### Ready for Implementation

All Phase 0 (Research) and Phase 1 (Design) artifacts are complete and meet quality standards. The implementation can proceed to Phase 2 (Task Generation) via `/speckit.tasks` command.

**Generated Artifacts**:
- ✅ `plan.md` - Complete technical context and structure
- ✅ `research.md` - 5 key decisions with rationale
- ✅ `data-model.md` - Config entity, Device entity, validation rules, selection algorithm
- ✅ `contracts/cli-commands.md` - 6 CLI commands with full specifications
- ✅ `quickstart.md` - User guide with examples and troubleshooting

**No blockers identified.** Feature design is sound and ready for task breakdown.

