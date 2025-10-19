# Data Model: Preferred Microphones Fallback

**Feature**: 001-preferred-microphones
**Date**: 2025-10-19
**Status**: Design Phase

---

## Overview

This document defines the data entities and their relationships for the Preferred Microphones Fallback feature. The feature extends the existing configuration system to support an ordered list of preferred microphone device names, enabling automatic fallback selection.

---

## Entities

### 1. Config (Modified)

**Location**: `internal/config/config.go`

**Purpose**: Represents the application configuration, including user preferences for microphone selection.

**Fields**:

| Field Name | Type | Required | Default | Description | Validation Rules |
|------------|------|----------|---------|-------------|------------------|
| `Microphone` | `string` | No | `""` | **(LEGACY)** Single microphone name for backward compatibility. Deprecated in favor of `PreferredMicrophones`. | If set, must be non-empty after trimming whitespace |
| `PreferredMicrophones` | `[]string` | No | `[]` | **(NEW)** Ordered list of preferred microphone device names. First available device in the list is selected. Empty list means use default microphone. | Each entry must be non-empty after trimming. No duplicate entries allowed. |
| `Model` | `string` | Yes | `"small"` | Whisper model size | Must be one of: {tiny, base, small, medium, large} |
| `Language` | `string` | No | `""` | Language code or empty for auto-detect | No specific validation (Whisper supports many languages) |
| `Hotkey` | `string` | Yes | `"Right Option"` | Keyboard shortcut for activation | Must be one of: {Left/Right Option, Shift, Command, Control} |
| `AutoPaste` | `bool` | Yes | `true` | Auto-paste transcribed text | N/A |
| `AudioFeedback` | `bool` | Yes | `true` | Play sounds on state changes | N/A |
| `Verbose` | `bool` | Yes | `false` | Debug output | N/A |

**YAML Representation**:

```yaml
# Example: User with multiple microphones configured
microphone: "Blue Yeti USB Microphone"  # Legacy field (backward compat)
preferred_microphones:
  - "Blue Yeti USB Microphone"          # Priority 1: Studio mic
  - "AirPods Pro"                       # Priority 2: Bluetooth headset
  - "MacBook Pro Microphone"            # Priority 3: Built-in fallback
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

```yaml
# Example: User with empty preferences (uses default)
preferred_microphones: []
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

**Relationships**:
- **Uses**: `Device` entities (from `internal/audio/devices.go`) for microphone selection

**State Transitions**:
- `Load()` → reads from `~/Library/Application Support/openscribe/config.yaml`
- `migrate()` → auto-migrates `Microphone` field to `PreferredMicrophones` if needed
- `Validate()` → checks all fields meet validation rules
- `Save()` → persists to YAML file

**Migration Logic**:
```go
// Automatically called after Load()
func (c *Config) migrate() {
    if len(c.PreferredMicrophones) == 0 && c.Microphone != "" {
        c.PreferredMicrophones = []string{c.Microphone}
        log.Printf("Migrated legacy 'microphone' field to 'preferred_microphones'")
        _ = c.Save()  // Optional: persist migration immediately
    }
}
```

---

### 2. Device (Existing - No Changes)

**Location**: `internal/audio/devices.go`

**Purpose**: Represents an audio input device (microphone) detected by the system.

**Fields**:

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `ID` | `string` | Yes | System-assigned device identifier (numeric string) |
| `Name` | `string` | Yes | Human-readable device name (e.g., "Blue Yeti USB Microphone") |
| `IsDefault` | `bool` | Yes | Whether this device is the system default input |
| `SampleRate` | `uint32` | Yes | Sample rate in Hz (16000 for Whisper compatibility) |
| `Channels` | `uint32` | Yes | Number of channels (1 for mono) |

**Example**:
```go
Device{
    ID:         "AppleHDAEngineInput:1B,0,1,0:1",
    Name:       "MacBook Pro Microphone",
    IsDefault:  true,
    SampleRate: 16000,
    Channels:   1,
}
```

**Lifecycle**:
- Created by `ListMicrophones()` via malgo device enumeration
- Matched against `Config.PreferredMicrophones` using `strings.EqualFold()`
- Selected device is passed to recorder for audio capture

---

### 3. DeviceMonitor (NEW - Optional for Future Enhancement)

**Location**: `internal/audio/monitor.go` (new file)

**Purpose**: Monitors audio device connections/disconnections via polling. **Note**: Not required for MVP but documented for future implementation.

**Fields**:

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `ctx` | `*malgo.AllocatedContext` | Yes | Malgo context for device enumeration |
| `ticker` | `*time.Ticker` | Yes | Ticker for periodic polling (every 3 seconds) |
| `stopChan` | `chan struct{}` | Yes | Channel to signal shutdown |
| `lastDevices` | `[]Device` | Yes | Cached list of devices from last check |
| `onChange` | `func([]Device)` | Yes | Callback invoked when device list changes |

