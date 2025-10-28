//go:build (linux || windows || darwin) && !noaudio

package audio

import (
	"io"
	"unsafe"

	"github.com/ebitengine/oto/v3"
	"github.com/cjbrigato/go-vtm/tracker"
)

// AudioPlayback handles real-time audio output
type AudioPlayback struct {
	otoContext  *oto.Context
	player      *Player
	sampleRate  int
	done        chan bool
	audioPlayer *oto.Player
}

// NewAudioPlayback creates a new audio playback system
func NewAudioPlayback(module *tracker.TrackerModule, sampleRate int) (*AudioPlayback, error) {
	// Initialize oto context
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		Format:       oto.FormatFloat32LE,
	}

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, err
	}

	// Wait for the audio system to be ready
	<-readyChan

	// Create player
	player := NewPlayer(module, float64(sampleRate))

	return &AudioPlayback{
		otoContext: otoCtx,
		player:     player,
		sampleRate: sampleRate,
		done:       make(chan bool),
	}, nil
}

// Play starts audio playback in a goroutine
func (ap *AudioPlayback) Play() error {
	ap.audioPlayer = ap.otoContext.NewPlayer(&audioReader{
		player:     ap.player,
		sampleRate: ap.sampleRate,
		done:       ap.done,
	})

	ap.audioPlayer.Play()

	return nil
}

// Stop stops audio playback
func (ap *AudioPlayback) Stop() {
	close(ap.done)
}

func (ap *AudioPlayback) IsPlaying() bool {
	return ap.audioPlayer.IsPlaying()
}

// IsDone returns true if the music has finished
func (ap *AudioPlayback) IsDone() bool {
	return ap.player.IsDone()
}

// audioReader implements io.Reader for audio streaming
type audioReader struct {
	player     *Player
	sampleRate int
	done       chan bool
	buffer     []float32
}

// Read fills the buffer with audio samples
func (ar *audioReader) Read(p []byte) (n int, err error) {
	// Check if playback should stop
	select {
	case <-ar.done:
		return 0, io.EOF
	default:
	}

	if ar.player.IsDone() {
		return 0, io.EOF
	}

	// Convert byte slice to float32 samples
	// Each float32 is 4 bytes, and we have 2 channels (stereo)
	numSamples := len(p) / 4 / 2

	// Generate samples
	for i := 0; i < numSamples; i++ {
		sample := float32(ar.player.Next())

		// Convert to bytes (little-endian float32)
		sampleBytes := float32ToBytes(sample)

		// Write stereo (same sample for both channels)
		offset := i * 8                         // 4 bytes per float32 * 2 channels
		copy(p[offset:offset+4], sampleBytes)   // Left channel
		copy(p[offset+4:offset+8], sampleBytes) // Right channel
	}

	return numSamples * 8, nil
}

// float32ToBytes converts a float32 to little-endian bytes
func float32ToBytes(f float32) []byte {
	bits := *(*uint32)(unsafe.Pointer(&f))
	return []byte{
		byte(bits),
		byte(bits >> 8),
		byte(bits >> 16),
		byte(bits >> 24),
	}
}
