package synth

import "math"

// FMOperator represents an FM synthesis operator
// An operator is an oscillator with its own envelope that can modulate other operators
type FMOperator struct {
	phase      float64
	frequency  float64
	sampleRate float64
	envelope   *Envelope
	outputLevel float64 // Operator output volume (0.0 to 1.0)
	ratio      float64  // Frequency ratio (e.g., 1.0 = fundamental, 2.0 = octave up)
}

// NewFMOperator creates a new FM operator
func NewFMOperator(sampleRate float64) *FMOperator {
	return &FMOperator{
		phase:       0.0,
		sampleRate:  sampleRate,
		envelope:    NewEnvelope(0.01, 0.1, 0.6, 0.2, sampleRate),
		outputLevel: 1.0,
		ratio:       1.0,
	}
}

// SetEnvelope configures the operator's envelope
func (op *FMOperator) SetEnvelope(attack, decay, sustain, release float64) {
	op.envelope = NewEnvelope(attack, decay, sustain, release, op.sampleRate)
}

// SetRatio sets the frequency ratio relative to the fundamental
func (op *FMOperator) SetRatio(ratio float64) {
	op.ratio = ratio
}

// SetOutputLevel sets the operator's output level
func (op *FMOperator) SetOutputLevel(level float64) {
	op.outputLevel = level
}

// SetFrequency sets the base frequency
func (op *FMOperator) SetFrequency(freq float64) {
	op.frequency = freq * op.ratio
}

// Trigger starts the operator's envelope
func (op *FMOperator) Trigger() {
	op.envelope.Trigger()
	op.phase = 0.0
}

// Release releases the operator's envelope
func (op *FMOperator) Release() {
	op.envelope.Release()
}

// Next generates the next sample with optional frequency modulation
// modulation is added to the phase (in radians)
func (op *FMOperator) Next(modulation float64) float64 {
	// Get envelope value
	env := op.envelope.Next()
	
	// Generate sine wave with phase modulation
	sample := math.Sin(2.0*math.Pi*op.phase + modulation)
	
	// Advance phase
	phaseIncrement := op.frequency / op.sampleRate
	op.phase += phaseIncrement
	if op.phase >= 1.0 {
		op.phase -= math.Floor(op.phase)
	}
	
	return sample * env * op.outputLevel
}

// IsActive returns true if the operator's envelope is active
func (op *FMOperator) IsActive() bool {
	return op.envelope.IsActive()
}

// Reset resets the operator
func (op *FMOperator) Reset() {
	op.phase = 0.0
	op.envelope.stage = Off
	op.envelope.level = 0.0
}

