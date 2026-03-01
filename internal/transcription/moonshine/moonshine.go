//go:build moonshine

// Package moonshine provides speech-to-text transcription using Moonshine via cgo.
//
// This package requires libmoonshine.a and moonshine-c-api.h extracted from the
// Moonshine XCFramework (see Makefile target: moonshine-deps).
//
// Build with: go build -tags moonshine
package moonshine

/*
#cgo CFLAGS: -I${SRCDIR}/../../../third_party/moonshine/include
#cgo LDFLAGS: -L${SRCDIR}/../../../third_party/moonshine/lib -lmoonshine -lstdc++ -framework Accelerate -framework CoreML
// moonshine-c-api.h is extracted from the Moonshine XCFramework by `make moonshine-deps`
#include "moonshine-c-api.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

// ModelArch maps model size names to the C constants
var ModelArch = map[string]C.uint32_t{
	"tiny":            C.MOONSHINE_MODEL_ARCH_TINY,
	"base":            C.MOONSHINE_MODEL_ARCH_BASE,
	"small-streaming":  C.MOONSHINE_MODEL_ARCH_SMALL_STREAMING,
	"medium-streaming": C.MOONSHINE_MODEL_ARCH_MEDIUM_STREAMING,
}

// Transcriber wraps the Moonshine C library for speech-to-text.
type Transcriber struct {
	handle C.int32_t
}

// New creates a Moonshine transcriber with the given model directory and architecture.
// modelDir should contain encoder_model.ort, decoder_model_merged.ort, and tokenizer.bin.
// arch is the model architecture name (e.g. "tiny", "base").
func New(modelDir, arch string) (*Transcriber, error) {
	cModelDir := C.CString(modelDir)
	defer C.free(unsafe.Pointer(cModelDir))

	modelArch, ok := ModelArch[arch]
	if !ok {
		return nil, fmt.Errorf("unknown moonshine model arch: %s", arch)
	}

	handle := C.moonshine_load_transcriber_from_files(
		cModelDir,
		modelArch,
		nil, // no options
		0,   // options count
		C.MOONSHINE_HEADER_VERSION,
	)
	if handle < 0 {
		errStr := C.GoString(C.moonshine_error_to_string(handle))
		return nil, fmt.Errorf("failed to load moonshine model from %s: %s", modelDir, errStr)
	}

	return &Transcriber{handle: handle}, nil
}

// Transcribe takes PCM audio samples (float32, mono) and sample rate, returns transcribed text.
func (t *Transcriber) Transcribe(samples []float32, sampleRate int32) (string, error) {
	if len(samples) == 0 {
		return "", fmt.Errorf("empty audio samples")
	}

	var transcript *C.struct_transcript_t

	errCode := C.moonshine_transcribe_without_streaming(
		t.handle,
		(*C.float)(&samples[0]),
		C.uint64_t(len(samples)),
		C.int32_t(sampleRate),
		0, // flags
		&transcript,
	)
	if errCode != C.MOONSHINE_ERROR_NONE {
		errStr := C.GoString(C.moonshine_error_to_string(errCode))
		return "", fmt.Errorf("moonshine transcription failed: %s", errStr)
	}

	if transcript == nil || transcript.line_count == 0 {
		return "", nil
	}

	// Collect all line texts
	var parts []string
	lines := unsafe.Slice(transcript.lines, transcript.line_count)
	for _, line := range lines {
		text := C.GoString(line.text)
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " "), nil
}

// Close releases the Moonshine model resources.
func (t *Transcriber) Close() {
	C.moonshine_free_transcriber(t.handle)
}
