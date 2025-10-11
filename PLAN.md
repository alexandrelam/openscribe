# OpenScribe Development Plan

## Overview

This document outlines a phased approach to building OpenScribe using LLM-assisted development (Claude Code). The decomposition is optimized for iterative development where each phase delivers testable, functional components that build upon previous work.

## Development Philosophy for LLM-Assisted Coding

- **Larger, cohesive chunks**: Group related functionality together since LLMs can handle multiple files simultaneously
- **Clear interfaces**: Each phase should have well-defined inputs/outputs for easy testing
- **Incremental complexity**: Start with foundations, add complexity gradually
- **Feature completeness**: Each phase should deliver something testable/runnable
- **Minimal dependencies**: Reduce coupling between phases where possible

---

## Phase 1: Foundation & Project Skeleton

**Goal**: Establish the project structure, dependency management, and basic CLI framework

**Deliverables**:
- Go module initialization (`go.mod`, `go.mod`)
- Project directory structure (`cmd/`, `internal/`, `assets/`, etc.)
- Cobra CLI framework integrated with basic command structure
- Basic `main.go` entry point
- Placeholder commands: `setup`, `start`, `config`, `models`, `logs`, `version`
- Can run `openscribe version` and `openscribe --help`

**Why this grouping**: Project setup is boilerplate that benefits from being done all at once. The CLI skeleton provides a testable framework to build upon.

**Test**: Run `go build` successfully, execute `openscribe --help` and see command structure

---

## Phase 2: Configuration & Path Management

**Goal**: Build the configuration system that other components will depend on

**Deliverables**:
- Configuration file structure (`config.yaml` schema)
- Config read/write logic with defaults
- macOS standard paths implementation:
  - `~/Library/Application Support/openscribe/` (config, models)
  - `~/Library/Caches/openscribe/` (temp files)
  - `~/Library/Logs/openscribe/` (logs)
- Directory creation on first run
- `openscribe config --show` command working
- Config validation and default value handling

**Why this grouping**: Configuration is foundational infrastructure that multiple components need. Getting it right early prevents refactoring later.

**Test**: Run `openscribe config --show`, verify config file created at correct path with defaults

---

## Phase 3: Model Management & Whisper Setup

**Goal**: Handle whisper.cpp integration and model lifecycle

**Deliverables**:
- Model download logic (fetch Whisper models from official sources)
- Model storage in `~/Library/Application Support/openscribe/models/`
- Model validation (checksum, file integrity)
- whisper.cpp detection via Homebrew
- `openscribe models list` command
- `openscribe models download [model]` command
- `openscribe setup` command (checks for whisper-cpp + downloads default small model)
- Progress indicators for downloads

**Implementation Note**: Originally planned to download and compile whisper.cpp from source, but switched to using Homebrew installation (`brew install whisper-cpp`). This is much simpler, faster, and more maintainable. The setup command now checks if `whisper-cli` is available and guides users to install via Homebrew if needed.

**Why this grouping**: Model management is a complete subsystem with no runtime dependencies on audio/recording. Can be tested independently.

**Test**: Run `openscribe setup`, verify whisper-cpp is installed via Homebrew and model downloaded. Run `openscribe models list` to see available models.

---

## Phase 4: Audio Recording Infrastructure

**Goal**: Implement microphone detection and audio recording to file

**Deliverables**:
- Microphone enumeration (list all input devices)
- Microphone selection logic
- Audio recording to WAV file (16kHz, mono, format suitable for Whisper)
- Save recording to temp cache directory
- `openscribe config --list-microphones` command
- `openscribe config --set-microphone` command
- Basic recording test (record 5 seconds, save to file)

**Why this grouping**: Audio recording is a discrete subsystem. We can test it independently before integrating transcription.

**Test**: Run microphone listing command, select a microphone, create a simple test command that records 5 seconds of audio to verify recording works.

---

## Phase 5: Transcription Engine

**Goal**: Integrate whisper.cpp for speech-to-text transcription

**Deliverables**:
- Go bindings to whisper.cpp (via CGo or CLI invocation)
- Model loading and caching
- Transcription function (audio file → text)
- Language detection and language parameter support
- Error handling for transcription failures
- Simple test: transcribe a pre-recorded WAV file

**Why this grouping**: Transcription is the core feature. Once we have recording (Phase 4) and transcription (Phase 5), we have the complete audio pipeline.

**Test**: Feed a known audio file through transcription, verify correct text output. Test with different models and languages.

---

## Phase 6: Logging & History

**Goal**: Implement transcription logging and history management

**Deliverables**:
- Log file structure and format (timestamp, duration, model, language, text)
- Write transcriptions to `~/Library/Logs/openscribe/transcriptions.log`
- `openscribe logs show` command
- `openscribe logs --tail` command
- `openscribe logs clear` command
- Log rotation or size management (optional)

**Why this grouping**: Logging is independent and straightforward. Better to implement it now so Phase 7-8 can use it.

**Test**: Manually trigger a transcription, verify log file created and contains correct data. Test log display commands.

---

## Phase 7: Hotkey Detection & Event Loop ✅