**State Transitions**:
- `New()` → creates monitor with given context and callback
- `Start()` → begins polling goroutine
- `Stop()` → stops polling and cleans up resources

**Usage Pattern** (future implementation):
```go
monitor := audio.NewDeviceMonitor(ctx, func(devices []Device) {
    log.Printf("Devices changed: %d available", len(devices))
    // Trigger re-selection logic here
})
monitor.Start()
defer monitor.Stop()
```

---

## Validation Rules

### Config Validation

**Implemented in**: `internal/config/config.go:Validate()`

1. **PreferredMicrophones**:
   - Each entry must be non-empty after `strings.TrimSpace()`
   - No duplicate entries (case-insensitive comparison)
   - No maximum length, but UI should warn if > 10 entries (performance consideration)

2. **Microphone** (legacy field):
   - If set alongside `PreferredMicrophones`, log a warning if not present in the new list
   - If set and `PreferredMicrophones` is empty, auto-migration should populate `PreferredMicrophones`

3. **Cross-Field Validation**:
   - Warning (not error) if configured device names don't match any currently connected devices
   - Helpful message: "Warning: Preferred microphone 'XYZ' not currently connected"

**Example Validation Code**:
```go
func (c *Config) Validate() error {
    // Validate preferred microphones
    seen := make(map[string]bool)
    for i, mic := range c.PreferredMicrophones {
        trimmed := strings.TrimSpace(mic)
        if trimmed == "" {
            return fmt.Errorf("preferred_microphones[%d] cannot be empty", i)
        }
        lowerMic := strings.ToLower(trimmed)
        if seen[lowerMic] {
            return fmt.Errorf("duplicate preferred microphone: %s", trimmed)
        }
        seen[lowerMic] = true
    }

    // Warning for non-connected devices
    devices, _ := audio.ListMicrophones()
    deviceNames := make(map[string]bool)
    for _, d := range devices {
        deviceNames[strings.ToLower(d.Name)] = true
    }
    for _, pref := range c.PreferredMicrophones {
        if !deviceNames[strings.ToLower(pref)] {
            log.Printf("[CONFIG] Warning: Preferred microphone '%s' not currently connected", pref)
        }
    }

    // ... existing validation for Model, Hotkey, etc. ...

    return nil
}
```

---

## Selection Algorithm

**Implemented in**: `internal/audio/devices.go:SelectMicrophone()`

**Purpose**: Given a Config, select the best available microphone device.

**Algorithm**:
```
1. Enumerate all available devices via ListMicrophones()
2. IF Config.PreferredMicrophones is non-empty:
     a. FOR EACH preferred name IN order:
          i. FOR EACH available device:
               - IF strings.EqualFold(preferred name, device.Name):
                   → RETURN device (first match wins)
     b. IF no matches found:
          → Log warning: "No preferred microphones available"
          → GOTO step 3 (fallback)
3. ELSE IF Config.Microphone is set (legacy):
     a. Find device matching Config.Microphone (case-insensitive exact match)
     b. IF found: RETURN device
     c. ELSE: Log warning, GOTO step 4
4. FALLBACK: Return GetDefaultMicrophone()
     a. First device with IsDefault=true
     b. OR first device in enumeration list
     c. OR error if no devices available
```

**Pseudo-code**:
```go
func SelectMicrophone(cfg *config.Config) (*Device, error) {
    devices, err := ListMicrophones()
    if err != nil {
        return nil, fmt.Errorf("failed to enumerate devices: %w", err)
    }

    // Try preferred microphones
    if len(cfg.PreferredMicrophones) > 0 {
        for i, prefName := range cfg.PreferredMicrophones {
            for _, dev := range devices {
                if strings.EqualFold(dev.Name, prefName) {
                    log.Printf("Selected preferred microphone #%d: %s", i+1, dev.Name)
                    return &dev, nil
                }
            }
        }
        log.Printf("No preferred microphones available, falling back to default")
        return GetDefaultMicrophone()
    }

    // Legacy: single microphone field
    if cfg.Microphone != "" {
        log.Printf("Using legacy 'microphone' config field: %s", cfg.Microphone)
        return FindMicrophoneByName(cfg.Microphone)
    }

    // Default behavior
    return GetDefaultMicrophone()
}
```

**Complexity**: O(P * D) where P = number of preferences, D = number of devices (typically P ≤ 5, D ≤ 10, so < 50 comparisons)

---

## Persistence

### Storage Location
- **Path**: `~/Library/Application Support/openscribe/config.yaml`
- **Format**: YAML (UTF-8)
- **Permissions**: User read/write only (`0600`)

### YAML Schema

