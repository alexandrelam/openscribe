# Research: Microphone Hotplug Detection and Automatic Switching

**Project**: OpenScribe
**Feature**: Preferred Microphones Fallback (001-preferred-microphones)
**Date**: 2025-10-19
**Scope**: macOS-only CLI application using malgo (wraps macOS Core Audio)

---

## Executive Summary

This research investigates best practices for implementing automatic microphone hotplug detection and switching in a Go CLI application using the malgo audio library on macOS. Key findings:

1. **Hotplug Detection**: Malgo/miniaudio supports device notifications through callbacks, but the Go wrapper (malgo v0.11.24) may have limited exposure. Polling is the most reliable fallback.
2. **Automatic Switching**: Switch only at application startup to avoid mid-recording disruptions; use a daemon pattern for hotplug responsiveness.
3. **Device Matching**: Case-insensitive exact matching is sufficient; avoid fuzzy matching complexity.
4. **Error Handling**: Graceful degradation with informative messages; preserve recorded audio on device loss.
5. **Configuration Migration**: Add new `PreferredMicrophones` field alongside existing `Microphone` field for backward compatibility.

---

## 1. Device Hotplug Detection

### Decision: **Hybrid Approach - Callbacks (if available) + Polling Fallback**

**Implementation Strategy**:
- **Primary**: Use malgo/miniaudio device notification callbacks if exposed in Go bindings
- **Fallback**: Implement periodic polling (every 2-5 seconds) to detect device changes
- **Trigger**: Only switch devices when recording is NOT active

### Rationale

1. **Malgo Library Capabilities**:
   - Malgo wraps miniaudio v0.11.x, which added a device notification system in 2021-2022
   - Miniaudio supports `ma_device_config.notificationCallback` for device state changes
   - Notification types include: `started`, `stopped`, `rerouted`, and `unlocked` (web only)
   - Core Audio (macOS) backend has **automatic stream routing** when default device changes
   - However, malgo v0.11.24 Go bindings primarily expose `DeviceCallbacks.Data` (audio streaming) and `DeviceCallbacks.Stop` (device stopped), not the newer notification system

2. **Polling is Reliable**:
   - Go's `time.Ticker` provides simple periodic device enumeration
   - `malgo.Context.Devices(malgo.Capture)` is lightweight (< 50ms on macOS)
   - Polling every 2-5 seconds is imperceptible to users and catches connect/disconnect events
   - No risk of missed events from callback timing issues

3. **macOS Core Audio Notifications** (Advanced Alternative):
   - Native Core Audio provides `AudioObjectAddPropertyListener` for hotplug events
   - Requires CGo integration with macOS frameworks (significant complexity)
   - Out of scope for initial implementation but viable for future enhancement

### Alternatives Considered

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Callbacks Only** | Real-time, efficient | May not be exposed in malgo Go bindings; requires custom C integration | ❌ Too complex for v1 |
| **Polling Only** | Simple, reliable, testable | 2-5 second delay in detection | ✅ **Recommended** |
| **Native Core Audio APIs** | Real-time OS notifications | Requires CGo, macOS-specific code, maintenance burden | ❌ Over-engineered |
| **No Detection** | Zero complexity | Poor UX - manual restart required | ❌ Defeats feature purpose |

### Implementation Guidance

**Recommended Polling Pattern**:

```go
// In audio package
type DeviceMonitor struct {
    ctx           *malgo.AllocatedContext
    ticker        *time.Ticker
    stopChan      chan struct{}
    lastDevices   []Device
    onChange      func([]Device) // Callback when devices change
}

func (m *DeviceMonitor) Start() {
    m.ticker = time.NewTicker(3 * time.Second)
    go func() {
        for {
            select {
            case <-m.ticker.C:
                current, _ := m.listDevices()
                if m.devicesChanged(current) {
                    m.lastDevices = current
                    m.onChange(current)
                }
            case <-m.stopChan:
                return
            }
        }
    }()
}

func (m *DeviceMonitor) devicesChanged(current []Device) bool {
    // Compare device IDs/names between current and lastDevices
    // Return true if list changed
}
```

