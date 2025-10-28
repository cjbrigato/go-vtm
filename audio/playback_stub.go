//go:build !(linux || windows || darwin) || noaudio

package audio

import "fmt"

// AudioPlayback stub for unsupported platforms
type AudioPlayback struct{}

// NewAudioPlayback returns an error on unsupported platforms
func NewAudioPlayback(module *TrackerModule, sampleRate int) (*AudioPlayback, error) {
	return nil, fmt.Errorf("audio playback not supported on this platform")
}

// Play is a no-op
func (ap *AudioPlayback) Play() error {
	return nil
}

// Stop is a no-op
func (ap *AudioPlayback) Stop() {}

// IsDone returns true
func (ap *AudioPlayback) IsDone() bool {
	return true
}
