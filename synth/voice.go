package synth

// Voice represents a single synthesis voice (oscillator + envelope)
type Voice struct {
	oscillator *Oscillator
	envelope   *Envelope
	volume     float64
	active     bool
}

// NewVoice creates a new synthesis voice
func NewVoice(waveType WaveType, sampleRate float64) *Voice {
	return &Voice{
		oscillator: NewOscillator(waveType, sampleRate),
		envelope:   NewEnvelope(0.01, 0.1, 0.6, 0.2, sampleRate),
		volume:     1.0,
		active:     false,
	}
}

// SetInstrument configures the voice with instrument parameters
func (v *Voice) SetInstrument(waveType WaveType, attack, decay, sustain, release float64) {
	v.oscillator.waveType = waveType
	v.envelope.attackTime = attack
	v.envelope.decayTime = decay
	v.envelope.sustainLevel = sustain
	v.envelope.releaseTime = release
}

// NoteOn triggers a note
func (v *Voice) NoteOn(note int, volume float64) {
	freq := NoteToFrequency(note)
	v.oscillator.SetFrequency(freq)
	v.volume = volume
	v.envelope.Trigger()
	v.active = true
}

// NoteOff releases the note
func (v *Voice) NoteOff() {
	v.envelope.Release()
}

// Next generates the next audio sample
func (v *Voice) Next() float64 {
	if !v.active && !v.envelope.IsActive() {
		return 0.0
	}

	osc := v.oscillator.Next()
	env := v.envelope.Next()

	if !v.envelope.IsActive() {
		v.active = false
	}

	return osc * env * v.volume
}

// IsActive returns true if the voice is producing sound
func (v *Voice) IsActive() bool {
	return v.active || v.envelope.IsActive()
}
