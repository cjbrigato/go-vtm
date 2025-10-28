package synth

// Voice represents a single synthesis voice (oscillator + envelope OR FM instrument)
type Voice struct {
	// Traditional synthesis
	oscillator *Oscillator
	envelope   *Envelope
	
	// FM synthesis
	fmInstrument *FMInstrument
	useFM        bool
	
	volume     float64
	active     bool
	sampleRate float64
}

// NewVoice creates a new synthesis voice with traditional oscillator
func NewVoice(waveType WaveType, sampleRate float64) *Voice {
	return &Voice{
		oscillator: NewOscillator(waveType, sampleRate),
		envelope:   NewEnvelope(0.01, 0.1, 0.6, 0.2, sampleRate),
		volume:     1.0,
		active:     false,
		useFM:      false,
		sampleRate: sampleRate,
	}
}

// NewFMVoice creates a new voice using FM synthesis
func NewFMVoice(fmInstrument *FMInstrument) *Voice {
	return &Voice{
		fmInstrument: fmInstrument,
		useFM:        true,
		volume:       1.0,
		active:       false,
		sampleRate:   fmInstrument.sampleRate,
	}
}

// NewPianoVoice creates a voice with piano FM preset
func NewPianoVoice(sampleRate float64) *Voice {
	return NewFMVoice(NewPianoFMInstrument(sampleRate))
}

// NewElectricPianoVoice creates a voice with electric piano FM preset
func NewElectricPianoVoice(sampleRate float64) *Voice {
	return NewFMVoice(NewElectricPianoFMInstrument(sampleRate))
}

// SetInstrument configures the voice with instrument parameters (traditional synthesis only)
func (v *Voice) SetInstrument(waveType WaveType, attack, decay, sustain, release float64) {
	if v.useFM {
		return // Can't change FM instrument parameters this way
	}
	v.oscillator.waveType = waveType
	v.envelope.attackTime = attack
	v.envelope.decayTime = decay
	v.envelope.sustainLevel = sustain
	v.envelope.releaseTime = release
}

// SetFMInstrument switches the voice to use an FM instrument
func (v *Voice) SetFMInstrument(fmInstrument *FMInstrument) {
	v.fmInstrument = fmInstrument
	v.useFM = true
}

// NoteOn triggers a note
func (v *Voice) NoteOn(note int, volume float64) {
	v.volume = volume
	v.active = true
	
	if v.useFM && v.fmInstrument != nil {
		v.fmInstrument.NoteOn(note, volume)
	} else {
		freq := NoteToFrequency(note)
		v.oscillator.SetFrequency(freq)
		v.envelope.Trigger()
	}
}

// NoteOff releases the note
func (v *Voice) NoteOff() {
	if v.useFM && v.fmInstrument != nil {
		v.fmInstrument.NoteOff()
	} else {
		v.envelope.Release()
	}
}

// Next generates the next audio sample
func (v *Voice) Next() float64 {
	if v.useFM && v.fmInstrument != nil {
		// FM synthesis path
		if !v.active && !v.fmInstrument.IsActive() {
			return 0.0
		}
		
		sample := v.fmInstrument.Next()
		
		if !v.fmInstrument.IsActive() {
			v.active = false
		}
		
		return sample
	} else {
		// Traditional synthesis path
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
}

// IsActive returns true if the voice is producing sound
func (v *Voice) IsActive() bool {
	if v.useFM && v.fmInstrument != nil {
		return v.active || v.fmInstrument.IsActive()
	}
	return v.active || v.envelope.IsActive()
}