**When to Check for Devices**:
- Application startup (initial selection)
- Every 3 seconds via background goroutine (when NOT recording)
- Pause monitoring during active recording (avoid mid-session switches)

**Performance Considerations**:
- Device enumeration takes ~10-50ms on macOS
- Run in background goroutine to avoid blocking main thread
- Stop monitoring during recording to reduce CPU usage

---

## 2. Automatic Device Switching

### Decision: **Switch at Application Startup Only (No Mid-Recording Switching)**

**Implementation Strategy**:
- When app starts: Select highest-priority available microphone from preferences
- During recording: Lock to current device; ignore hotplug events
- Between recordings: Re-evaluate device priorities on next recording start
- If device disconnects mid-recording: Stop recording, save audio, notify user

### Rationale

1. **Mid-Recording Switching is Risky**:
   - **Audio Pipeline Disruption**: Switching devices requires stopping/restarting the malgo device, which causes a recording gap (100-500ms)
   - **Sample Rate Mismatches**: Different microphones may have different native sample rates (44.1kHz vs 48kHz), requiring resampling mid-stream
   - **Buffer Discontinuities**: Accumulated audio data may have timestamp/buffer alignment issues when concatenating from different sources
   - **Whisper Incompatibility**: Whisper expects consistent audio properties (sample rate, format); mid-stream changes could corrupt transcription

2. **User Experience**:
   - Users don't expect devices to switch while actively recording
   - Notification of device change is better than silent switching
   - Users can manually restart recording if they want to use a newly connected device

3. **Simplicity**:
   - No need for complex stream merging or resampling logic
   - Clearer error handling and state management
   - Easier to test and reason about

4. **Alternative Daemon Pattern** (Recommended for OpenScribe):
   - OpenScribe runs as a background daemon waiting for hotkey press
   - Between recording sessions, device monitor can re-evaluate preferences
   - On hotkey press, select best available device at that moment
   - This gives "automatic switching" feel without mid-recording complexity

### Alternatives Considered

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Startup Only** | Simple, predictable, safe | Requires manual restart to use new device | ✅ **Recommended for v1** |
| **Between Sessions** | Automatic feel, no mid-recording risk | Requires daemon/background process | ✅ **Enhanced v1** (OpenScribe already has daemon) |
| **Mid-Recording** | Seamless UX | Complex, error-prone, audio glitches, buffer alignment issues | ❌ Not worth the risk |
| **User Confirmation** | Safe, explicit | Interrupts recording flow, annoying prompts | ❌ Poor UX |

### Implementation Guidance

**Device Selection Flow**:

```go
// At application startup or between recording sessions
func SelectMicrophone(cfg *config.Config) (*Device, error) {
    devices, err := ListMicrophones()
    if err != nil {
        return nil, err
    }

    // Try preferred microphones in order
    for _, prefName := range cfg.PreferredMicrophones {
        for _, dev := range devices {
            if strings.EqualFold(dev.Name, prefName) {
                log.Printf("Selected preferred microphone: %s", dev.Name)
                return &dev, nil
            }
        }
    }

    // Fallback to default
    return GetDefaultMicrophone()
}
```

**Mid-Recording Device Loss Handling**:

```go
// In recorder.go - Stop callback
func onDeviceStopped() {
    if recorder.isRecording {
        // Save whatever audio was captured
        audioData := recorder.GetCapturedAudio()
        SavePartialRecording(audioData)

        // Notify user
        log.Error("Microphone disconnected during recording")
        ShowNotification("Recording stopped: microphone disconnected")

        // Clean up state
        recorder.isRecording = false
    }
}
```

---

## 3. Device Name Matching

### Decision: **Case-Insensitive Exact Matching Only**

**Implementation Strategy**:
- Use `strings.EqualFold(configName, deviceName)` for comparison
- No substring matching, no fuzzy matching, no partial matches
- Require users to configure exact device names as reported by system

