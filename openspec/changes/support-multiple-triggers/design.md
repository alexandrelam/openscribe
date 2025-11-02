# Design: Support Multiple Triggers

**Change ID:** support-multiple-triggers

## Overview

This document outlines the architectural design for supporting multiple trigger inputs (keyboard keys and mouse buttons) in OpenScribe.

---

## System Architecture

### Current Architecture

```
Config (single hotkey)
    ↓
CLI Start
    ↓
hotkey.NewListener(single key) → Listener
    ↓
Platform-specific event tap (keyboard only)
    ↓
Callback on double-press
```

### Proposed Architecture

```
Config (triggers array)
    ↓
CLI Start
    ↓
hotkey.NewMultiListener(triggers) → MultiListener
    ↓                                      ↓
    ├─ Listener (keyboard)              Listener (mouse)
    │      ↓                                  ↓
    │  Event tap (kCGEventFlagsChanged)  Event tap (kCGEventOtherMouseDown)
    │      ↓                                  ↓
    └────────→ Shared callback ←─────────────┘
               (double-press detection)
```

---

## Key Design Decisions

### 1. Config Schema: Array vs. Object

**Decision:** Use `triggers: []string` (simple string array)

**Alternatives Considered:**
- Object array with metadata: `triggers: [{type: "keyboard", name: "Right Option"}]`
- Separate fields: `keyboard_triggers: []` and `mouse_triggers: []`

**Rationale:**
- Simple string array is easier for users to configure
- Trigger names are self-descriptive ("Forward Button" vs "Right Option")
- Less YAML complexity
- Can add object format later if needed (backward compatible)

**Example:**
```yaml
# Simple (chosen)
triggers:
  - "Right Option"
  - "Forward Button"

# Complex (rejected for v1)
triggers:
  - type: keyboard
    key: "Right Option"
  - type: mouse
    button: 4
```

---

### 2. Multi-Listener vs. Single Listener

**Decision:** Create `MultiListener` wrapper that manages multiple `Listener` instances

**Alternatives Considered:**
- Single listener with multiple event taps
- Single event tap monitoring all input types

**Rationale:**
- **Separation of concerns:** Each listener handles one input type
- **Easier debugging:** Issues isolated to specific listeners
- **Minimal changes to existing code:** Existing `Listener` logic unchanged
- **Flexibility:** Easy to add/remove listeners dynamically

**Trade-offs:**
- Slightly more memory (multiple listeners)
- More goroutines (acceptable on modern systems)
- **Benefit:** Cleaner code, easier testing

---

### 3. Mouse Button Detection Method

**Decision:** Use `kCGEventOtherMouseDown` event type with button number field

**Alternatives Considered:**
- NSEvent monitoring
- IOKit HID device monitoring
- Third-party mouse driver integration

**Rationale:**
- `CGEvent` API is the standard macOS event system
- Already using CGEvent for keyboard monitoring
- Single permission model (Accessibility)
- Low-level access, works with all mice
- No external dependencies

**Mouse Button Mapping:**
- Button 0: Left click (ignored)
- Button 1: Right click (ignored)
- Button 2: Middle click (could add later)
- Button 3: Back button ← **SUPPORTED**
- Button 4: Forward button ← **SUPPORTED**

---

### 4. Backward Compatibility Strategy

**Decision:** Auto-migrate `hotkey` → `triggers` on first load, preserve legacy field

**Alternatives Considered:**
- Breaking change: Remove `hotkey` field entirely
- Manual migration tool: `openscribe config migrate`
- Dual support forever: Check both fields

**Rationale:**
- **Zero user action required:** Migration happens automatically
- **Safe:** Legacy field preserved (can rollback)
- **Clear deprecation path:** Log message informs users
- **One-time migration:** Future loads use `triggers` only

**Migration Logic:**
```go
func (c *Config) migrate() {
    if len(c.Triggers) == 0 && c.Hotkey != "" {
        c.Triggers = []string{c.Hotkey}
        log.Printf("Migrated 'hotkey' to 'triggers': %s", c.Hotkey)
        c.Save()
    }
}
```

---

### 5. Double-Press Detection: Per-Trigger vs. Global

**Decision:** Independent double-press detection per trigger

**Alternatives Considered:**
- Global double-press: Any two trigger presses (even different ones)
- Hybrid: Configurable per-trigger or global

**Rationale:**
- **Intuitive:** Users expect double-press of *same* trigger
- **Avoids accidental activation:** Different triggers don't combine
- **Simpler state management:** Each listener tracks its own state
- **Example:** User presses "Right Option" once, then "Forward Button" once → No activation

---

### 6. Error Handling: Partial Success vs. All-or-Nothing

**Decision:** All-or-nothing - if any trigger fails to register, exit

