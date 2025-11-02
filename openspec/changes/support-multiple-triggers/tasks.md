# Tasks: Support Multiple Triggers

**Change ID:** support-multiple-triggers

This document outlines the implementation tasks for adding multiple trigger support to OpenScribe.

---

## Phase 1: Config Schema & Migration (Foundation)

### 1.1 Update Config struct with Triggers array

**File:** `internal/config/config.go`

- Add `Triggers []string` field to `Config` struct with YAML tag
- Mark `Hotkey string` field as deprecated with comment
- Update `DefaultConfig()` to use `Triggers: []string{"Right Option"}` instead of `Hotkey`

**Validation:** Config compiles, tests pass

---

### 1.2 Implement config migration logic

**File:** `internal/config/config.go`

- Extend `migrate()` function to handle `Hotkey` → `Triggers` conversion
- If `Triggers` is empty and `Hotkey` is set, migrate: `Triggers = []string{Hotkey}`
- Log migration message: `"Migrated legacy 'hotkey' to 'triggers': [value]"`
- Save migrated config to disk

**Validation:** Unit test for migration with legacy config file

---

### 1.3 Update config validation for Triggers array

**File:** `internal/config/config.go`

- Extend `Validate()` to validate `Triggers` array
- Check for empty triggers: `if trimmed == "" { return error }`
- Check for duplicate triggers (case-insensitive)
- Validate each trigger name against valid list (keyboard + mouse)
- If both `Hotkey` and `Triggers` are set, prefer `Triggers` and log warning

**Validation:** Unit tests for:
- Empty trigger
- Duplicate triggers
- Invalid trigger names
- Valid multi-trigger config

---

### 1.4 Update Config.String() display method

**File:** `internal/config/config.go`

- Update `String()` method to display triggers as numbered list
- Format: `Triggers: \n    1. Right Option\n    2. Forward Button`
- Show "(legacy)" label if old `Hotkey` field still present

**Validation:** Run `openscribe config --show` and verify output format

---

## Phase 2: Mouse Button Support (Hotkey Package)

### 2.1 Add mouse button constants to hotkey package

**File:** `internal/hotkey/hotkey.go`

- Add mouse button codes as constants:
  - `ButtonForward KeyCode = 0x10001` (synthetic, needs platform mapping)
  - `ButtonBack KeyCode = 0x10002` (synthetic, needs platform mapping)
- Update `KeyNameMap` to include:
  - `"Forward Button": ButtonForward`
  - `"Back Button": ButtonBack`
- Update `KeyCodeToName` reverse map

**Validation:** Code compiles, constants defined

---

### 2.2 Extend macOS hotkey detection for mouse buttons

**File:** `internal/hotkey/hotkey_darwin.go`

- Update C event tap callback to monitor `kCGEventOtherMouseDown` events
- Add mouse button detection logic:
  ```c
  if (type == kCGEventOtherMouseDown) {
      int64_t buttonNumber = CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
      // Button 3 = Back, Button 4 = Forward
  }
  ```
- Map mouse button numbers to KeyCode equivalents
- Call `goHotkeyCallback()` on mouse button press

**Validation:** Manual test with gaming mouse, verify Forward/Back detection

---

### 2.3 Update event mask to include mouse events

**File:** `internal/hotkey/hotkey_darwin.go`

- Update `registerHotkey()` event mask:
  ```c
  CGEventMask eventMask = CGEventMaskBit(kCGEventFlagsChanged) |
                          CGEventMaskBit(kCGEventOtherMouseDown);
  ```
- Ensure event tap processes both keyboard and mouse events

**Validation:** Test keyboard + mouse triggers simultaneously

---

## Phase 3: Multi-Listener Support

### 3.1 Refactor Listener to support multiple triggers

**File:** `internal/hotkey/hotkey.go`

- Create `MultiListener` struct that manages multiple `Listener` instances
- Implement `NewMultiListener(triggerNames []string, callback func()) (*MultiListener, error)`
- Each trigger gets its own `Listener` with shared callback
- Implement `Start()` to start all listeners
- Implement `Stop()` to stop all listeners

**Validation:** Unit test with multiple triggers (mock keyboard events)

---

### 3.2 Update CLI start command to use multi-trigger

**File:** `internal/cli/start.go`

- Replace `hotkey.NewListener(cfg.Hotkey, ...)` with `hotkey.NewMultiListener(cfg.Triggers, ...)`
- Update startup display message to show all triggers:
  ```go
  fmt.Printf("  Triggers:        %s (double-press)\n", strings.Join(cfg.Triggers, ", "))
  ```
- Handle listener creation errors for any trigger

**Validation:**
- Run `openscribe start` with single trigger
- Run `openscribe start` with multiple triggers
- Verify both keyboard and mouse triggers work

---

## Phase 4: Testing & Documentation

### 4.1 Write comprehensive unit tests

**Files:**
- `internal/config/config_test.go`
- `internal/hotkey/hotkey_test.go`

- Test config migration from `Hotkey` → `Triggers`
- Test config validation (empty, duplicates, invalid names)
- Test multi-listener start/stop
- Test mouse button mapping

**Validation:** Run `make test` or `go test ./...`

---

### 4.2 Update test fixtures

**File:** `tests/fixtures/test_config.yaml`

- Add example multi-trigger config:
  ```yaml
  triggers:
    - "Right Option"
    - "Forward Button"
  ```

**Validation:** Tests use updated fixture

---

### 4.3 Update README.md

**File:** `README.md`

- Add section explaining trigger configuration
- Document supported mouse buttons (Forward, Back)
- Show example multi-trigger YAML config
- Add troubleshooting note for mouse button detection

**Validation:** Review README for clarity

---

### 4.4 Manual end-to-end testing

**Scenarios:**
1. Fresh install: Verify default config uses `triggers: ["Right Option"]`
2. Legacy migration: Create config with `hotkey: "Right Shift"`, verify migration
3. Multiple keyboard triggers: Configure 2 keyboard keys, test both
4. Mouse button trigger: Configure `Forward Button`, test with gaming mouse
5. Mixed triggers: Configure keyboard + mouse, test both methods
6. Invalid config: Test validation errors (duplicate, empty, invalid)

**Validation:** All scenarios pass without errors

---

## Dependencies & Parallelization

**Can be done in parallel:**
- Phase 1 (Config) and Phase 2 (Mouse support) are independent
- Documentation (4.3) can be drafted early

**Must be sequential:**
- Phase 3 depends on Phase 1 and Phase 2 completing
- Testing (4.1, 4.2) depends on Phase 3
- E2E testing (4.4) is final validation

---

## Rollback Plan

If issues arise:
1. Keep legacy `Hotkey` field functional as fallback
2. Add feature flag in config: `use_legacy_hotkey: true`
3. Revert CLI to use single listener if multi-listener fails

---

## Success Metrics

- [ ] Config migration works for 100% of legacy configs
- [ ] Both keyboard and mouse triggers are detected reliably
- [ ] Multiple triggers can activate recording without conflicts
- [ ] Validation prevents all invalid configurations
- [ ] Zero regressions in existing single-trigger functionality
- [ ] Documentation clearly explains new trigger system