### Rationale

1. **Exact Matching Prevents Ambiguity**:
   - Multiple devices may share similar names: "USB Audio Device", "USB Audio Device 2"
   - Substring matching could inadvertently match wrong device
   - Exact matching ensures user intent is clear and predictable

2. **Case-Insensitive is Sufficient**:
   - macOS device names are generally case-consistent but users may type differently
   - `strings.EqualFold()` handles case variations without complexity
   - No need for Unicode normalization on macOS (English-centric device names)

3. **Device Names are Stable on macOS**:
   - Core Audio provides consistent device names for same hardware
   - Names typically include manufacturer and model: "Blue Yeti USB Microphone"
   - Users can discover exact names via CLI command: `openscribe config --list-microphones`

4. **Avoid Fuzzy Matching Complexity**:
   - Libraries like Levenshtein distance add dependencies
   - Fuzzy matching introduces unpredictability ("did it match correctly?")
   - Error-prone: typos could match wrong device silently
   - Performance overhead (O(n*m) for each comparison)

### Alternatives Considered

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Case-Insensitive Exact** | Clear, predictable, simple | Requires exact name entry | ✅ **Recommended** |
| **Case-Sensitive Exact** | Strictest matching | User-hostile, error-prone | ❌ Too strict |
| **Substring Matching** | More flexible for users | Ambiguous, could match wrong device | ❌ Risky |
| **Fuzzy Matching** (Levenshtein) | Typo-tolerant | Adds complexity, dependencies, unpredictable | ❌ Over-engineered |
| **Regex Matching** | Powerful for advanced users | Complex, error-prone, security risk | ❌ Unnecessary |

### Implementation Guidance

**Matching Function**:

```go
// In devices.go
func FindMicrophoneByPreferences(preferences []string) (*Device, error) {
    devices, err := ListMicrophones()
    if err != nil {
        return nil, err
    }

    // Try each preference in order
    for _, prefName := range preferences {
        for _, dev := range devices {
            // Case-insensitive exact match
            if strings.EqualFold(dev.Name, prefName) {
                return &dev, nil
            }
        }
    }

    return nil, fmt.Errorf("no preferred microphones available")
}
```

**Helper Command for Users**:

```bash
# CLI command to help users discover exact device names
$ openscribe config --list-microphones

Available Microphones:
  1. MacBook Pro Microphone (default)
  2. Blue Yeti USB Microphone
  3. AirPods Pro

To set preferences:
  $ openscribe config --add-preference "Blue Yeti USB Microphone"
  $ openscribe config --add-preference "AirPods Pro"
```

**Validation and User Feedback**:

```go
// When saving config, warn about non-matching preferences
func (c *Config) Validate() error {
    devices, _ := audio.ListMicrophones()
    deviceNames := make(map[string]bool)
    for _, d := range devices {
        deviceNames[strings.ToLower(d.Name)] = true
    }

    for _, pref := range c.PreferredMicrophones {
        if !deviceNames[strings.ToLower(pref)] {
            log.Printf("Warning: Preferred microphone '%s' not currently connected", pref)
        }
    }
    return nil
}
```

---

## 4. Error Handling Patterns

### Decision: **Graceful Degradation with Informative Messages**

**Implementation Strategy**:
- Never crash the application due to audio device issues
- Preserve captured audio data whenever possible
- Provide actionable error messages to guide users
- Fall back to default device if preferences fail
- Log device state changes for debugging

### Rationale

1. **Device Removal is Common**:
   - Users unplug USB devices, disconnect Bluetooth, close laptop lids
   - Temporary disconnections (Bluetooth interference, USB hub issues)
   - OS power management may disable devices

2. **User Expectations**:
   - Expect app to handle device issues gracefully, not crash
   - Want to know *why* something failed and *how* to fix it
   - Prefer saving partial recordings over losing all data

