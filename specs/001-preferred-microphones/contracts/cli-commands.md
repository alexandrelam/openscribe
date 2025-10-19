# CLI Commands Contract: Preferred Microphones

**Feature**: 001-preferred-microphones
**Date**: 2025-10-19
**Interface Type**: Command-Line Interface (Cobra framework)

---

## Overview

This document defines the CLI command contracts for managing preferred microphone configurations. All commands follow the existing Cobra CLI pattern used in OpenScribe and integrate with the `internal/config` and `internal/audio` packages.

---

## Command Structure

### Base Command

All microphone preference commands are subcommands of the existing `config` command:

```
openscribe config [subcommand]
```

**Existing Behavior**: The `config` command currently displays the full configuration or allows setting individual fields. This feature extends it with microphone preference management subcommands.

---

## Commands

### 1. List All Microphones (Enhanced)

**Command**: `openscribe config list-microphones` or `openscribe config --list-microphones`

**Purpose**: Display all currently available audio input devices with their indices, names, and default status. Helps users discover exact device names for configuration.

**Usage**:
```bash
openscribe config list-microphones
openscribe config --list-microphones
```

**Options**: None

**Output** (stdout):
```
Available Microphones:
  1. MacBook Pro Microphone (default)
  2. Blue Yeti USB Microphone
  3. AirPods Pro

To set preferences:
  openscribe config add-preference "Blue Yeti USB Microphone"
  openscribe config add-preference "AirPods Pro"
```

**Exit Codes**:
- `0`: Success
- `1`: Error enumerating devices (e.g., permission denied, no devices found)

**Error Examples**:
```bash
$ openscribe config list-microphones
Error: No microphones found

Please check:
  1. Microphone is connected
  2. System Preferences > Sound > Input
  3. Microphone permissions granted in System Preferences > Security & Privacy > Privacy > Microphone
```

**Implementation Notes**:
- Calls `audio.ListMicrophones()` to enumerate devices
- Displays in user-friendly numbered format
- Indicates which device is system default
- Shows helper text for setting preferences

---

### 2. Show Current Preferences

**Command**: `openscribe config show-preferences` or `openscribe config --show-preferences`

**Purpose**: Display the current ordered list of preferred microphones from the configuration.

**Usage**:
```bash
openscribe config show-preferences
openscribe config --show-preferences
```

**Options**: None

**Output** (stdout):
```
Preferred Microphones (in priority order):
  1. Blue Yeti USB Microphone
  2. AirPods Pro
  3. MacBook Pro Microphone

Fallback: System default microphone
```

**Output (when no preferences configured)**:
```
No preferred microphones configured.

Using: System default microphone

To add preferences:
  openscribe config add-preference "Microphone Name"
```

**Exit Codes**:
- `0`: Always succeeds (even if no preferences configured)

**Implementation Notes**:
- Reads `config.PreferredMicrophones` array
- Displays in numbered list format
- Shows helpful hint if list is empty

---

### 3. Add Preference

**Command**: `openscribe config add-preference <microphone-name>`

**Purpose**: Add a microphone to the end of the preferences list.

**Usage**:
```bash
openscribe config add-preference "Blue Yeti USB Microphone"
openscribe config add-preference "AirPods Pro"
```

**Arguments**:
- `<microphone-name>` (required): Exact name of the microphone device (case-insensitive)

**Options**: None

**Output** (stdout):
```
✓ Added "Blue Yeti USB Microphone" to preferred microphones (priority 1)
```

**Validation**:
- Device name must not be empty or whitespace-only
- Warning (not error) if device name doesn't match any currently connected devices:
  ```
  ⚠ Warning: "Studio Mic XYZ" is not currently connected
  ✓ Added "Studio Mic XYZ" to preferred microphones (priority 2)
  ```
- Duplicate detection (case-insensitive):
  ```
  Error: "Blue Yeti USB Microphone" is already in your preferences (priority 1)
  ```

**Exit Codes**:
- `0`: Success
- `1`: Invalid input (empty name, duplicate)

**Implementation Notes**:
- Appends to `config.PreferredMicrophones` array
- Calls `config.Validate()` before saving
- Saves updated config to YAML file

---

### 4. Remove Preference

**Command**: `openscribe config remove-preference <microphone-name-or-index>`

