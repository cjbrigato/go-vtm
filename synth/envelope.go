package synth

// EnvelopeStage represents the current stage of the envelope
type EnvelopeStage int

const (
	Attack EnvelopeStage = iota
	Decay
	Sustain
	Release
	Off
)

// Envelope implements ADSR envelope generator
type Envelope struct {
	attackTime  float64 // seconds
	decayTime   float64 // seconds
	sustainLevel float64 // 0.0 to 1.0
	releaseTime float64 // seconds

	stage       EnvelopeStage
	level       float64
	sampleRate  float64
	sampleCount int
}

// NewEnvelope creates a new ADSR envelope
func NewEnvelope(attack, decay, sustain, release, sampleRate float64) *Envelope {
	return &Envelope{
		attackTime:   attack,
		decayTime:    decay,
		sustainLevel: sustain,
		releaseTime:  release,
		sampleRate:   sampleRate,
		stage:        Off,
		level:        0.0,
	}
}

// Trigger starts the envelope from attack stage
func (e *Envelope) Trigger() {
	e.stage = Attack
	e.sampleCount = 0
}

// Release moves to release stage
func (e *Envelope) Release() {
	if e.stage != Off {
		e.stage = Release
		e.sampleCount = 0
	}
}

// Next generates the next envelope sample
func (e *Envelope) Next() float64 {
	switch e.stage {
	case Attack:
		attackSamples := int(e.attackTime * e.sampleRate)
		if attackSamples == 0 {
			e.level = 1.0
			e.stage = Decay
			e.sampleCount = 0
		} else if e.sampleCount >= attackSamples {
			e.level = 1.0
			e.stage = Decay
			e.sampleCount = 0
		} else {
			e.level = float64(e.sampleCount) / float64(attackSamples)
			e.sampleCount++
		}

	case Decay:
		decaySamples := int(e.decayTime * e.sampleRate)
		if decaySamples == 0 {
			e.level = e.sustainLevel
			e.stage = Sustain
		} else if e.sampleCount >= decaySamples {
			e.level = e.sustainLevel
			e.stage = Sustain
		} else {
			t := float64(e.sampleCount) / float64(decaySamples)
			e.level = 1.0 + t*(e.sustainLevel-1.0)
			e.sampleCount++
		}

	case Sustain:
		e.level = e.sustainLevel

	case Release:
		releaseSamples := int(e.releaseTime * e.sampleRate)
		startLevel := e.sustainLevel
		if releaseSamples == 0 {
			e.level = 0.0
			e.stage = Off
		} else if e.sampleCount >= releaseSamples {
			e.level = 0.0
			e.stage = Off
		} else {
			t := float64(e.sampleCount) / float64(releaseSamples)
			e.level = startLevel * (1.0 - t)
			e.sampleCount++
		}

	case Off:
		e.level = 0.0
	}

	return e.level
}

// IsActive returns true if envelope is not in Off stage
func (e *Envelope) IsActive() bool {
	return e.stage != Off
}