3. **Audio Pipeline Fragility**:
   - `malgo.Device.Start()` fails immediately if device unavailable
   - Data callback stops firing if device disconnected mid-recording
   - Stop callback may not always fire reliably (backend-dependent)

4. **Debugging Requirements**:
   - Users may report "microphone not working" without details
   - Logs should capture device enumeration, selection, and failures
   - Include device names, IDs, and system info in error context

### Error Categories and Handling

| Error Scenario | Handling Strategy | User Message |
|----------------|-------------------|--------------|
| **No devices found** | Exit gracefully, guide to system settings | "No microphones found. Check System Preferences > Sound > Input and grant microphone permissions." |
| **Preferred device not available** | Fallback to next preference or default | "Preferred microphone 'Blue Yeti' not found. Using 'MacBook Pro Microphone' instead." |
| **All preferences unavailable** | Use system default, log warning | "None of your preferred microphones are available. Using system default microphone." |
| **Device disconnected mid-recording** | Stop recording, save audio, notify | "Recording stopped: microphone disconnected. Audio saved to [path]." |
| **Device initialization fails** | Try next preference or default | "Failed to initialize 'Blue Yeti'. Trying next preference..." |
| **Permissions denied** | Exit with actionable instructions | "Microphone access denied. Grant permissions in System Preferences > Security & Privacy > Microphone." |

### Implementation Guidance

**Defensive Device Selection**:

```go
func SelectBestMicrophone(cfg *config.Config) (*Device, error) {
    devices, err := ListMicrophones()
    if err != nil {
        return nil, fmt.Errorf("failed to enumerate devices: %w\n\nPlease check:\n  1. Microphone is connected\n  2. System Preferences > Sound > Input\n  3. Microphone permissions granted", err)
    }

    // Try preferences in order
    for i, prefName := range cfg.PreferredMicrophones {
        for _, dev := range devices {
            if strings.EqualFold(dev.Name, prefName) {
                log.Printf("✓ Selected preferred microphone #%d: %s", i+1, dev.Name)
                return &dev, nil
            }
        }
        log.Printf("⚠ Preferred microphone #%d not available: %s", i+1, prefName)
    }

    // Fallback to default
    defaultDev, err := GetDefaultMicrophone()
    if err != nil {
        return nil, fmt.Errorf("no microphones available: %w", err)
    }

    if len(cfg.PreferredMicrophones) > 0 {
        log.Printf("⚠ Using fallback (default microphone): %s", defaultDev.Name)
    } else {
        log.Printf("✓ Using default microphone: %s", defaultDev.Name)
    }

    return defaultDev, nil
}
```

**Mid-Recording Error Handling**:

```go
// In recorder.go
func (r *Recorder) handleDeviceLoss() {
    // Save whatever audio was captured
    r.audioDataMutex.Lock()
    audioData := make([]byte, len(r.audioData))
    copy(audioData, r.audioData)
    r.audioDataMutex.Unlock()

    // Save to recovery file
    timestamp := time.Now().Format("20060102_150405")
    recoveryPath := filepath.Join(os.TempDir(), fmt.Sprintf("openscribe_recovery_%s.wav", timestamp))

    if err := SaveWAV(recoveryPath, audioData, r.sampleRate, r.channels); err != nil {
        log.Printf("ERROR: Failed to save recovery audio: %v", err)
    } else {
        log.Printf("✓ Partial recording saved to: %s", recoveryPath)
    }

    // Notify user (if UI available)
    // ShowNotification("Recording interrupted", "Microphone disconnected. Partial audio saved.")

    // Clean up
    r.isRecording = false
}
```

**Comprehensive Error Messages**:

```go
// Example from existing recorder.go (good pattern to continue)
if err := device.Start(); err != nil {
    device.Uninit()
    _ = ctx.Uninit()
    ctx.Free()
    return fmt.Errorf("failed to start audio recording: %w\n\nPossible causes:\n  1. Microphone disconnected or disabled\n  2. Microphone permissions not granted\n  3. Another app has exclusive access\n\nPlease check System Preferences > Security & Privacy > Privacy > Microphone", err)
}
```

