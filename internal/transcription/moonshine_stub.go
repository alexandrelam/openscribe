//go:build !moonshine

package transcription

import (
	"fmt"

	"github.com/alexandrelam/openscribe/internal/config"
)

func newMoonshineTranscriber(_ *config.Config) (Transcriber, error) {
	return nil, fmt.Errorf("moonshine backend not available: rebuild with -tags moonshine")
}
