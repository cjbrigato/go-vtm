package audio

import (
	"github.com/cjbrigato/go-vtm/synth"
)

// Player plays tracker modules
type Player struct {
	module         *TrackerModule
	voices         []*synth.Voice
	sampleRate     float64
	currentPos     int // Position in sequence
	currentRow     int // Current row in pattern
	sampleCounter  int // Sample counter for row timing
	samplesPerRow  int // Samples before advancing to next row
	done           bool
}

// NewPlayer creates a new tracker player
func NewPlayer(module *TrackerModule, sampleRate float64) *Player {
	// Calculate samples per row based on tempo
	// tempo = beats per minute
	// 1 beat = 4 rows (typically)
	// samplesPerRow = (60 / tempo / 4) * sampleRate
	samplesPerSecond := sampleRate
	secondsPerBeat := 60.0 / float64(module.Tempo)
	secondsPerRow := secondsPerBeat / 4.0 // 4 rows per beat
	samplesPerRow := int(secondsPerRow * samplesPerSecond)

	// Create voices (max 8 channels)
	numChannels := 8
	voices := make([]*synth.Voice, numChannels)
	for i := range voices {
		voices[i] = synth.NewVoice(synth.Square, sampleRate)
	}

	// Configure instruments
	for i, inst := range module.Instruments {
		if i < len(voices) {
			voices[i].SetInstrument(inst.WaveType, inst.Attack, inst.Decay, inst.Sustain, inst.Release)
		}
	}

	return &Player{
		module:        module,
		voices:        voices,
		sampleRate:    sampleRate,
		currentPos:    0,
		currentRow:    0,
		sampleCounter: 0,
		samplesPerRow: samplesPerRow,
		done:          false,
	}
}

// Next generates the next audio sample (mono)
func (p *Player) Next() float64 {
	if p.done {
		return 0.0
	}

	// Check if we need to process a new row
	if p.sampleCounter == 0 {
		p.processRow()
	}

	// Generate audio by mixing all voices
	var sample float64
	for _, voice := range p.voices {
		sample += voice.Next()
	}

	// Simple mixing (average)
	sample /= float64(len(p.voices))

	// Advance sample counter
	p.sampleCounter++
	if p.sampleCounter >= p.samplesPerRow {
		p.sampleCounter = 0
		p.currentRow++

		// Check if we've finished the current pattern
		if p.currentPos < len(p.module.Sequence) {
			patternIdx := p.module.Sequence[p.currentPos]
			if patternIdx < len(p.module.Patterns) {
				if p.currentRow >= p.module.Patterns[patternIdx].Rows {
					p.currentRow = 0
					p.currentPos++

					// Check if we've finished the entire sequence
					if p.currentPos >= len(p.module.Sequence) {
						p.done = true
					}
				}
			}
		}
	}

	return sample
}

// processRow processes the current row and triggers notes
func (p *Player) processRow() {
	if p.currentPos >= len(p.module.Sequence) {
		return
	}

	patternIdx := p.module.Sequence[p.currentPos]
	if patternIdx >= len(p.module.Patterns) {
		return
	}

	pattern := p.module.Patterns[patternIdx]

	// Process each channel
	for ch := 0; ch < len(pattern.Channels) && ch < len(p.voices); ch++ {
		if p.currentRow < len(pattern.Channels[ch]) {
			note := pattern.Channels[ch][p.currentRow]

			if note.Note >= 0 {
				// Trigger note
				p.voices[ch].NoteOn(note.Note, note.Volume)
			} else if note.Note == -2 {
				// Note off command (not used in current parser, but could be)
				p.voices[ch].NoteOff()
			}
			// -1 is rest, do nothing
		}
	}
}

// IsDone returns true if playback is complete
func (p *Player) IsDone() bool {
	return p.done
}

// Reset resets the player to the beginning
func (p *Player) Reset() {
	p.currentPos = 0
	p.currentRow = 0
	p.sampleCounter = 0
	p.done = false
}

// Stream implements a simple audio stream interface
func (p *Player) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if p.done {
			return i, false
		}

		sample := p.Next()
		samples[i][0] = sample // Left channel
		samples[i][1] = sample // Right channel (mono for now)
	}
	return len(samples), true
}