**Logging Best Practices**:

```go
// Log device state changes for debugging
log.Printf("[AUDIO] Devices enumerated: %d found", len(devices))
for i, dev := range devices {
    log.Printf("[AUDIO]   %d. %s (ID: %s, Default: %v)", i+1, dev.Name, dev.ID, dev.IsDefault)
}

log.Printf("[AUDIO] Selected device: %s (from preferences: %v)", selectedDevice.Name, fromPreferences)

log.Printf("[AUDIO] Device monitor: detected change (added: %d, removed: %d)", numAdded, numRemoved)
```

---

## 5. Configuration Migration

### Decision: **Additive Migration - Keep Both Fields with Backward Compatibility**

**Implementation Strategy**:
- Add new `PreferredMicrophones []string` field to config
- Keep existing `Microphone string` field for backward compatibility
- Migration logic: If `PreferredMicrophones` is empty and `Microphone` is set, auto-populate `PreferredMicrophones` with single entry
- Selection logic: Use `PreferredMicrophones` if non-empty, else fall back to legacy `Microphone` field
- No breaking changes; older configs continue working

### Rationale

1. **Backward Compatibility is Critical**:
   - Existing users have `Microphone: "device name"` in configs
   - Removing or renaming breaks their configurations
   - Auto-migration provides seamless upgrade experience

2. **Additive Changes are Safe**:
   - YAML unmarshaling ignores unknown fields in old configs
   - New field has zero value (`[]string{}`) if not present
   - No data loss, no user intervention required

3. **Gradual Transition Path**:
   - Old configs work immediately after upgrade
   - Users can migrate to new field at their own pace
   - Future versions can deprecate old field with warning

4. **Clear Semantics**:
   - `PreferredMicrophones` array clearly indicates priority order
   - Empty array means "use default" (consistent with spec)
   - Legacy `Microphone` field is intuitive for single-device users

### Alternatives Considered

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Add field, keep both** | No breaking changes, seamless migration | Minor config complexity (two fields) | ✅ **Recommended** |
| **Replace field entirely** | Cleaner config schema | Breaks existing configs, requires manual migration | ❌ Too disruptive |
| **Separate config version** | Explicit schema evolution | Complex migration code, version tracking | ❌ Over-engineered |
| **Auto-migrate on first load** | Transparent to users | Hidden behavior, harder to debug | ⚠️ Combine with recommended |

### Implementation Guidance

**Config Schema**:

```go
// In config/config.go
type Config struct {
    // LEGACY: Single microphone name (backward compatibility)
    // Deprecated: Use PreferredMicrophones instead
    Microphone string `yaml:"microphone"`

    // NEW: Ordered list of preferred microphone names
    // If empty, falls back to Microphone field or system default
    PreferredMicrophones []string `yaml:"preferred_microphones,omitempty"`

    Model         string `yaml:"model"`
    Language      string `yaml:"language"`
    Hotkey        string `yaml:"hotkey"`
    AutoPaste     bool   `yaml:"auto_paste"`
    AudioFeedback bool   `yaml:"audio_feedback"`
    Verbose       bool   `yaml:"verbose"`
}
```

**Migration Logic**:

```go
// In config/config.go - called after loading config
func (c *Config) migrate() {
    // Auto-migrate: If new field is empty but old field is set, populate new field
    if len(c.PreferredMicrophones) == 0 && c.Microphone != "" {
        c.PreferredMicrophones = []string{c.Microphone}
        log.Printf("[CONFIG] Migrated legacy 'microphone' field to 'preferred_microphones'")

        // Optionally save migrated config immediately
        if err := c.Save(); err != nil {
            log.Printf("[CONFIG] Warning: Failed to save migrated config: %v", err)
        }
    }
}

// Call migration after loading
func Load() (*Config, error) {
    // ... existing load logic ...

    cfg.migrate()
    return cfg, nil
}
```

