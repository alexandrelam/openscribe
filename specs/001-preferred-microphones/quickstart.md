# Quickstart Guide: Preferred Microphones

**Feature**: 001-preferred-microphones
**Audience**: OpenScribe users
**Time to Complete**: 2-5 minutes

---

## What is This Feature?

OpenScribe now supports **automatic microphone fallback**. You can configure an ordered list of preferred microphones, and OpenScribe will automatically select the first available device from your list when you start recording.

This is perfect for users who:
- Frequently switch between multiple microphones (e.g., USB headset, built-in mic, studio mic)
- Want automatic device selection without manual reconfiguration
- Use different microphones in different contexts (home office, travel, meetings)

---

## Quick Start

### Step 1: See What Microphones Are Available

```bash
openscribe config list-microphones
```

**Output**:
```
Available Microphones:
  1. MacBook Pro Microphone (default)
  2. Blue Yeti USB Microphone
  3. AirPods Pro
```

---

### Step 2: Add Your Preferred Microphones (In Priority Order)

```bash
# Priority 1: Your main microphone
openscribe config add-preference "Blue Yeti USB Microphone"

# Priority 2: Backup microphone
openscribe config add-preference "AirPods Pro"

# Priority 3: Fallback to built-in
openscribe config add-preference "MacBook Pro Microphone"
```

Each command confirms the addition:
```
✓ Added "Blue Yeti USB Microphone" to preferred microphones (priority 1)
```

---

### Step 3: Verify Your Configuration

```bash
openscribe config show-preferences
```

**Output**:
```
Preferred Microphones (in priority order):
  1. Blue Yeti USB Microphone
  2. AirPods Pro
  3. MacBook Pro Microphone

Fallback: System default microphone
```

---

### Step 4: Start Using OpenScribe

That's it! Now when you use OpenScribe:

1. Press your hotkey (default: Right Option)
2. OpenScribe automatically selects the first available microphone from your list
3. Start speaking

**Example scenarios**:
- If "Blue Yeti USB Microphone" is connected → uses it
- If "Blue Yeti" is unplugged but "AirPods Pro" is connected → uses AirPods
- If neither is available → uses "MacBook Pro Microphone"
- If none of your preferences are available → uses system default

---

## Common Use Cases

### Use Case 1: Home Office + Travel Setup

```bash
# At home: Studio mic is first priority
openscribe config add-preference "Shure SM7B"

# When traveling: USB headset is second priority
openscribe config add-preference "Logitech USB Headset"

# Always available: Built-in mic as last resort
openscribe config add-preference "MacBook Pro Microphone"
```

**Result**: Automatically uses your studio mic at home, switches to USB headset when traveling, falls back to built-in mic if nothing else is available.

---

### Use Case 2: Single Preferred Microphone

```bash
# Just add one device - simpler than the old single-device config!
openscribe config add-preference "Blue Yeti USB Microphone"
```

**Result**: Always tries to use Blue Yeti; falls back to system default if unplugged.

---

### Use Case 3: Reset to Default Behavior

```bash
# Remove all preferences
openscribe config clear-preferences
```

**Result**: OpenScribe will always use the system default microphone (same as before this feature existed).

---

## Managing Your Preferences

### View Current Preferences

```bash
openscribe config show-preferences
```

---

### Add a Microphone

```bash
openscribe config add-preference "Microphone Name"
```

**Tip**: Use exact device names as shown by `list-microphones` command. Names are case-insensitive.

---

### Remove a Microphone

**Option 1: Remove by name**
```bash
openscribe config remove-preference "AirPods Pro"
```

**Option 2: Remove by priority number**
```bash
openscribe config remove-preference 2
```

---

### Remove All Preferences

```bash
openscribe config clear-preferences
```

---

## Tips and Best Practices

### Tip 1: Add Devices Even If Not Currently Connected

You can add a device to your preferences even if it's not currently plugged in:

```bash
$ openscribe config add-preference "Studio Mic Not Connected"
⚠ Warning: "Studio Mic Not Connected" is not currently connected
✓ Added "Studio Mic Not Connected" to preferred microphones (priority 2)
```

This is useful for setting up preferences for devices you use occasionally.

---