**Purpose**: Remove a microphone from the preferences list by name or priority index.

**Usage**:
```bash
# Remove by name
openscribe config remove-preference "Blue Yeti USB Microphone"

# Remove by index (1-based)
openscribe config remove-preference 2
```

**Arguments**:
- `<microphone-name-or-index>` (required): Either the exact device name (case-insensitive) or the priority index (1-based)

**Options**: None

**Output** (stdout):
```
✓ Removed "AirPods Pro" from preferred microphones
```

**Validation**:
- If name provided: must exist in preferences (case-insensitive match)
  ```
  Error: "Studio Mic XYZ" not found in preferences
  ```
- If index provided: must be valid 1-based index
  ```
  Error: Invalid index: 5 (valid range: 1-3)
  ```

**Exit Codes**:
- `0`: Success
- `1`: Device not found or invalid index

**Implementation Notes**:
- Accepts both name and numeric index (detected via `strconv.Atoi`)
- Removes matching entry from `config.PreferredMicrophones`
- Saves updated config to YAML file

---

### 5. Clear All Preferences

**Command**: `openscribe config clear-preferences`

**Purpose**: Remove all microphones from the preferences list (revert to default behavior).

**Usage**:
```bash
openscribe config clear-preferences
```

**Arguments**: None

**Options**: None

**Output** (stdout):
```
✓ Cleared all preferred microphones
  Will now use system default microphone
```

**Exit Codes**:
- `0`: Always succeeds

**Implementation Notes**:
- Sets `config.PreferredMicrophones = []string{}`
- Saves updated config to YAML file

---

### 6. Reorder Preferences (Optional - Future Enhancement)

**Command**: `openscribe config reorder-preferences <from-index> <to-index>`

**Purpose**: Change the priority order of a microphone in the preferences list.

**Usage**:
```bash
# Move item at position 3 to position 1 (highest priority)
openscribe config reorder-preferences 3 1
```

**Arguments**:
- `<from-index>` (required): Current 1-based index
- `<to-index>` (required): Desired 1-based index

**Options**: None

**Output** (stdout):
```
✓ Moved "MacBook Pro Microphone" from position 3 to position 1

Updated preferences:
  1. MacBook Pro Microphone
  2. Blue Yeti USB Microphone
  3. AirPods Pro
```

**Exit Codes**:
- `0`: Success
- `1`: Invalid index

**Implementation Notes**:
- **Status**: OPTIONAL - Not required for MVP
- Can be implemented later if users request it
- Alternative: users can remove and re-add to change order

---

## Integration with Existing Commands

### Existing: `openscribe config` (No Arguments)

**Behavior**: Display full configuration including new `preferred_microphones` field.

**Updated Output**:
```yaml
Config: /Users/username/Library/Application Support/openscribe/config.yaml

model: small
language:
hotkey: Right Option
auto_paste: true
audio_feedback: true
verbose: false
microphone: Blue Yeti USB Microphone  # (legacy - deprecated)
preferred_microphones:
  - Blue Yeti USB Microphone
  - AirPods Pro
  - MacBook Pro Microphone
```

---

### Existing: `openscribe config --set <key> <value>`

**Enhancement**: Support setting `preferred_microphones` as comma-separated list.

**Usage**:
```bash
# Set single preference
openscribe config --set preferred_microphones "Blue Yeti USB Microphone"

# Set multiple preferences (comma-separated)
openscribe config --set preferred_microphones "Blue Yeti,AirPods Pro,MacBook Pro Microphone"
```

**Implementation Notes**:
- If value contains commas, split into array
- Validate each entry
- Save to config

**Alternative**: This might be confusing syntax. Consider deprecating `--set` for preferences and using dedicated subcommands only.

---

## Examples

### Scenario 1: First-Time Setup

```bash
# Step 1: Discover available devices
$ openscribe config list-microphones
Available Microphones:
  1. MacBook Pro Microphone (default)
  2. Blue Yeti USB Microphone

# Step 2: Add preferences in priority order
$ openscribe config add-preference "Blue Yeti USB Microphone"
✓ Added "Blue Yeti USB Microphone" to preferred microphones (priority 1)

$ openscribe config add-preference "MacBook Pro Microphone"
✓ Added "MacBook Pro Microphone" to preferred microphones (priority 2)

# Step 3: Verify configuration
$ openscribe config show-preferences
Preferred Microphones (in priority order):
  1. Blue Yeti USB Microphone
  2. MacBook Pro Microphone

Fallback: System default microphone
```

