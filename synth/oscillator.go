package synth

import "math"

// WaveType represents different oscillator waveforms
type WaveType int

const (
	Square WaveType = iota
	Saw
	Triangle
	Sine
	Noise
)

// Oscillator generates audio waveforms
type Oscillator struct {
	waveType  WaveType
	frequency float64
	phase     float64
	sampleRate float64
}

// NewOscillator creates a new oscillator
func NewOscillator(waveType WaveType, sampleRate float64) *Oscillator {
	return &Oscillator{
		waveType:   waveType,
		sampleRate: sampleRate,
		phase:      0.0,
	}
}

// SetFrequency sets the oscillator frequency
func (o *Oscillator) SetFrequency(freq float64) {
	o.frequency = freq
}

// Next generates the next sample
func (o *Oscillator) Next() float64 {
	var sample float64

	switch o.waveType {
	case Square:
		if o.phase < 0.5 {
			sample = 1.0
		} else {
			sample = -1.0
		}
	case Saw:
		sample = 2.0*o.phase - 1.0
	case Triangle:
		if o.phase < 0.5 {
			sample = 4.0*o.phase - 1.0
		} else {
			sample = 3.0 - 4.0*o.phase
		}
	case Sine:
		sample = math.Sin(2.0 * math.Pi * o.phase)
	case Noise:
		// Simple white noise
		sample = 2.0*math.Mod(o.phase*12345.6789, 1.0) - 1.0
	}

	// Advance phase
	phaseIncrement := o.frequency / o.sampleRate
	o.phase += phaseIncrement
	if o.phase >= 1.0 {
		o.phase -= 1.0
	}

	return sample
}

// Reset resets the oscillator phase
func (o *Oscillator) Reset() {
	o.phase = 0.0
}