### Tip 2: Order Matters

Add your microphones in the order you prefer them:

```bash
# Best → Good → Okay
openscribe config add-preference "Best Mic"
openscribe config add-preference "Good Mic"
openscribe config add-preference "Okay Mic"
```

OpenScribe will always try "Best Mic" first, then "Good Mic", then "Okay Mic".

---

### Tip 3: Use Exact Names

Device names must match exactly (but case doesn't matter):

- ✅ `openscribe config add-preference "Blue Yeti USB Microphone"`
- ✅ `openscribe config add-preference "blue yeti usb microphone"` (case-insensitive)
- ❌ `openscribe config add-preference "Blue Yeti"` (partial name - won't match)

Use `list-microphones` to see exact names!

---

### Tip 4: Keep Your List Short

**Recommended**: 2-5 devices maximum

Why? A long list makes it harder to remember what you've configured. Keep it simple:
1. Your main microphone
2. Your backup microphone
3. Built-in microphone (optional fallback)

---

## Troubleshooting

### Problem: OpenScribe Not Using My Preferred Microphone

**Solution 1: Check if device is connected**
```bash
openscribe config list-microphones
```

Make sure your preferred device appears in the list.

---

**Solution 2: Check if name matches exactly**
```bash
# See your configured preferences
openscribe config show-preferences

# Compare with available devices
openscribe config list-microphones
```

If names don't match exactly, remove and re-add with correct name.

---

**Solution 3: Check device priority**

OpenScribe uses the first available device in your list. If a higher-priority device is connected, it will use that one instead.

---

### Problem: OpenScribe Says "No Microphones Found"

**Possible causes**:
1. No microphone is connected (plug in a device or ensure built-in mic is enabled)
2. Microphone permissions not granted to OpenScribe

**Fix permissions**:
1. Open **System Preferences > Security & Privacy > Privacy > Microphone**
2. Ensure OpenScribe is checked
3. Restart OpenScribe

---

### Problem: Want to Remove All Preferences and Start Over

```bash
openscribe config clear-preferences
```

This resets to default behavior (uses system default microphone).

---

## Migration from Old Configuration

If you previously used the single `microphone` field in your config, don't worry - it still works!

**Your old config** (v1.0):
```yaml
microphone: "Blue Yeti USB Microphone"
```

**Automatically migrated to** (v1.1):
```yaml
microphone: "Blue Yeti USB Microphone"  # Preserved
preferred_microphones:
  - "Blue Yeti USB Microphone"          # Auto-populated
```

No action required! Your config will be automatically upgraded when you first use OpenScribe v1.1+.

---

## Advanced: Editing Config File Directly

**Location**: `~/Library/Application Support/openscribe/config.yaml`

You can edit the `preferred_microphones` list directly in the YAML file:

```yaml
preferred_microphones:
  - "Blue Yeti USB Microphone"
  - "AirPods Pro"
  - "MacBook Pro Microphone"
```

**Rules**:
- Must be a YAML array (list)
- Each entry must be a non-empty string
- No duplicates (case-insensitive)
- Empty array `[]` means use default microphone

---

## How It Works (Technical Overview)

When you start recording:

1. **Device Enumeration**: OpenScribe lists all available audio input devices
2. **Priority Matching**: For each preferred microphone (in order):
   - Check if a device name matches (case-insensitive exact match)
   - If match found → use that device
3. **Fallback**: If no preferred devices available → use system default microphone
4. **Recording**: Start recording with selected device

**Selection happens**:
- At application startup
- Each time you press the hotkey to start recording (if using daemon mode)

**Selection does NOT happen**:
- During an active recording session (to avoid audio glitches)

---

## Feedback and Support

Found a bug or have a feature request?

1. Check logs: `~/Library/Logs/openscribe/transcriptions.log`
2. File an issue: [GitHub Issues](https://github.com/alexandrelam/openscribe-go/issues)

---

## What's Next?

Now that you've set up your preferred microphones, you can:

- Adjust your Whisper model: `openscribe config --set model medium`
- Change your hotkey: `openscribe config --set hotkey "Left Option"`
- Enable verbose logging to see device selection: `openscribe config --set verbose true`

Happy transcribing!
