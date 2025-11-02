# Spec: Trigger Configuration

**Capability:** trigger-configuration
**Related Change:** support-multiple-triggers

## Overview

This spec defines the requirements for configuring multiple triggers (keyboard keys and mouse buttons) to activate OpenScribe recording. It replaces the single-hotkey system with a flexible multi-trigger system while maintaining backward compatibility.

## ADDED Requirements

### Requirement: System SHALL support multiple triggers configuration

The system SHALL support configuration of multiple triggers through a `triggers` array in the configuration file, where each trigger specifies an input device and double-press detection.

**ID:** TRIG-001
**Rationale:** Users need flexibility to activate recording from different input devices (keyboard, mouse) based on their workflow and hardware setup.

#### Scenario: User configures multiple triggers via YAML

**Given** a user has OpenScribe installed
**When** they edit `~/.config/openscribe/config.yaml` and add:
```yaml
triggers:
  - "Right Option"
  - "Forward Button"
```
**Then** OpenScribe accepts both triggers as valid
**And** double-pressing either Right Option OR Forward Button activates recording

#### Scenario: User configures single trigger (backward compatibility)

**Given** a user has an existing config with:
```yaml
hotkey: "Right Option"
```
**When** OpenScribe loads the configuration
**Then** the system automatically migrates `hotkey` to `triggers: ["Right Option"]`
**And** the migrated config is saved to disk
**And** the user sees a migration message in logs

---

### Requirement: System SHALL support mouse button detection

The system SHALL detect and support mouse button presses for trigger activation, including Forward and Back side buttons commonly found on gaming and productivity mice.

**ID:** TRIG-002
**Rationale:** Users with mice that have programmable side buttons want to use these as triggers without needing keyboard access.

#### Scenario: User triggers recording with Forward mouse button

**Given** the config contains `triggers: ["Forward Button"]`
**When** the user double-presses the Forward mouse button within 500ms
**Then** OpenScribe starts/stops recording
**And** the system displays "🔴 Recording started..." message

#### Scenario: User triggers recording with Back mouse button

**Given** the config contains `triggers: ["Back Button"]`
**When** the user double-presses the Back mouse button within 500ms
**Then** OpenScribe starts/stops recording
**And** the system displays "🔴 Recording started..." message

---

### Requirement: System SHALL validate trigger configuration

The system SHALL validate trigger configuration on load, rejecting duplicate triggers, unknown trigger names, and empty trigger values.

**ID:** TRIG-003
**Rationale:** Prevent configuration errors that could cause unexpected behavior or system failures.

#### Scenario: User specifies duplicate triggers

**Given** a user edits config with:
```yaml
triggers:
  - "Right Option"
  - "Right Option"
```
**When** OpenScribe loads the configuration
**Then** validation fails with error "duplicate trigger: Right Option"
**And** OpenScribe exits with status code 1

#### Scenario: User specifies invalid trigger name

**Given** a user edits config with:
```yaml
triggers:
  - "Invalid Button"
```
**When** OpenScribe loads the configuration
**Then** validation fails with error listing valid trigger options
**And** OpenScribe exits with status code 1

#### Scenario: User specifies empty trigger

**Given** a user edits config with:
```yaml
triggers:
  - ""
```
**When** OpenScribe loads the configuration
**Then** validation fails with error "trigger cannot be empty"
**And** OpenScribe exits with status code 1

---

### Requirement: System SHALL display active triggers in CLI

The system SHALL display all configured triggers in the startup message and config output, clearly showing which input methods are active.

**ID:** TRIG-004
**Rationale:** Users need visibility into what triggers are configured, especially when using multiple triggers.

#### Scenario: Display multiple triggers on startup

**Given** the config contains:
```yaml
triggers:
  - "Right Option"
  - "Forward Button"
```
**When** the user runs `openscribe start`
**Then** the startup message displays:
```
Triggers:        Right Option (double-press), Forward Button (double-press)
```

#### Scenario: Display triggers in config output

**Given** the config contains multiple triggers
**When** the user runs `openscribe config --show`
**Then** the output displays:
```
Settings:
  Triggers:
    1. Right Option
    2. Forward Button
```

---

## MODIFIED Requirements

### Requirement: System SHALL support config schema backward compatibility

The system SHALL support both legacy `hotkey` field and new `triggers` array, with automatic migration from legacy format.

**ID:** TRIG-005 (MODIFIED)

**Changes:**
- **Before:** Only `hotkey: string` field existed
- **After:** Both `hotkey: string` (deprecated) and `triggers: []string` exist
- Migration automatically converts `hotkey` → `triggers` on first load

#### Scenario: Legacy config auto-migration

**Given** an existing config file contains only:
```yaml
hotkey: "Right Shift"
```
**When** OpenScribe loads the config
**Then** the in-memory config has `triggers: ["Right Shift"]`
**And** the saved config file is updated with `triggers: ["Right Shift"]`
**And** the `hotkey` field is removed or marked deprecated

---

## REMOVED Requirements

None. This change is purely additive with backward compatibility.

---

## Implementation Notes

### Supported Trigger Names

The following trigger names MUST be recognized:

**Keyboard (existing):**
- "Left Option", "Right Option"
- "Left Shift", "Right Shift"
- "Left Command", "Right Command"
- "Left Control", "Right Control"

**Mouse (new):**
- "Forward Button" (mouse button 4, typically side button)
- "Back Button" (mouse button 3, typically side button)

### Technical Approach

1. **Config Schema:** Add `Triggers []string` field to `Config` struct
2. **Migration:** Implement `migrate()` logic to convert `Hotkey` → `Triggers`
3. **Validation:** Extend `Validate()` to check trigger array for duplicates/validity
4. **Hotkey Package:** Extend `hotkey.Listener` to support mouse event monitoring
5. **CGEvent Types:** Use `kCGEventOtherMouseDown` for mouse buttons 3-31

### Platform Requirements

- macOS 10.10+ (for CGEvent mouse button support)
- Accessibility permissions (required for both keyboard and mouse monitoring)

---

## Testing Strategy

- Unit tests for config validation (duplicates, empty, invalid)
- Unit tests for migration logic (hotkey → triggers)
- Integration tests for keyboard trigger detection
- Integration tests for mouse button trigger detection
- Manual testing with physical gaming mouse (Forward/Back buttons)

---

## Dependencies

- Related capability: None (standalone change)
- External dependencies: macOS CGEvent API for mouse button detection
