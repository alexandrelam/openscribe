# Tasks: Preferred Microphones Fallback

**Feature Branch**: `001-preferred-microphones`
**Input**: Design documents from `/specs/001-preferred-microphones/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-commands.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Tests**: Unit tests are required; integration tests with build tags for local testing only.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and testing infrastructure

- [X] T001 [P] Create test fixtures directory structure at `tests/fixtures/`
- [X] T002 [P] Create integration tests directory structure at `tests/integration/`
- [X] T003 [P] Add sample test configuration files in `tests/fixtures/test_config.yaml`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data model changes that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Add `PreferredMicrophones []string` field to Config struct in `internal/config/config.go`
- [X] T005 Implement config validation for `PreferredMicrophones` in `internal/config/config.go`
- [X] T006 Implement auto-migration from `Microphone` to `PreferredMicrophones` in `internal/config/config.go`
- [X] T007 Update Config `Load()` function to call migration in `internal/config/config.go`
- [X] T008 Add unit tests for config parsing with `PreferredMicrophones` in `internal/config/config_test.go`
- [X] T009 Add unit tests for config validation (empty strings, duplicates) in `internal/config/config_test.go`
- [X] T010 Add unit tests for auto-migration logic in `internal/config/config_test.go`

**Checkpoint**: Configuration foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Automatic Microphone Selection from Preferences (Priority: P1) 🎯 MVP

**Goal**: Enable automatic selection of preferred microphones at application startup and between recording sessions. Users should be able to configure a preference list and have the system automatically select the first available device.

**Independent Test**: Configure a preference list, plug/unplug microphones in different orders, and verify the application selects the highest-priority available device. Can be tested without User Story 2 (CLI commands) by manually editing the config YAML file.

### Unit Tests for User Story 1

- [X] T011 [P] [US1] Add unit tests for device selection with preferences in `internal/audio/devices_unit_test.go`
- [X] T012 [P] [US1] Add unit tests for device selection fallback logic in `internal/audio/devices_unit_test.go`
- [X] T013 [P] [US1] Add unit tests for case-insensitive device name matching in `internal/audio/devices_unit_test.go`

### Implementation for User Story 1

- [X] T014 [US1] Implement `SelectMicrophone(cfg *config.Config)` function in `internal/audio/devices.go`
- [X] T015 [US1] Update device selection logic to prioritize preferred microphones in `internal/audio/devices.go`
- [X] T016 [US1] Add case-insensitive exact matching using `strings.EqualFold()` in `internal/audio/devices.go`
- [X] T017 [US1] Implement fallback to default microphone when no preferences match in `internal/audio/devices.go`
- [X] T018 [US1] Add logging for device selection process (selected device, from preferences flag) in `internal/audio/devices.go`
- [X] T019 [US1] Add logging for preference iteration (tried device, match/no match) in `internal/audio/devices.go`

### Integration Tests for User Story 1

- [ ] T020 [US1] Add integration test for real device enumeration with `//go:build integration` tag in `tests/integration/device_selection_test.go`
- [ ] T021 [US1] Add integration test for preference selection with real devices in `tests/integration/device_selection_test.go`
- [ ] T022 [US1] Add integration test for fallback behavior with real devices in `tests/integration/device_selection_test.go`

**Checkpoint**: At this point, User Story 1 should be fully functional. The application can automatically select preferred microphones when loading config or starting recording, even without CLI commands (users can manually edit config YAML).

---

## Phase 4: User Story 2 - Configure Preferred Microphone List (Priority: P1)

**Goal**: Provide CLI commands for users to view and edit their preferred microphone list. This makes the feature accessible without manually editing YAML files.

**Independent Test**: Use CLI commands to add/remove/reorder microphones, verify changes persist in config file, and verify User Story 1 (automatic selection) uses the updated preferences. Can be tested independently by running commands and checking the config file.

### CLI Command Implementation for User Story 2

- [X] T023 [P] [US2] Implement `list-microphones` command in `internal/cli/config.go`
- [X] T024 [P] [US2] Implement `show-preferences` command in `internal/cli/config.go`
- [X] T025 [P] [US2] Implement `add-preference <name>` command in `internal/cli/config.go`
- [X] T026 [P] [US2] Implement `remove-preference <name-or-index>` command in `internal/cli/config.go`
- [X] T027 [P] [US2] Implement `clear-preferences` command in `internal/cli/config.go`

### CLI Command Validation and Error Handling for User Story 2

