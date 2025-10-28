package audio

import (
	"github.com/cjbrigato/go-vtm/synth"
	"github.com/cjbrigato/go-vtm/tracker"
)

// Player plays tracker modules
type Player struct {
	module          *tracker.TrackerModule
	VoiceAllocators []*VoiceAllocator // One allocator per channel for polyphony - PUBLIC for direct access
	sampleRate      float64
	currentPos      int // Position in sequence
	currentRow      int // Current row in pattern
	sampleCounter   int // Sample counter for row timing
	samplesPerRow   int // Samples before advancing to next row
	done            bool
	maxPolyphony    int // Max simultaneous notes per channel
}

// NewPlayer creates a new tracker player
func NewPlayer(module *tracker.TrackerModule, sampleRate float64) *Player {
	// Calculate samples per row based on tempo
	// tempo = beats per minute
	// 1 beat = 4 rows (typically)
	// samplesPerRow = (60 / tempo / 4) * sampleRate
	samplesPerSecond := sampleRate
	secondsPerBeat := 60.0 / float64(module.Tempo)
	secondsPerRow := secondsPerBeat / 4.0 // 4 rows per beat
	samplesPerRow := int(secondsPerRow * samplesPerSecond)

	// Create voice allocators (max 8 channels, 4 voices per channel for polyphony)
	numChannels := 8
	maxPolyphony := 4 // Allow up to 4 simultaneous notes per channel

	voiceAllocators := make([]*VoiceAllocator, numChannels)

	// Create default instrument if none provided
	defaultInst := tracker.Instrument{
		Name:     "Default",
		WaveType: synth.Square,
		Attack:   0.01,
		Decay:    0.1,
		Sustain:  0.6,
		Release:  0.2,
		IsFM:     false,
	}

	// Configure voice allocators with instruments
	for i := range voiceAllocators {
		var inst *tracker.Instrument
		if i < len(module.Instruments) {
			inst = &module.Instruments[i]
		} else {
			inst = &defaultInst
		}
		voiceAllocators[i] = NewVoiceAllocator(inst, sampleRate, maxPolyphony)
	}

	return &Player{
		module:          module,
		VoiceAllocators: voiceAllocators,
		sampleRate:      sampleRate,
		currentPos:      0,
		currentRow:      0,
		sampleCounter:   0,
		samplesPerRow:   samplesPerRow,
		done:            false,
		maxPolyphony:    maxPolyphony,
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

	// Generate audio by mixing all channels
	var sample float64
	for _, allocator := range p.VoiceAllocators {
		sample += allocator.Next()
	}

	// Simple mixing (average)
	sample /= float64(len(p.VoiceAllocators))

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
	for ch := 0; ch < len(pattern.Channels) && ch < len(p.VoiceAllocators); ch++ {
		if p.currentRow < len(pattern.Channels[ch]) {
			note := pattern.Channels[ch][p.currentRow]

			if note.Note >= 0 {
				// Trigger note with proper velocity
				velocity := note.Volume
				if velocity == 0 {
					velocity = 1.0 // Default velocity if not specified
				}

				// Trigger main note
				p.VoiceAllocators[ch].NoteOn(note.Note, velocity)

				// If this is a chord, trigger additional notes
				if note.Chord != nil && len(note.Chord) > 0 {
					for _, chordNote := range note.Chord {
						if chordNote >= 0 {
							p.VoiceAllocators[ch].NoteOn(chordNote, velocity)
						}
					}
				}
			} else if note.Note == -2 {
				// Note off command - releases all notes on this channel
				p.VoiceAllocators[ch].AllNotesOff()
			}
			// -1 is rest, do nothing (notes continue to sustain)
		}
	}
}

// GetChannelVoices returns the voice allocator for a specific channel
// Convenience method for accessing channel harmonies
func (p *Player) GetChannelVoices(channel int) *VoiceAllocator {
	if channel >= 0 && channel < len(p.VoiceAllocators) {
		return p.VoiceAllocators[channel]
	}
	return nil
}

// GetChannelCount returns the number of channels
func (p *Player) GetChannelCount() int {
	return len(p.VoiceAllocators)
}

// GetMaxPolyphony returns the max voices per channel
func (p *Player) GetMaxPolyphony() int {
	return p.maxPolyphony
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
