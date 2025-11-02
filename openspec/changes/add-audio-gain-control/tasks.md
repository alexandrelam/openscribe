# Implementation Tasks

## Phase 1: Audio Level Monitoring (Foundation)

### Task 1.1: Create audio levels module
- Create `internal/audio/levels.go` with package structure
- Define `AudioLevelMetrics` struct with RMS, dBFS, peak, sample rate, duration fields
- Add package documentation and comments
- **Verifiable:** File exists with proper structure

### Task 1.2: Implement RMS calculation
- Implement `AnalyzeLevel(audioData []byte, sampleRate uint32) (AudioLevelMetrics, error)`
- Convert byte array to int16 samples
- Calculate sum of squares and RMS value
- Handle edge cases: empty data, silent audio, odd-length arrays
- **Verifiable:** Function compiles and handles basic inputs

### Task 1.3: Implement dBFS conversion
- Add RMS to dBFS conversion in `AnalyzeLevel()`
- Use formula: `20 * log10(rms / 32768)`
- Handle silence case (return -120.0 instead of -∞)
- Calculate audio duration from sample count and rate
- **Verifiable:** Returns valid dBFS values in range [-120, 0]

### Task 1.4: Implement peak detection
- Add peak amplitude tracking in `AnalyzeLevel()`
- Iterate samples and track maximum absolute value
- Return peak as int16 value
- **Verifiable:** Correctly identifies peak values in test data

### Task 1.5: Write unit tests for audio levels
- Create `internal/audio/levels_test.go`
- Test RMS calculation with known sine wave (predictable RMS)
- Test dBFS conversion accuracy
- Test peak detection with various waveforms
- Test edge cases: silence, maximum amplitude, empty data
- **Verifiable:** All tests pass with `go test ./internal/audio/...`

## Phase 2: Configuration (User Control)

### Task 2.1: Add configuration fields
- Edit `internal/config/config.go`
- Add `AutoGain bool` field (default: true)
- Add `TargetLevelDB float64` field (default: -20.0)
- Add `MinThresholdDB float64` field (default: -40.0)
- Add `MaxGainDB float64` field (default: 20.0)
- Add `ShowAudioLevels bool` field (default: false)
- Add YAML struct tags for all new fields
- **Verifiable:** Config struct compiles with new fields

### Task 2.2: Set default configuration values
- Update `DefaultConfig()` function to set sensible defaults
- Add validation for config values (e.g., target < 0, max gain > 0)
- Add comments explaining each field's purpose
- **Verifiable:** Default config includes gain control settings

### Task 2.3: Update configuration tests
- Edit `internal/config/config_test.go`
- Add tests for default gain control values
- Test config loading from YAML with gain settings
- Test validation of invalid values (if implemented)
- **Verifiable:** Config tests pass with new fields

## Phase 3: Gain Control (Core Feature)

### Task 3.1: Create gain control module
- Create `internal/audio/gain.go` with package structure
- Define `GainControlConfig` struct with enabled, target, threshold, max gain, prevent clipping fields
- Add package documentation
- **Verifiable:** File exists with proper structure

### Task 3.2: Implement gain calculation
- Implement `CalculateGain(currentDB, targetDB, maxGainDB float64) float64`
- Calculate required gain: `target - current`
- Apply maximum gain limit
- Return actual gain to apply in dB
- **Verifiable:** Function returns correct gain values

### Task 3.3: Implement dB to linear conversion
- Implement `DBToLinear(gainDB float64) float64`
- Use formula: `10^(gainDB / 20)`
- Test with known values (20dB = 10x, 6dB = 2x, etc.)
- **Verifiable:** Conversion matches expected values

### Task 3.4: Implement gain application
- Implement `ApplyGain(audioData []byte, gain float64) ([]byte, error)`
- Convert byte array to int16 samples
- Multiply each sample by linear gain
- Round to nearest integer
- Clamp to valid range [-32768, 32767]
- Convert back to byte array
- **Verifiable:** Function processes audio without errors

### Task 3.5: Implement clipping prevention
- Add peak detection to `ApplyGain()`
- Calculate safe gain: `32767 / currentPeak`
- Use minimum of calculated gain and safe gain
- Track if gain was reduced to prevent clipping
- Return adjusted audio and actual gain applied
- **Verifiable:** Peak never exceeds 32767 after gain

### Task 3.6: Create main gain control function
- Implement `ProcessAudioGain(audioData []byte, metrics AudioLevelMetrics, config GainControlConfig) ([]byte, float64, error)`
- Check if gain is needed (below threshold)
- Calculate required gain
- Convert to linear and apply
- Return processed audio and gain applied (in dB)
- **Verifiable:** Function orchestrates gain control correctly