---

### Scenario 2: Removing a Preference

```bash
# Remove by name
$ openscribe config remove-preference "AirPods Pro"
✓ Removed "AirPods Pro" from preferred microphones

# Remove by index
$ openscribe config remove-preference 2
✓ Removed "MacBook Pro Microphone" from preferred microphones
```

---

### Scenario 3: Reset to Default

```bash
$ openscribe config clear-preferences
✓ Cleared all preferred microphones
  Will now use system default microphone
```

---

## Error Handling

### Common Error Scenarios

| Scenario | Exit Code | Message |
|----------|-----------|---------|
| No microphones found | 1 | "Error: No microphones found. Check System Preferences > Sound > Input." |
| Permission denied | 1 | "Error: Microphone access denied. Grant permissions in System Preferences > Security & Privacy > Microphone." |
| Empty device name | 1 | "Error: Microphone name cannot be empty" |
| Duplicate preference | 1 | "Error: \"Blue Yeti\" is already in your preferences (priority 1)" |
| Device not in preferences | 1 | "Error: \"Studio Mic\" not found in preferences" |
| Invalid index | 1 | "Error: Invalid index: 5 (valid range: 1-3)" |
| Config file locked/inaccessible | 1 | "Error: Failed to save config: [specific error]" |

### Warning Messages (Non-Fatal)

```bash
$ openscribe config add-preference "Fancy Studio Mic"
⚠ Warning: "Fancy Studio Mic" is not currently connected
✓ Added "Fancy Studio Mic" to preferred microphones (priority 3)
```

---

## Implementation Checklist

### Required Functions

- [ ] `listMicrophonesCommand()` - Enhanced version with helper text
- [ ] `showPreferencesCommand()` - Display current preferences
- [ ] `addPreferenceCommand(name string)` - Append to preferences
- [ ] `removePreferenceCommand(nameOrIndex string)` - Remove by name/index
- [ ] `clearPreferencesCommand()` - Clear all preferences

### Required Changes to Existing Code

- [ ] `internal/cli/config.go`: Add new subcommands to Cobra CLI
- [ ] `internal/config/config.go`: Add `AddPreference()`, `RemovePreference()`, `ClearPreferences()` methods
- [ ] Update `config` command output to display `preferred_microphones` field

---

## Testing Requirements

### Unit Tests

1. Parse arguments correctly (name vs index)
2. Validate empty names, duplicates
3. Handle out-of-range indices
4. Config save/load with preferences

### Integration Tests

1. Add preference → verify config file updated
2. Remove preference → verify config file updated
3. Clear preferences → verify empty array in config
4. List microphones with real devices

### Manual Testing

1. Connect/disconnect devices and verify `list-microphones` output
2. Add preferences for non-connected devices (should warn but succeed)
3. Remove preferences by name and by index
4. Clear preferences and verify fallback to default behavior

---

## Future Enhancements

1. **Interactive Mode**: `openscribe config --setup-preferences` with TUI menu
2. **Auto-Detection**: `openscribe config auto-preferences` to auto-populate based on currently connected devices
3. **Import/Export**: `openscribe config export-preferences` / `import-preferences <file>` for sharing configs
4. **Reorder Command**: `openscribe config reorder-preferences <from> <to>`
5. **Validation Command**: `openscribe config validate-preferences` to check all preferences against connected devices

---

## Summary

This CLI interface provides intuitive commands for managing microphone preferences while maintaining consistency with OpenScribe's existing Cobra-based CLI structure. All commands provide clear feedback, helpful error messages, and validation to guide users toward correct usage.

**Key Design Principles**:
- **Discoverability**: `list-microphones` helps users find exact device names
- **Simplicity**: Single-purpose commands (`add`, `remove`, `clear`)
- **Flexibility**: Support both name and index for removal
- **Safety**: Validate inputs, warn about non-connected devices
- **Consistency**: Follow existing OpenScribe CLI patterns

**Next Steps**: Generate quickstart.md guide for users.
