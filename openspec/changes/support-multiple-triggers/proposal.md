# Proposal: Support Multiple Triggers

**Change ID:** `support-multiple-triggers`
**Status:** Draft
**Created:** 2025-11-02

## Why

Users need multiple ways to activate OpenScribe recording beyond a single keyboard hotkey. They want to use mouse buttons (Forward/Back side buttons) alongside or instead of keyboard triggers for more flexible, ergonomic control across different workflows and hardware setups.

## What Changes

- Convert single `hotkey: string` config field to `triggers: []string` array
- Add mouse button support (Forward Button, Back Button) to hotkey detection system
- Implement automatic migration from legacy `hotkey` field to `triggers` array
- Extend config validation to check trigger duplicates and validity
- Update CLI startup message to display all configured triggers
- **BREAKING**: `hotkey` field deprecated (auto-migrated, not removed)

## Impact

- **Affected specs:** New capability `trigger-configuration`
- **Affected code:**
  - `internal/config/config.go` (schema, migration, validation)
  - `internal/hotkey/hotkey.go` (multi-listener wrapper)
  - `internal/hotkey/hotkey_darwin.go` (mouse event detection)
  - `internal/cli/start.go` (multi-trigger initialization)

## User Impact

**Positive:**
- Greater flexibility in how users activate recording
- Better accessibility for users with different hardware setups
- Support for specialized input devices (gaming mice, custom keyboards)

**Risks:**
- Potential confusion between legacy `hotkey` and new `triggers` config
- Need to test across different mouse/keyboard combinations
- Accessibility permissions apply to both keyboard and mouse

## Alternatives Considered

1. **Keep single trigger only** - Rejected: Doesn't address user needs for multiple input methods
2. **Add mouse-only separate config** - Rejected: Creates duplicate trigger detection logic
3. **Use string parsing (e.g., "Right Option OR Forward Button")** - Rejected: Less flexible, harder to validate

## Success Criteria

- [ ] Users can configure multiple keyboard and mouse triggers
- [ ] Existing single-hotkey configs automatically migrate
- [ ] All trigger types support double-press detection
- [ ] Config validation prevents duplicate or invalid triggers
- [ ] Documentation clearly explains trigger configuration

## Related Changes

None.

## References

- Current hotkey implementation: `internal/hotkey/hotkey.go`
- Current config schema: `internal/config/config.go`
- macOS CGEvent API: kCGEventOtherMouseDown for mouse buttons