### Task 3.7: Write unit tests for gain control
- Create `internal/audio/gain_test.go`
- Test gain calculation accuracy
- Test linear conversion
- Test gain application with known inputs
- Test clipping prevention (peak = 32767 after gain)
- Test gain limiting (max gain cap)
- Test bypass when level adequate
- **Verifiable:** All tests pass with `go test ./internal/audio/...`

## Phase 4: Integration (Connect the Pieces)

### Task 4.1: Integrate into start command
- Edit `internal/cli/start.go`
- After `recorder.Stop()`, call `audio.AnalyzeLevel()`
- Store `levelMetrics` result
- **Verifiable:** Code compiles and level analysis runs

### Task 4.2: Add verbose audio level display
- In start.go, check `cfg.Verbose || cfg.ShowAudioLevels`
- If true, display: `"Audio level: %.1f dBFS (peak: %d)"`
- Format with 1 decimal place precision
- **Verifiable:** Level displayed when verbose flag used

### Task 4.3: Add low level warning
- After level analysis, check if `levelMetrics.DecibelsFS < cfg.MinThresholdDB`
- If auto-gain disabled, display warning: `"⚠️  Low audio level detected (%.1f dBFS)"`
- Suggest user increase microphone input if possible
- **Verifiable:** Warning appears for quiet audio when gain disabled

### Task 4.4: Apply gain control
- After level analysis, check `cfg.AutoGain && levelMetrics.DecibelsFS < cfg.MinThresholdDB`
- If true, create `GainControlConfig` from config values
- Call `audio.ProcessAudioGain()` with audio data and metrics
- Update `audioData` with result
- Display message: `"⚠️  Low audio level detected (%.1f dBFS), applying gain..."`
- If verbose, also display: `"Applied gain: +%.1f dB"`
- **Verifiable:** Gain applied to quiet recordings

### Task 4.5: Update logging
- Edit `internal/logging/logger.go` if needed
- Consider adding audio level metrics to transcription log
- Add fields: audio_level_db, gain_applied_db (optional)
- **Verifiable:** Logs include audio metrics (if implemented)

## Phase 5: Testing & Validation

### Task 5.1: Create integration test
- Create test recordings at various volumes (loud, normal, quiet, silent)
- Process each through gain control
- Verify levels normalized to target range
- **Verifiable:** Integration test passes, audio normalized correctly

### Task 5.2: Test with actual transcription
- Record quiet audio samples
- Compare transcription quality with/without gain control
- Document improvement in accuracy
- **Verifiable:** Quiet audio transcribes better with gain enabled

### Task 5.3: Test configuration options
- Test with `auto_gain: false` - verify gain skipped
- Test with custom `target_level_db` - verify custom target used
- Test with custom `min_threshold_db` - verify custom threshold used
- Test with custom `max_gain_db` - verify gain limited
- **Verifiable:** All config options work as documented

### Task 5.4: Test edge cases
- Test with silence (all zeros)
- Test with maximum amplitude audio
- Test with already-optimal audio
- Test with extremely quiet audio (below -70 dBFS)
- **Verifiable:** No crashes, graceful handling of edge cases

### Task 5.5: Performance testing
- Measure processing time for typical 30-second recording
- Verify gain control adds < 10ms overhead
- Test with various recording lengths
- **Verifiable:** Performance impact is negligible

## Phase 6: Documentation & Polish

### Task 6.1: Update README
- Add section about automatic gain control feature
- Explain what it does and why it helps
- Document configuration options
- **Verifiable:** README mentions gain control

### Task 6.2: Update configuration documentation
- Document all new config fields in README or docs
- Provide examples of customization
- Explain when to adjust settings
- **Verifiable:** Config options documented

### Task 6.3: Add troubleshooting guidance
- Update TROUBLESHOOTING.md if exists
- Add section on audio level issues
- Explain how gain control helps
- **Verifiable:** Troubleshooting docs updated

### Task 6.4: Update CHANGELOG
- Add entry for new audio gain control feature
- List configuration options added
- **Verifiable:** CHANGELOG has entry

## Dependencies & Parallelization

**Sequential Dependencies:**
- Phase 1 must complete before Phase 3 (gain control needs level metrics)
- Phase 2 can be done in parallel with Phase 1
- Phase 3 depends on Phase 1 being complete
- Phase 4 depends on Phases 1, 2, and 3
- Phases 5 and 6 depend on Phase 4

**Can be parallelized:**
- Tasks 1.* and 2.* (levels and config are independent)
- Tasks 6.* (documentation tasks can be done in any order)

## Success Criteria

- ✅ All unit tests pass
- ✅ Integration tests demonstrate normalized audio levels
- ✅ Transcription quality improves for quiet recordings
- ✅ No degradation for normal-level recordings
- ✅ Configuration options work as expected
- ✅ Verbose mode shows audio level information
- ✅ Warning messages appear for low audio when appropriate
- ✅ No performance regression (< 10ms added processing time)
- ✅ Code is well-documented with clear comments
- ✅ User-facing documentation is complete and accurate