**Selection Logic**:

```go
// In audio/devices.go - unified device selection
func SelectMicrophone(cfg *config.Config) (*Device, error) {
    devices, err := ListMicrophones()
    if err != nil {
        return nil, err
    }

    // NEW: Try preferred microphones list first
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

    // LEGACY: Fall back to single microphone field
    if cfg.Microphone != "" {
        log.Printf("Using legacy 'microphone' config field: %s", cfg.Microphone)
        return FindMicrophoneByName(cfg.Microphone)
    }

    // DEFAULT: No preferences configured
    return GetDefaultMicrophone()
}
```

**Example YAML Configs**:

```yaml
# OLD CONFIG (v1.0) - still works
microphone: "Blue Yeti USB Microphone"
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

```yaml
# NEW CONFIG (v1.1+) - after migration
microphone: "Blue Yeti USB Microphone"  # Preserved for backward compat
preferred_microphones:
  - "Blue Yeti USB Microphone"
  - "AirPods Pro"
  - "MacBook Pro Microphone"
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

```yaml
# NEW CONFIG (v1.1+) - fresh install
preferred_microphones: []  # Empty = use default
model: "small"
language: ""
hotkey: "Right Option"
auto_paste: true
audio_feedback: true
verbose: false
```

**Validation Updates**:

```go
func (c *Config) Validate() error {
    // ... existing validation ...

    // NEW: Validate preferred microphones list
    for i, mic := range c.PreferredMicrophones {
        if strings.TrimSpace(mic) == "" {
            return fmt.Errorf("preferred_microphones[%d] cannot be empty", i)
        }
    }

    // Warn if both fields are set and differ
    if c.Microphone != "" && len(c.PreferredMicrophones) > 0 {
        if !contains(c.PreferredMicrophones, c.Microphone) {
            log.Printf("[CONFIG] Warning: Legacy 'microphone' field (%s) not in 'preferred_microphones' list", c.Microphone)
        }
    }

    return nil
}
```

**Future Deprecation Path** (v2.0+):

```go
// In Load() function - add deprecation warning
if cfg.Microphone != "" && len(cfg.PreferredMicrophones) == 0 {
    log.Printf("[CONFIG] DEPRECATED: The 'microphone' field is deprecated. Use 'preferred_microphones' instead.")
    log.Printf("[CONFIG] Run 'openscribe config --migrate' to update your configuration.")
}
```

---

## Summary of Recommendations

| Topic | Recommended Approach | Implementation Complexity |
|-------|---------------------|---------------------------|
| **Hotplug Detection** | Polling every 3 seconds (no callbacks) | Low |
| **Automatic Switching** | Between recording sessions only (daemon pattern) | Low |
| **Device Matching** | Case-insensitive exact matching | Low |
| **Error Handling** | Graceful degradation + informative messages | Medium |
| **Config Migration** | Additive field + auto-migration | Low |

**Total Implementation Effort**: ~2-3 days for experienced Go developer

**Testing Requirements**:
- Unit tests: Config parsing, device matching logic, migration
- Integration tests: Real device enumeration, selection with preferences
- Manual tests: Physical device connect/disconnect scenarios

**Performance Impact**: Negligible (<1ms for selection logic, ~30ms for polling)

**User Experience**: Seamless - existing configs work, new feature opt-in, clear error messages

---

## References

1. **Malgo Library**: https://github.com/gen2brain/malgo
2. **Miniaudio Documentation**: https://miniaud.io/docs/manual/index.html
3. **Miniaudio CHANGES.md**: Device notification system added ~v0.11
4. **Stack Overflow - macOS Audio Notifications**: https://stackoverflow.com/questions/9674666/
5. **Config Migration Best Practices**: Database migration patterns (expand-migrate-contract)
6. **String Matching**: Case-insensitive matching via `strings.EqualFold()` (Go stdlib)

---

**Next Steps**: Use this research to inform Phase 1 design artifacts (data model, contracts, quickstart guide).
