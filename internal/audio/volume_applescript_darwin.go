//go:build darwin
// +build darwin

package audio

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// appleScriptVolume remembers the system output state captured before muting
// through AppleScript, so it can be put back exactly.
//
// This is the last-resort path for output devices that expose no writable
// CoreAudio mute or volume property. `set volume` is a Standard Additions
// command handled in-process by osascript, so it needs no automation
// permission and prompts the user for nothing.
type appleScriptVolume struct {
	// usedMute is true when the output was silenced via the mute switch
	// rather than by dropping the volume to zero
	usedMute bool

	// prevMuted is the mute state to restore when usedMute is true
	prevMuted bool

	// prevVolume is the 0-100 output volume to restore when usedMute is false
	prevVolume int
}

var (
	appleMutedPattern  = regexp.MustCompile(`output muted:(true|false)`)
	appleVolumePattern = regexp.MustCompile(`output volume:(\d+)`)
)

// runOsascript executes a single AppleScript statement and returns its output
func runOsascript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("osascript failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("osascript failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// muteViaAppleScript silences system output and returns the state to restore.
// It prefers the mute switch and falls back to zeroing the output volume,
// since `output muted` reads as "missing value" on some devices.
func muteViaAppleScript() (*appleScriptVolume, error) {
	settings, err := runOsascript("get volume settings")
	if err != nil {
		return nil, err
	}

	if match := appleMutedPattern.FindStringSubmatch(settings); match != nil {
		state := &appleScriptVolume{usedMute: true, prevMuted: match[1] == "true"}
		if _, err := runOsascript("set volume with output muted"); err != nil {
			return nil, err
		}
		return state, nil
	}

	match := appleVolumePattern.FindStringSubmatch(settings)
	if match == nil {
		return nil, fmt.Errorf("could not read system volume settings: %q", settings)
	}

	volume, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse system volume %q: %w", match[1], err)
	}

	state := &appleScriptVolume{prevVolume: volume}
	if _, err := runOsascript("set volume output volume 0"); err != nil {
		return nil, err
	}
	return state, nil
}

// restore puts the system output back to the captured state
func (a *appleScriptVolume) restore() error {
	if a.usedMute {
		script := "set volume without output muted"
		if a.prevMuted {
			script = "set volume with output muted"
		}
		_, err := runOsascript(script)
		return err
	}

	_, err := runOsascript(fmt.Sprintf("set volume output volume %d", a.prevVolume))
	return err
}