- [X] T028 [US2] Add validation for empty device names in `add-preference` command in `internal/cli/config.go`
- [X] T029 [US2] Add duplicate detection (case-insensitive) in `add-preference` command in `internal/cli/config.go`
- [X] T030 [US2] Add warning when adding non-connected device in `add-preference` command in `internal/cli/config.go`
- [X] T031 [US2] Add validation for invalid indices in `remove-preference` command in `internal/cli/config.go`
- [X] T032 [US2] Add support for both name and index in `remove-preference` command in `internal/cli/config.go`

### CLI Command Tests for User Story 2

- [ ] T033 [P] [US2] Add unit tests for `add-preference` command in `internal/cli/config_test.go`
- [ ] T034 [P] [US2] Add unit tests for `remove-preference` command in `internal/cli/config_test.go`
- [ ] T035 [P] [US2] Add unit tests for `clear-preferences` command in `internal/cli/config_test.go`
- [ ] T036 [P] [US2] Add unit tests for argument parsing (name vs index) in `internal/cli/config_test.go`

### Integration with Existing Config Command for User Story 2

- [X] T037 [US2] Update `openscribe config` (no args) output to display `preferred_microphones` field in `internal/cli/config.go`
- [ ] T038 [US2] Add integration test for config display with preferences in `tests/integration/device_selection_test.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work. Users can configure preferences via CLI commands and see automatic device selection in action.

---

## Phase 5: User Story 3 - View Current Active Microphone and Selection Status (Priority: P2)

**Goal**: Provide visibility into which microphone is currently active and whether it was selected from preferences or is a fallback. This helps users verify the system is working correctly.

**Independent Test**: Connect different microphones, run status command, and verify it shows the correct active device and selection method. Can be tested independently of other stories by checking device status at any time.

### Implementation for User Story 3

- [ ] T039 [P] [US3] Add `CurrentMicrophone` field to track active device in appropriate struct (determine location based on architecture)
- [ ] T040 [P] [US3] Add `SelectedFromPreferences` boolean flag to track selection source
- [ ] T041 [US3] Update device selection logic to set `CurrentMicrophone` and `SelectedFromPreferences` flag
- [ ] T042 [US3] Implement `show-status` or `get-active-microphone` CLI command in `internal/cli/config.go`
- [ ] T043 [US3] Add status output formatting (device name, selection source, fallback indicator) in `internal/cli/config.go`

### Tests for User Story 3

- [ ] T044 [P] [US3] Add unit tests for status command output formatting in `internal/cli/config_test.go`
- [ ] T045 [US3] Add integration test for status command with real device selection in `tests/integration/device_selection_test.go`

**Checkpoint**: All user stories should now be independently functional. Users have full visibility and control over microphone preferences and selection.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T046 [P] Update README.md with usage examples for new CLI commands
- [ ] T047 [P] Add comprehensive error messages for all device enumeration failures
- [ ] T048 [P] Add comprehensive error messages for permission denied scenarios
- [X] T049 Code review and refactoring for consistency with existing codebase patterns
- [X] T050 Run `make fmt` to format all Go code
- [X] T051 Run `make lint` to check code quality with golangci-lint
- [X] T052 Run `make test` to verify all unit tests pass
- [ ] T053 Manually validate quickstart.md scenarios with real devices
- [ ] T054 Test backward compatibility with existing config files (legacy `microphone` field)
- [ ] T055 Test all edge cases from spec.md (duplicate entries, device name typos, rapid connect/disconnect)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Stories 1 and 2 can proceed in parallel after Phase 2
  - User Story 3 can proceed in parallel after Phase 2
  - User Story 2 enhances User Story 1 but both are independently functional
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Automatic Selection**: Can start after Foundational (Phase 2) - No dependencies on other stories. Core feature functionality.
- **User Story 2 (P1) - Configure List**: Can start after Foundational (Phase 2) - Enhances User Story 1 but both work independently. Users can manually edit YAML without this.
- **User Story 3 (P2) - View Status**: Can start after Foundational (Phase 2) - Provides visibility into User Story 1 results but doesn't affect functionality.

### Within Each User Story

- **User Story 1**: Unit tests → Core implementation → Integration tests → Logging
- **User Story 2**: CLI command structure → Validation → Error handling → Tests
- **User Story 3**: Data tracking → Status command → Output formatting → Tests

### Parallel Opportunities

**Phase 1 (Setup)**: All 3 tasks can run in parallel (different directories)

**Phase 2 (Foundational)**: Tasks T008, T009, T010 can run in parallel after T004-T007 complete (test files vs implementation files)

**Phase 3 (User Story 1)**:
- Tasks T011, T012, T013 can run in parallel (different test scenarios in same file)
- Tasks T018, T019 can run in parallel with T020-T022 (different files)

**Phase 4 (User Story 2)**:
- Tasks T023, T024, T025, T026, T027 can run in parallel (different command functions)
- Tasks T033, T034, T035, T036 can run in parallel (different test functions)

**Phase 5 (User Story 3)**:
- Tasks T039, T040 can run in parallel (different fields/locations)
- Tasks T044, T045 can run in parallel after T039-T043 complete (test files vs implementation)

**Phase 6 (Polish)**:
- Tasks T046, T047, T048 can run in parallel (different files)

---

## Parallel Example: User Story 1

```bash
# Launch all unit tests for User Story 1 together:
Task: "Add unit tests for device selection with preferences in internal/audio/devices_unit_test.go"
Task: "Add unit tests for device selection fallback logic in internal/audio/devices_unit_test.go"
Task: "Add unit tests for case-insensitive device name matching in internal/audio/devices_unit_test.go"