**Goal**: Implement global hotkey listener for macOS

**Deliverables**:
- macOS hotkey detection using CGo and Carbon/Cocoa APIs
- Double-press detection logic (time threshold between presses)
- Configurable hotkey (default: Right Option)
- Event loop that listens for hotkey
- `openscribe config --set-hotkey` command
- Basic test mode that prints to terminal when hotkey detected
- Graceful Ctrl+C handling to exit

**Why this grouping**: Hotkey detection is a complex system integration piece. It's isolated from audio/transcription but requires macOS-specific code.

**Test**: Run a test mode that prints "Hotkey detected!" when double-press occurs. Verify Ctrl+C exits cleanly.

**Implementation Notes**:
- Created `internal/hotkey` package with platform-specific macOS implementation
- Uses Carbon Event Manager API for global hotkey registration
- Double-press detection with configurable 500ms window
- Supports 8 common modifier keys (Option, Shift, Cmd, Ctrl - left and right)
- Added `--list-hotkeys` flag to config command
- Integrated hotkey listener into `openscribe start` command with proper signal handling
- All C functions marked as `static` to avoid duplicate symbol linker errors

**Testing**:
- Run `openscribe config --list-hotkeys` to see available hotkeys
- Run `openscribe config --set-hotkey "Right Option"` to configure hotkey
- Run `openscribe start` to test hotkey detection (requires Accessibility permissions)
- Note: Actual hotkey detection requires macOS Accessibility permissions to be granted

---

## Phase 8: Audio Feedback System ✅

**Goal**: Add sound effects for recording state changes

**Deliverables**:
- Three distinct sound files (start, stop, complete) or system sound selection
- Sound playback using AVFoundation/NSSound via CGo
- Bundle sound files in `assets/` directory or use system sounds
- Play sounds at appropriate times:
  - Start recording → start sound
  - Stop recording → stop sound
  - Transcription complete → complete sound
- Configuration option to disable audio feedback

**Why this grouping**: Audio feedback is a nice-to-have that's independent of core functionality. Adding it late means it won't block critical features.

**Test**: Trigger each sound manually, verify they're distinct and brief.

**Implementation Notes**:
- Created `internal/audio/feedback.go` with platform-agnostic interface
- macOS implementation in `feedback_darwin.go` uses NSSound via CGo
- Uses built-in macOS system sounds (no bundled files needed):
  - Start: "Tink" (short ascending beep)
  - Stop: "Pop" (short neutral beep)
  - Complete: "Glass" (pleasant ding)
- Added config commands:
  - `--list-sounds`: Lists all available macOS system sounds
  - `--test-sounds`: Plays all three feedback sounds in sequence
  - `--enable-audio-feedback` / `--disable-audio-feedback`: Toggle audio feedback
- Integrated into `openscribe start` command with graceful fallback if initialization fails
- Audio feedback respects the `audio_feedback` config setting

**Testing**:
- Run `openscribe config --list-sounds` to see available sounds
- Run `openscribe config --test-sounds` to hear the three feedback sounds
- Run `openscribe config --enable-audio-feedback` to enable
- Run `openscribe config --disable-audio-feedback` to disable
- Start command plays sounds at appropriate state transitions

---

## Phase 9: Keyboard Simulation & Auto-Paste ✅

**Goal**: Implement direct text injection at cursor position

**Deliverables**:
- CGEvent-based keyboard simulation (NOT clipboard-based)
- Type text character-by-character at cursor position
- Request and verify Accessibility permissions
- Clear error messages if permissions denied (with System Preferences instructions)
- `--no-paste` flag to disable auto-paste
- Test with various applications (TextEdit, browsers, terminals)

**Why this grouping**: Keyboard simulation is macOS-specific and requires permissions handling. It's complex enough to warrant its own phase.

**Test**: Run transcription, verify text appears at cursor in active application. Test permission denial scenario.

**Implementation Notes**:
- Created `internal/keyboard` package with platform-agnostic interface
- macOS implementation in `keyboard_darwin.go` uses CGEvent APIs for direct text injection
- Uses `CGEventCreateKeyboardEvent` and `CGEventKeyboardSetUnicodeString` for Unicode character typing
- 2ms delay between characters for reliability
- Comprehensive accessibility permissions checking with `AXIsProcessTrustedWithOptions`
- `RequestPermissions()` function prompts user to grant permissions
- Detailed error messages guide users to System Preferences > Security & Privacy > Accessibility
- `--no-paste` flag support integrated into start command
- Graceful fallback if permissions denied (shows text in terminal only)
- Stub implementation for non-Darwin platforms

**Testing**:
- Build succeeds with `make build`
- Start command initializes keyboard simulation when auto-paste enabled
- Permissions check happens on startup with helpful error messages
- `--no-paste` flag properly disables keyboard initialization
- Ready for integration with actual transcription in Phase 10

---

## Phase 10: Integration - The `start` Command

**Goal**: Wire everything together into the main application flow