```yaml
# Required fields
model: string                          # Whisper model size
hotkey: string                         # Keyboard shortcut
auto_paste: boolean                    # Auto-paste toggle
audio_feedback: boolean                # Sound effects toggle
verbose: boolean                       # Debug logging toggle

# Optional fields
microphone: string                     # LEGACY: single device name
preferred_microphones: array[string]   # NEW: ordered preference list
language: string                       # Language code (empty = auto)
```

### Example Configurations

**1. Fresh Install (No Preferences)**:
```yaml
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

**2. User with Single Preferred Mic**:
```yaml
microphone: "Blue Yeti USB Microphone"  # Preserved after migration
preferred_microphones:
  - "Blue Yeti USB Microphone"
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

**3. User with Multiple Preferences**:
```yaml
preferred_microphones:
  - "Shure SM7B"                       # Priority 1: Studio mic
  - "Blue Yeti USB Microphone"         # Priority 2: Backup USB mic
  - "AirPods Pro"                      # Priority 3: Wireless
  - "MacBook Pro Microphone"           # Priority 4: Built-in
model: "medium"
language: "en"
hotkey: "Left Option"
auto_paste: true
audio_feedback: false
verbose: false
```

---

## Error States

### Device Selection Errors

| Error Condition | Handling | User-Visible Message |
|----------------|----------|---------------------|
| No devices found | Exit gracefully with error | "No microphones found. Check System Preferences > Sound > Input." |
| Permission denied | Exit gracefully with error | "Microphone access denied. Grant permissions in System Preferences > Security & Privacy > Microphone." |
| All preferences unavailable | Fallback to default | "None of your preferred microphones are available. Using default microphone: [name]." |
| Device disconnected mid-recording | Stop recording, save audio | "Recording stopped: microphone disconnected. Audio saved to [path]." |

### Config Validation Errors

| Error Condition | Handling | User-Visible Message |
|----------------|----------|---------------------|
| Empty string in preferences | Reject config, show error | "Error: preferred_microphones[2] cannot be empty." |
| Duplicate entries | Reject config, show error | "Error: Duplicate preferred microphone: 'Blue Yeti USB Microphone'." |
| Invalid YAML syntax | Reject config, show error | "Error: Invalid config file at line 5: [parse error]" |

---

## Testing Considerations

### Unit Test Scenarios

1. **Config Parsing**:
   - Parse valid YAML with `preferred_microphones` array
   - Parse legacy YAML with only `microphone` field
   - Reject empty strings in preferences
   - Reject duplicate preferences (case-insensitive)

2. **Migration**:
   - Auto-migrate single `Microphone` → `PreferredMicrophones` array
   - Don't overwrite existing `PreferredMicrophones` if already set
   - Handle empty/nil values gracefully

3. **Device Selection**:
   - Select first matching device from preferences
   - Fall back to second preference if first unavailable
   - Fall back to default if all preferences unavailable
   - Handle empty preferences list (use default)

### Integration Test Scenarios

1. **Real Device Enumeration**:
   - List available devices on test machine
   - Select device by exact name match (case-insensitive)
   - Handle disconnected devices gracefully

2. **End-to-End**:
   - Load config with preferences → Start recording → Verify correct device used
   - Simulate device disconnect mid-recording → Verify audio saved

---

## Migration Path

### From v1.0 (Single Microphone) → v1.1 (Preferred Microphones)

**Step 1: Auto-Migration on Load**

```go
// In config.Load()
func Load() (*Config, error) {
    cfg := &Config{}
    // ... load from YAML ...

    cfg.migrate()  // Auto-convert Microphone → PreferredMicrophones
    return cfg, nil
}
```

**Step 2: Preserve Legacy Field**

Keep `Microphone` field in struct but mark as deprecated in comments and documentation.

**Step 3: No User Action Required**

Existing configs work immediately without user intervention.

**Step 4: Future Deprecation (v2.0+)**

Add warning logs:
```go
if cfg.Microphone != "" && len(cfg.PreferredMicrophones) == 0 {
    log.Warn("DEPRECATED: 'microphone' field will be removed in v2.0. Use 'preferred_microphones' instead.")
}
```

---

## Summary

This data model extends the existing `Config` struct with a new `PreferredMicrophones` field, enabling automatic fallback selection while maintaining full backward compatibility. The design follows established patterns in the codebase (YAML config, case-insensitive matching, graceful error handling) and requires minimal changes to existing code.

**Key Design Decisions**:
- **Additive Change**: New field alongside legacy field (no breaking changes)
- **Auto-Migration**: Transparent upgrade experience
- **Simple Matching**: Case-insensitive exact match (no fuzzy logic)
- **Graceful Fallback**: Always select *some* device, never crash
- **Clear Validation**: Helpful error messages for misconfigurations

**Next Steps**: Generate API contracts (CLI commands) and quickstart guide.