# After implementation is complete, launch integration tests and logging together:
Task: "Add logging for device selection process in internal/audio/devices.go"
Task: "Add logging for preference iteration in internal/audio/devices.go"
Task: "Add integration test for real device enumeration in tests/integration/device_selection_test.go"
```

## Parallel Example: User Story 2

```bash
# Launch all CLI command implementations together:
Task: "Implement list-microphones command in internal/cli/config.go"
Task: "Implement show-preferences command in internal/cli/config.go"
Task: "Implement add-preference command in internal/cli/config.go"
Task: "Implement remove-preference command in internal/cli/config.go"
Task: "Implement clear-preferences command in internal/cli/config.go"

# Launch all CLI command tests together:
Task: "Add unit tests for add-preference command in internal/cli/config_test.go"
Task: "Add unit tests for remove-preference command in internal/cli/config_test.go"
Task: "Add unit tests for clear-preferences command in internal/cli/config_test.go"
Task: "Add unit tests for argument parsing in internal/cli/config_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (test infrastructure)
2. Complete Phase 2: Foundational (config data model) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (automatic selection)
4. Complete Phase 4: User Story 2 (CLI commands)
5. **STOP and VALIDATE**: Test both stories independently
   - Test 1: Manually edit YAML, verify automatic selection works
   - Test 2: Use CLI commands, verify config updates, verify automatic selection uses new config
6. Deploy/demo if ready

### Incremental Delivery

1. **Foundation** (Phase 1 + 2): Config model ready
2. **MVP** (Phase 3 + 4): User Stories 1 + 2 → Test independently → Deploy/Demo
3. **Enhanced** (Phase 5): User Story 3 → Test independently → Deploy/Demo
4. **Polished** (Phase 6): Code quality and edge cases → Deploy/Demo

### Parallel Team Strategy

With 2 developers:

1. Both developers complete Setup + Foundational together (Phase 1 + 2)
2. Once Foundational is done:
   - Developer A: User Story 1 (automatic selection)
   - Developer B: User Story 2 (CLI commands)
3. Both stories complete and integrate (US2 uses US1's selection logic)
4. One developer takes User Story 3 while other does testing/polish

---

## Testing Strategy

### Unit Tests
- Config parsing with `PreferredMicrophones` array
- Config validation (empty strings, duplicates)
- Auto-migration from `Microphone` to `PreferredMicrophones`
- Device selection logic with mocked device lists
- Case-insensitive matching
- Fallback behavior
- CLI command argument parsing
- CLI command validation

### Integration Tests (with `//go:build integration` tag)
- Real device enumeration on test machine
- Device selection with real devices
- Preference fallback with real devices
- Config file persistence
- CLI commands modifying config file

### Manual Tests (per quickstart.md)
- Physical device connect/disconnect scenarios
- Multiple microphones in different orders
- Edge cases: duplicate names, non-matching names, rapid connect/disconnect
- Migration from old config format

---

## Notes

- [P] tasks = different files or independent functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Integration tests use `//go:build integration` tag (local only, not CI)
- Unit tests run in CI for every commit
- Commit after each task or logical group
- Stop at checkpoints to validate story independently
- Use exact device names from `ListMicrophones()` for configuration
- Case-insensitive exact matching only (no fuzzy matching)
- Empty `PreferredMicrophones` array means use system default
- Legacy `Microphone` field preserved for backward compatibility

---

## Success Metrics (from spec.md)

- ✅ **SC-001**: Users can configure preferences in under 1 minute (via CLI commands)
- ✅ **SC-002**: Device selection happens within 2 seconds when device connected
- ✅ **SC-003**: Correct highest-priority device selected in 100% of test scenarios
- ✅ **SC-004**: 80% reduction in manual configuration time for multi-mic users
- ✅ **SC-005**: 100% persistence and restoration of preferences across restarts