**Deliverables**:
- Implement complete `openscribe start` command
- Integration of all subsystems:
  1. Check for models (error if missing)
  2. Check permissions (microphone, accessibility)
  3. Load configuration
  4. Initialize hotkey listener
  5. Start event loop (wait for hotkey double-press)
  6. Record audio on first press, stop on second press
  7. Transcribe audio
  8. Play feedback sounds at each stage
  9. Display transcription in terminal
  10. Auto-paste text at cursor (unless `--no-paste`)
  11. Log transcription
  12. Return to listening state
- Support all flags: `--microphone`, `--model`, `--language`, `--no-paste`, `--verbose`
- Terminal UI with clear status indicators

**Why this grouping**: This is the "assembly" phase where we connect all previous phases. It's mostly integration work with minimal new logic.

**Test**: Full end-to-end flow: `openscribe start` → double-press → record → double-press → transcribe → paste → verify in log.

---

## Phase 11: Error Handling & Edge Cases

**Goal**: Robust error handling and user experience polish

**Deliverables**:
- Comprehensive error messages for all failure modes:
  - No models found → suggest `openscribe setup`
  - No microphone available → list available devices
  - Permission denied → show System Preferences navigation
  - Network errors during model download
  - Disk space issues
  - Corrupted models
  - Invalid audio format
  - Transcription failures
- Timeout handling (e.g., max recording time)
- Handle hotkey press during active transcription
- Validate configuration values
- Recovery from partial failures
- Progress indicators for long operations

**Why this grouping**: Error handling is tedious but critical. Better to address it systematically after core functionality works.

**Test**: Deliberately trigger each error condition, verify helpful message displayed.

---

## Phase 12: Polish, Documentation & Testing

**Goal**: Prepare for release

**Deliverables**:
- End-to-end testing of all workflows
- Documentation:
  - README.md with installation and usage instructions
  - TROUBLESHOOTING.md for common issues
  - Permission setup guide
- Performance optimization (if needed)
- Memory usage verification
- Code cleanup and comments
- Version command with build info
- Prepare for Homebrew distribution (formula skeleton)

**Why this grouping**: Final polish after all features work. Documentation benefits from having a complete, working system to describe.

**Test**: Fresh installation testing on a clean macOS system. Verify all documentation is accurate.

---

## Phase 13 (Optional): Homebrew Distribution

**Goal**: Package and distribute via Homebrew

**Deliverables**:
- Create Homebrew tap repository
- Write Homebrew formula
- Binary release automation (GitHub Actions or similar)
- Support for ARM64 and x86_64 architectures
- Post-install instructions
- Caveats about permissions and setup

**Why separate**: Distribution is independent of core functionality. Can be done after v1 is feature-complete.

**Test**: Install via Homebrew on a fresh system, verify it works as expected.

---

## Development Tips for Each Phase

1. **Start each phase with a clear goal**: Ask Claude Code to implement Phase N with its specific deliverables
2. **Test incrementally**: Don't move to the next phase until current phase is working
3. **Create simple test commands**: Add temporary test/debug commands to verify functionality
4. **Use verbose logging**: Add debug output early, clean it up in Phase 11
5. **Commit after each phase**: Keep git history clean with logical checkpoints
6. **Iterate within phases**: It's OK to refine and improve within a phase before moving on
7. **Update this plan**: Add notes about what worked/didn't work for future reference

---

## Dependency Graph

```
Phase 1 (Foundation)
  ↓
Phase 2 (Config) ←──────────────┐
  ↓                              │
Phase 3 (Models) ←──┐            │
  ↓                 │            │
Phase 4 (Audio)     │            │
  ↓                 │            │
Phase 5 (Transcription)          │
  ↓                 │            │
Phase 6 (Logging) ──┘            │
  ↓                              │
Phase 7 (Hotkeys) ───────────────┤
  ↓                              │
Phase 8 (Audio Feedback)         │
  ↓                              │
Phase 9 (Keyboard Sim)           │
  ↓                              │
Phase 10 (Integration) ──────────┘
  ↓
Phase 11 (Error Handling)
  ↓
Phase 12 (Polish)
  ↓
Phase 13 (Distribution)
```

---

## Quick Start Checklist

- [x] Phase 1: Can run `openscribe --help`
- [x] Phase 2: Can run `openscribe config --show`
- [x] Phase 3: Can run `openscribe setup` and download models
- [x] Phase 4: Can list and select microphones
- [x] Phase 5: Can transcribe a test audio file
- [x] Phase 6: Can view logs with `openscribe logs show`
- [x] Phase 7: Can detect hotkey double-press
- [x] Phase 8: Can hear feedback sounds
- [x] Phase 9: Can auto-paste text at cursor
- [ ] Phase 10: Full flow works end-to-end
- [ ] Phase 11: All error cases handled gracefully
- [ ] Phase 12: Documentation complete and tested
- [ ] Phase 13: Available via Homebrew

---

## Next Steps

Start with Phase 1. Ask Claude Code:

> "Let's implement Phase 1 from PLAN.md. Set up the Go project structure, initialize modules, integrate Cobra CLI framework, and create the basic command skeleton for OpenScribe."

Good luck building! 🚀
