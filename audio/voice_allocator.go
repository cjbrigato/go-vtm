package audio

import (
	"github.com/cjbrigato/go-vtm/synth"
	"github.com/cjbrigato/go-vtm/tracker"
)

// VoiceAllocator manages polyphonic voice allocation for a channel
type VoiceAllocator struct {
	voices     []*synth.Voice
	noteMap    map[int]*synth.Voice // Maps note number to voice
	instrument *tracker.Instrument
	sampleRate float64
	maxVoices  int
	activeNotes []int // Track which notes are currently active
}

// NewVoiceAllocator creates a voice allocator with max polyphony
func NewVoiceAllocator(instrument *tracker.Instrument, sampleRate float64, maxVoices int) *VoiceAllocator {
	voices := make([]*synth.Voice, maxVoices)
	
	// Create voices based on instrument type
	for i := range voices {
		if instrument.IsFM {
			// Create FM instrument
			var fmInst *synth.FMInstrument
			switch instrument.FMPreset {
			case "PIANO":
				fmInst = synth.NewPianoFMInstrument(sampleRate)
			case "EPIANO":
				fmInst = synth.NewElectricPianoFMInstrument(sampleRate)
			case "BASS":
				fmInst = synth.NewFMBassFMInstrument(sampleRate)
			case "LEAD":
				fmInst = synth.NewFMLeadFMInstrument(sampleRate)
			case "BRASS":
				fmInst = synth.NewFMBrassFMInstrument(sampleRate)
			case "BELL":
				fmInst = synth.NewFMBellFMInstrument(sampleRate)
			case "ARP":
				fmInst = synth.NewFMArpFMInstrument(sampleRate)
			default:
				fmInst = synth.NewFMLeadFMInstrument(sampleRate)
			}
			voices[i] = synth.NewFMVoice(fmInst)
		} else {
			// Traditional instrument
			voices[i] = synth.NewVoice(instrument.WaveType, sampleRate)
			voices[i].SetInstrument(instrument.WaveType, instrument.Attack, instrument.Decay, instrument.Sustain, instrument.Release)
		}
	}
	
	return &VoiceAllocator{
		voices:      voices,
		noteMap:     make(map[int]*synth.Voice),
		instrument:  instrument,
		sampleRate:  sampleRate,
		maxVoices:   maxVoices,
		activeNotes: make([]int, 0, maxVoices),
	}
}

// NoteOn triggers a note (finds or allocates a voice)
func (va *VoiceAllocator) NoteOn(note int, velocity float64) {
	// If this note is already playing, restart it
	if voice, exists := va.noteMap[note]; exists {
		voice.NoteOn(note, velocity)
		return
	}
	
	// Find an available voice
	var selectedVoice *synth.Voice
	
	// First, try to find an inactive voice
	for _, voice := range va.voices {
		if !voice.IsActive() {
			selectedVoice = voice
			break
		}
	}
	
	// If no inactive voice, steal the oldest one
	if selectedVoice == nil {
		// Steal the first note in activeNotes (oldest)
		if len(va.activeNotes) > 0 {
			oldestNote := va.activeNotes[0]
			selectedVoice = va.noteMap[oldestNote]
			delete(va.noteMap, oldestNote)
			// Remove from activeNotes
			va.activeNotes = va.activeNotes[1:]
		} else {
			// Fallback: use first voice
			selectedVoice = va.voices[0]
		}
	}
	
	// Trigger the voice
	selectedVoice.NoteOn(note, velocity)
	va.noteMap[note] = selectedVoice
	va.activeNotes = append(va.activeNotes, note)
}

// NoteOff releases a specific note
func (va *VoiceAllocator) NoteOff(note int) {
	if voice, exists := va.noteMap[note]; exists {
		voice.NoteOff()
		delete(va.noteMap, note)
		// Remove from activeNotes
		for i, n := range va.activeNotes {
			if n == note {
				va.activeNotes = append(va.activeNotes[:i], va.activeNotes[i+1:]...)
				break
			}
		}
	}
}

// AllNotesOff releases all notes
func (va *VoiceAllocator) AllNotesOff() {
	for _, voice := range va.voices {
		voice.NoteOff()
	}
	va.noteMap = make(map[int]*synth.Voice)
	va.activeNotes = va.activeNotes[:0] // Clear slice
}

// Next generates the next sample by mixing all active voices
func (va *VoiceAllocator) Next() float64 {
	var sample float64
	activeCount := 0
	
	for _, voice := range va.voices {
		if voice.IsActive() {
			sample += voice.Next()
			activeCount++
		}
	}
	
	// Normalize by number of active voices to prevent clipping
	if activeCount > 0 {
		sample /= float64(activeCount)
	}
	
	return sample
}

// IsActive returns true if any voice is active
func (va *VoiceAllocator) IsActive() bool {
	for _, voice := range va.voices {
		if voice.IsActive() {
			return true
		}
	}
	return false
}

// GetVoices returns all voices in this allocator (direct access)
func (va *VoiceAllocator) GetVoices() []*synth.Voice {
	return va.voices
}

// GetVoice returns a specific voice by index (0-3 for default 4-voice polyphony)
func (va *VoiceAllocator) GetVoice(index int) *synth.Voice {
	if index >= 0 && index < len(va.voices) {
		return va.voices[index]
	}
	return nil
}

// GetActiveNotes returns a list of currently active note numbers
func (va *VoiceAllocator) GetActiveNotes() []int {
	// Return a copy to prevent external modification
	notes := make([]int, len(va.activeNotes))
	copy(notes, va.activeNotes)
	return notes
}

// GetActiveVoiceCount returns the number of currently active voices
func (va *VoiceAllocator) GetActiveVoiceCount() int {
	return len(va.activeNotes)
}

// GetVoiceForNote returns the voice playing a specific note (if any)
func (va *VoiceAllocator) GetVoiceForNote(note int) *synth.Voice {
	return va.noteMap[note]
}

// SetVoiceNote directly controls a specific voice (bypass allocation)
// Useful for advanced control over harmonies
func (va *VoiceAllocator) SetVoiceNote(voiceIndex int, note int, velocity float64) {
	if voiceIndex >= 0 && voiceIndex < len(va.voices) {
		va.voices[voiceIndex].NoteOn(note, velocity)
	}
}

// ReleaseVoice directly releases a specific voice by index
func (va *VoiceAllocator) ReleaseVoice(voiceIndex int) {
	if voiceIndex >= 0 && voiceIndex < len(va.voices) {
		va.voices[voiceIndex].NoteOff()
	}
}

// GetMaxVoices returns the maximum number of voices available
func (va *VoiceAllocator) GetMaxVoices() int {
	return va.maxVoices
}

// PlayChord triggers multiple notes simultaneously (convenience method)
func (va *VoiceAllocator) PlayChord(notes []int, velocity float64) {
	for _, note := range notes {
		va.NoteOn(note, velocity)
	}
}

// ReleaseChord releases multiple notes simultaneously
func (va *VoiceAllocator) ReleaseChord(notes []int) {
	for _, note := range notes {
		va.NoteOff(note)
	}
}