**Alternatives Considered:**
- Partial success: Register working triggers, log failures
- Retry logic: Attempt registration multiple times

**Rationale:**
- **Clear failure mode:** User knows something is wrong
- **Consistent behavior:** All configured triggers work or none
- **Avoid confusion:** Don't want users unsure which triggers are active

**Error Example:**
```
Error: Failed to register trigger 'Forward Button': mouse events require accessibility permissions
```

---

## Implementation Details

### MultiListener Struct

```go
type MultiListener struct {
    listeners []*Listener
    callback  func()
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

func NewMultiListener(triggerNames []string, callback func()) (*MultiListener, error) {
    // Create listener for each trigger
    // Share same callback
    // Return multi-listener wrapper
}

func (ml *MultiListener) Start() error {
    // Start all listeners
    // If any fail, stop all and return error
}

func (ml *MultiListener) Stop() {
    // Stop all listeners gracefully
}
```

---

### Mouse Event Detection (C Code)

```c
// In hotkey_darwin.go CGO section
static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventOtherMouseDown) {
        int64_t buttonNumber = CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);

        // Map button numbers to synthetic key codes
        uint16_t syntheticKeyCode = 0;
        if (buttonNumber == 3) syntheticKeyCode = 0x10002; // Back
        if (buttonNumber == 4) syntheticKeyCode = 0x10001; // Forward

        if (syntheticKeyCode == gTargetKeyCode) {
            goHotkeyCallback();
        }
    }
    // ... existing keyboard handling
}
```

---

### Config Validation Flow

```
Load config.yaml
    ↓
Parse YAML → Config struct
    ↓
Run migrate() → Convert Hotkey to Triggers if needed
    ↓
Run Validate()
    ├─ Check Triggers is not empty
    ├─ Check each trigger is non-empty string
    ├─ Check for duplicates (case-insensitive)
    ├─ Validate trigger names against known list
    └─ Return error if any validation fails
    ↓
Config ready to use
```

---

## Security & Permissions

### Accessibility Permissions

**Current:** Required for keyboard monitoring
**New:** Still required (same permissions model)

Mouse button detection uses same CGEventTap mechanism, so no additional permissions needed.

### Privacy Considerations

- Events only monitored when OpenScribe is running
- No event data is logged or transmitted
- Only trigger presses are detected (not full mouse/keyboard input)

---

## Performance Considerations

### Multiple Event Taps

Each listener creates an event tap. macOS limits event taps per process, but typical limit is 128+.

**Expected:** 2-5 triggers → 2-5 event taps (well within limits)

### Memory Overhead

Each `Listener` instance: ~1KB
With 5 triggers: ~5KB total (negligible)

### CPU Usage

Event processing is asynchronous and event-driven. No polling. CPU impact is minimal (<0.1% even with multiple triggers).

---

## Testing Strategy

### Unit Tests
- Config validation (duplicates, empty, invalid)
- Migration logic (hotkey → triggers)
- MultiListener start/stop

### Integration Tests
- Keyboard trigger detection
- Mouse button trigger detection
- Multi-trigger simultaneous activation

### Manual Testing
- Test with physical gaming mouse
- Test all supported keyboard modifiers
- Test mixed keyboard + mouse configuration

---

## Future Enhancements

### Potential Additions (out of scope for this change)

1. **Middle mouse button support**
   - Add `"Middle Button"` to supported triggers

2. **Custom double-press delay**
   - Config: `trigger_delay_ms: 500`

3. **Single-press triggers**
   - Config: `trigger_mode: "single" | "double"`

4. **Chord triggers**
   - Config: `triggers: ["Right Option + Forward Button"]`

5. **Per-trigger actions**
   - Different triggers for start vs. stop recording

---

## Risks & Mitigations

### Risk 1: Mouse button codes vary by vendor

**Mitigation:** Use standard macOS button numbers (3=Back, 4=Forward). Most mice follow this convention.

**Fallback:** Document how to check button numbers with system utilities.

---

### Risk 2: Accessibility permissions confusion

**Mitigation:**
- Clear error messages
- Documentation explains permissions apply to all triggers
- Permission prompt happens once for all input types

---

### Risk 3: Migration fails for complex legacy configs

**Mitigation:**
- Migration is simple: single string → array of one string
- If migration fails, use default config and log error
- User can manually fix config file

---

## Conclusion

The proposed architecture:
- ✅ Supports multiple keyboard and mouse triggers
- ✅ Maintains backward compatibility
- ✅ Uses platform-standard APIs (CGEvent)
- ✅ Minimal performance overhead
- ✅ Clear error handling
- ✅ Extensible for future trigger types

This design balances user flexibility, implementation simplicity, and maintainability.
