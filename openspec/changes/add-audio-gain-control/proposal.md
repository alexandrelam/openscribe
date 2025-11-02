# Proposal: Add Audio Gain Control

**Change ID:** `add-audio-gain-control`
**Status:** Draft
**Created:** 2025-11-02

## Why

Users report that microphone audio is sometimes too quiet, causing Whisper transcription to fail or produce poor results. Currently, OpenScribe has no visibility into audio levels and cannot automatically adjust them. Users must manually fiddle with system audio settings, which is frustrating and interrupts the workflow. Audio that's too quiet for transcription should be automatically boosted to an optimal level.

## What Changes

- Add `internal/audio/levels.go` module for RMS and dBFS calculation
- Add `internal/audio/gain.go` module for automatic gain normalization
- Modify `internal/config/config.go` to add gain control configuration fields
- Modify `internal/cli/start.go` to analyze levels and apply gain after recording
- Add audio level logging in verbose mode
- New config fields: `auto_gain`, `target_level_db`, `min_threshold_db`, `max_gain_db`, `show_audio_levels`

## Impact

- **Affected specs:** New capabilities `audio-level-monitoring` and `audio-gain-control`
- **Affected code:**
  - `internal/audio/levels.go` (new)
  - `internal/audio/gain.go` (new)
  - `internal/config/config.go` (modified)
  - `internal/cli/start.go` (modified)

## User Impact

**Positive:**
- Improved transcription accuracy for quiet microphones
- No manual system settings adjustment needed
- Transparent audio level information in verbose mode
- Configurable behavior for power users

**Risks:**
- Over-amplification could introduce noise if audio is extremely quiet
- Gain processing adds minor overhead (~1-5ms per recording)
- Users may need to understand dBFS concept for advanced configuration

## Alternatives Considered

1. **Manual warnings only (no auto-gain)** - Rejected: Still requires manual adjustment
2. **Pre-recording audio test** - Rejected: Adds friction, doesn't handle varying levels
3. **Real-time gain during recording** - Rejected: More complex, potential for artifacts
4. **Cloud-based enhancement** - Rejected: Violates privacy guarantee

Selected: **Post-recording normalization** - Simple, effective, no real-time complexity

## Success Criteria

- [ ] Audio levels measured accurately with RMS and dBFS
- [ ] Low audio automatically boosted to target level (-20 dBFS default)
- [ ] Transcription quality improves for quiet recordings
- [ ] No quality degradation for normal-level audio
- [ ] Configuration options work as documented
- [ ] Processing overhead < 10ms per recording
- [ ] All tests pass (unit + integration)

## Related Changes

None.

## References

- Audio engineering best practices: -20 dBFS target for speech
- RMS calculation: Root Mean Square for signal strength
- dBFS standard: Decibels Full Scale (0 = max, -∞ = silence)
- Whisper optimal input: Testing needed for model-specific targets
